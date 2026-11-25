package monitor

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	v1 "ad-monitor/api/monitor/v1"
	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/adapter"
	// 防御性 blank import: 即使没有 packed 也保证适配器自注册
	_ "ad-monitor/internal/logic/adapter/weibo"
	_ "ad-monitor/internal/logic/adapter/weibodirect"
	_ "ad-monitor/internal/logic/adapter/xiaohongshu"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/logic/dedup"
	"ad-monitor/internal/logic/queue"
	"ad-monitor/internal/logic/stats"
	"ad-monitor/internal/model/bo"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct {
	dedup *dedup.Dedup
}

func New(d *dedup.Dedup) *Controller {
	return &Controller{dedup: d}
}

// Receive 广告平台监测链接回调入口
// 无论成功失败都返回 204 (避免广告平台判定监测链接故障, 返回 4xx 可能导致链接被禁用)
func (c *Controller) Receive(ctx context.Context, req *v1.MonitorReq) (res *v1.MonitorRes, err error) {
	r := g.RequestFromCtx(ctx)
	// defer 写回 204
	defer r.Response.WriteStatus(http.StatusNoContent)

	accountID := req.AccountID

	// 1. QPS 计数 (fire-and-forget, 常驻 worker 消费)
	stats.RecordClickAsync()

	// 2. 加载账户
	acct, err := cache.GetAccount(ctx, accountID)
	if err != nil || acct == nil || acct.Status != 1 {
		stats.IncrCounter(ctx, "drop", accountID)
		g.Log().Warningf(ctx, "account %d not found or disabled: %v", accountID, err)
		return nil, nil
	}
	stats.RecordRequest(accountID, r.URL.RawQuery)

	// 3. 加载平台 + 适配器
	platform, err := cache.GetPlatform(ctx, acct.PlatformId)
	if err != nil {
		stats.IncrCounter(ctx, "drop", accountID)
		g.Log().Warningf(ctx, "get platform %d failed: %v", acct.PlatformId, err)
		return nil, nil
	}
	ad := adapter.Get(platform.Name)
	if ad == nil {
		stats.IncrCounter(ctx, "drop", accountID)
		g.Log().Warningf(ctx, "adapter not found: %s", platform.Name)
		return nil, nil
	}

	// 4. 解析请求 (必填缺失返回 error, 当作 drop 处理但仍返回 204)
	data, parseErr := adapter.ParseClickRequest(r, ad.Spec())
	if parseErr != nil {
		stats.IncrCounter(ctx, "drop", accountID)
		g.Log().Warningf(ctx, "parse click request failed: %v, raw=%s", parseErr, r.URL.RawQuery)
		return nil, nil
	}
	data.AccountID = accountID

	// 5. 点击计数
	stats.IncrCounter(ctx, "click", accountID)

	// 6. 设备号去重 + CSV (监测维度: {平台名}_{设备号类型}.csv)
	devices := adapter.ExtractDeviceIDs(data)

	// 6.1 透传非设备号字段, 经淘联 callbackUrl 的 d 参数带回回传 worker:
	//   advertiser_id (小红书回传), ad_id/creative_id/os/androidid (上报淘联 adid/cid/os/androididMd5),
	//   ts = 我方收到点击的时间(毫秒, 微博直达订单 click_time)
	// 必须在启动 goroutine 之前写入, 避免对 devices map 的并发读写竞争
	// 这些字段不在 DeviceFields 中, 不参与去重
	for _, f := range passThroughFields {
		if v := data.Values[f]; v != "" {
			devices[f] = v
		}
	}
	now := time.Now()
	devices[consts.FieldTS] = strconv.FormatInt(now.UnixMilli(), 10)

	go c.dedup.ProcessClickDevices(context.Background(), platform.Name, devices)

	// 7. 联调 / 正常 / 虚假 三种模式分支
	clickID := data.Values[consts.FieldClickID]
	ts := now.Unix()

	if acct.DebugMode {
		// === 联调模式 ===
		// 跳过淘联上报, 直接用本次点击模拟转化回传给广告平台
		// 遍历 debug_events (或全部已启用映射), 每个事件入 callback_report 队列
		g.Log().Infof(ctx, "monitor: debug mode, skip taolian, simulate callback for account=%d click=%s", accountID, clickID)

		var debugEvents []string
		_ = json.Unmarshal([]byte(acct.DebugEvents), &debugEvents)

		mappings, _ := cache.GetMappings(ctx, accountID)
		for _, m := range mappings {
			if !m.Enabled {
				continue
			}
			// 如果配置了 debug_events, 只回传指定的广告平台事件
			if len(debugEvents) > 0 {
				found := false
				for _, e := range debugEvents {
					if e == m.PlatformType {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			// 立即入回传队列 (不延迟)
			_ = queue.EnqueueCallbackReport(ctx, &bo.CallbackTaskPayload{
				AccountID:   accountID,
				ClickID:     clickID,
				TaolianType: m.TaolianType,
				MappedType:  m.PlatformType,
				Devices:     devices,
				TS:          now.UnixMilli(),
				Debug:       true,
			})
			g.Log().Debugf(ctx, "monitor: debug enqueue callback event=%s for account=%d", m.PlatformType, accountID)
		}

	} else {
		// === 正常模式 ===
		// 上报淘联
		if err := queue.EnqueueTaolianReport(ctx, &bo.TaolianTaskPayload{
			AccountID: accountID,
			ClickID:   clickID,
			Devices:   devices,
			TS:        ts,
		}); err != nil {
			g.Log().Warningf(ctx, "enqueue taolian failed: %v", err)
		}

		// 虚假上报判定
		if acct.FakeEnabled && hitRate(acct.FakeRate) {
			taolianType, mappedType := pickFakeType(ctx, acct)
			if mappedType != "" {
				delayMin, delayMax := acct.FakeDelayMin, acct.FakeDelayMax
				if delayMax <= delayMin {
					delayMax = delayMin + 1
				}
				delay := time.Duration(rand.Intn(delayMax-delayMin)+delayMin) * time.Second

				_ = queue.EnqueueFakeReport(ctx, &bo.FakeTaskPayload{
					AccountID:   accountID,
					ClickID:     clickID,
					Devices:     devices,
					TaolianType: taolianType,
					MappedType:  mappedType,
					TS:          now.Add(delay).UnixMilli(),
				}, delay)
			}
		}
	}

	return nil, nil
}

var passThroughFields = []consts.InternalField{
	consts.FieldAdvertiserID, consts.FieldAdID, consts.FieldCreativeID, consts.FieldOS, consts.FieldAndroidID,
}

// hitRate 按概率 (1-100) 决定是否命中
func hitRate(rate int) bool {
	if rate <= 0 {
		return false
	}
	if rate >= 100 {
		return true
	}
	return rand.Intn(100)+1 <= rate
}

// pickFakeType 从账户已启用的映射中选一个作为虚假回传类型
// 返回 (taolianType, mappedType)
// 优先级: acct.FakeTypes (已配置的类型列表) > 账户所有已启用映射随机
func pickFakeType(ctx context.Context, acct *entity.AdAccount) (int, string) {
	mappings, err := cache.GetMappings(ctx, acct.Id)
	if err != nil || len(mappings) == 0 {
		return 0, ""
	}
	// 过滤启用的
	enabled := make([]*entity.ActionMapping, 0, len(mappings))
	for _, m := range mappings {
		if m.Enabled {
			enabled = append(enabled, m)
		}
	}
	if len(enabled) == 0 {
		return 0, ""
	}

	// 如果配置了 FakeTypes, 过滤只剩这些类型
	var fakeTypes []int
	_ = json.Unmarshal([]byte(acct.FakeTypes), &fakeTypes)
	if len(fakeTypes) > 0 {
		filtered := make([]*entity.ActionMapping, 0)
		for _, m := range enabled {
			for _, t := range fakeTypes {
				if m.TaolianType == t {
					filtered = append(filtered, m)
					break
				}
			}
		}
		if len(filtered) > 0 {
			enabled = filtered
		}
	}

	pick := enabled[rand.Intn(len(enabled))]
	return pick.TaolianType, pick.PlatformType
}
