package callback

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	v1 "ad-monitor/api/callback/v1"
	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/logic/dedup"
	"ad-monitor/internal/logic/queue"
	"ad-monitor/internal/logic/stats"
	"ad-monitor/internal/model/bo"

	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct {
	dedup *dedup.Dedup
}

func New(d *dedup.Dedup) *Controller {
	return &Controller{dedup: d}
}

// Receive 淘联回传入口
func (c *Controller) Receive(ctx context.Context, req *v1.CallbackReq) (res *v1.CallbackRes, err error) {
	r := g.RequestFromCtx(ctx)
	defer func() {
		r.Response.WriteStatus(http.StatusOK)
		r.Response.WriteString("success")
	}()

	accountID := req.A
	stats.IncrCounter(ctx, "cb:recv", accountID)

	// 1. 加载账户
	acct, err := cache.GetAccount(ctx, accountID)
	if err != nil || acct == nil || acct.Status != 1 {
		stats.IncrCounter(ctx, "cb:drop", accountID)
		g.Log().Warningf(ctx, "callback: account %d not found or disabled", accountID)
		return nil, nil
	}

	// 2. 查找行为映射
	mappings, err := cache.GetMappings(ctx, accountID)
	if err != nil {
		stats.IncrCounter(ctx, "cb:drop", accountID)
		return nil, nil
	}
	var mappedType string
	var taolianDesc string
	for _, m := range mappings {
		if m.TaolianType == req.TransformType && m.Enabled {
			mappedType = m.PlatformType
			taolianDesc = m.TaolianDesc
			break
		}
	}
	if mappedType == "" {
		stats.IncrCounter(ctx, "cb:drop", accountID)
		g.Log().Warningf(ctx, "callback: no mapping for account=%d transformType=%d", accountID, req.TransformType)
		return nil, nil
	}

	// 2.1 联调模式过滤
	if acct.DebugMode {
		var debugEvents []string
		_ = json.Unmarshal([]byte(acct.DebugEvents), &debugEvents)
		if len(debugEvents) > 0 {
			allowed := false
			for _, e := range debugEvents {
				if e == mappedType {
					allowed = true
					break
				}
			}
			if !allowed {
				g.Log().Infof(ctx, "callback: debug mode skip event=%s account=%d", mappedType, accountID)
				return nil, nil
			}
		}
	}

	// 3. 解出设备号 (从 callbackUrl 透传的 d 参数, base64 JSON)
	// 解码失败只记日志不中断: 设备号缺失仍可回传 (按 click_id 归因)
	var devices map[consts.InternalField]string
	if req.D != "" {
		devJSON, decErr := base64.RawURLEncoding.DecodeString(req.D)
		if decErr == nil {
			decErr = json.Unmarshal(devJSON, &devices)
		}
		if decErr != nil {
			g.Log().Warningf(ctx, "callback: decode d param failed: %v, d=%s", decErr, req.D)
		}
	}

	// 4. 设备号按事件类型写 CSV (异步)
	// CSV 命名: {事件名称}_{设备号类型}.csv, 如 淘客订单支付_imei_md5.csv
	// 分片后: 淘客订单支付_imei_md5_1.csv
	if c.dedup != nil && len(devices) > 0 {
		go c.dedup.ProcessClickDevices(context.Background(), taolianDesc, devices)
	}

	// 5. 入真实回传队列
	if err := queue.EnqueueCallbackReport(ctx, &bo.CallbackTaskPayload{
		AccountID:   accountID,
		ClickID:     req.C,
		TaolianType: req.TransformType,
		MappedType:  mappedType,
		Devices:     devices,
		TS:          time.Now().UnixMilli(),
	}); err != nil {
		g.Log().Warningf(ctx, "enqueue callback failed: %v", err)
	}

	return nil, nil
}
