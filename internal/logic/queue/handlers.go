package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ad-monitor/internal/logic/adapter"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/logic/reporter"
	"ad-monitor/internal/logic/stats"
	"ad-monitor/internal/model/bo"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hibiken/asynq"
)

// retryIf5xx 平台返回 5xx 属于临时故障, 返回 error 让 asynq 按退避重试;
// 2xx/4xx 业务性失败 (参数错误等) 重试也无意义, 返回 nil 不重试
func retryIf5xx(status int, kind string) error {
	if status >= 500 {
		return fmt.Errorf("%s report got http %d, retry", kind, status)
	}
	return nil
}

// countErr 失败只在任务最终失败时计一次 (不再重试: SkipRetry / nil / 已到最大重试次数),
// 否则重试会让 成功+失败 大于任务数。返回 ret 本身, 由 asynq 决定是否重试
func countErr(ctx context.Context, kind string, accountID int, ret error) error {
	retried, _ := asynq.GetRetryCount(ctx)
	maxRetry, ok := asynq.GetMaxRetry(ctx)
	if ret == nil || errors.Is(ret, asynq.SkipRetry) || !ok || retried >= maxRetry {
		stats.IncrCounter(ctx, kind+":err", accountID)
	}
	return ret
}

// HandleTaolianReport 淘联上报任务处理器
// asynq worker 触发, 调用 reporter.Report 发 HTTP 到淘联
func HandleTaolianReport(ctx context.Context, t *asynq.Task) error {
	var p bo.TaolianTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	acct, err := cache.GetAccount(ctx, p.AccountID)
	if err != nil || acct == nil {
		return countErr(ctx, "taolian", p.AccountID, asynq.SkipRetry) // 账户不存在, 不重试
	}
	platform, err := cache.GetPlatform(ctx, acct.PlatformId)
	if err != nil {
		return countErr(ctx, "taolian", p.AccountID, asynq.SkipRetry)
	}

	rep := reporter.NewReporterFromConfig(ctx)
	req := reporter.BuildReqFromAccount(ctx, acct, platform, p.ClickID, p.Devices, p.TS)
	result, err := rep.Report(ctx, req)
	if err != nil || result == nil || !result.Success {
		rawBody := ""
		if result != nil {
			rawBody = result.RawBody
		}
		g.Log().Warningf(ctx, "taolian report failed: err=%v body=%s", err, rawBody)
		if err != nil {
			return countErr(ctx, "taolian", p.AccountID, err) // 网络/解析错误 → asynq 重试
		}
		if result != nil {
			return countErr(ctx, "taolian", p.AccountID, retryIf5xx(result.HTTPStatus, "taolian")) // 5xx 临时故障 → 重试
		}
		return countErr(ctx, "taolian", p.AccountID, nil)
	}
	stats.IncrCounter(ctx, "taolian:ok", p.AccountID)
	return nil
}

// HandleCallbackReport 广告平台真实回传任务
func HandleCallbackReport(ctx context.Context, t *asynq.Task) error {
	var p bo.CallbackTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	kind := "callback" // 联调模式模拟的回传单独计数, 不算真实回传
	if p.Debug {
		kind = "debug"
	}

	acct, err := cache.GetAccount(ctx, p.AccountID)
	if err != nil || acct == nil {
		return countErr(ctx, kind, p.AccountID, asynq.SkipRetry)
	}
	platform, err := cache.GetPlatform(ctx, acct.PlatformId)
	if err != nil {
		return countErr(ctx, kind, p.AccountID, asynq.SkipRetry)
	}

	ad := adapter.Get(platform.Name)
	if ad == nil {
		g.Log().Warningf(ctx, "adapter not found: %s", platform.Name)
		return countErr(ctx, kind, p.AccountID, asynq.SkipRetry)
	}

	payload := &bo.CallbackPayload{
		Account:     acct,
		ClickID:     p.ClickID,
		Devices:     p.Devices,
		MappedType:  p.MappedType,
		TaolianType: p.TaolianType,
		ConvTS:      convTS(p.TS),
		IsFake:      false,
	}
	result, err := ad.ReportConversion(ctx, payload)
	if err != nil || result == nil || !result.Success {
		rawBody := ""
		if result != nil {
			rawBody = result.RawBody
		}
		g.Log().Warningf(ctx, "callback report failed: err=%v body=%s", err, rawBody)
		if err != nil {
			return countErr(ctx, kind, p.AccountID, err)
		}
		if result != nil {
			return countErr(ctx, kind, p.AccountID, retryIf5xx(result.HTTPStatus, "callback"))
		}
		return countErr(ctx, kind, p.AccountID, nil)
	}
	stats.IncrCounter(ctx, kind+":ok", p.AccountID)
	return nil
}

// HandleFakeReport 虚假回传任务
// 二次校验开关: 用户可能在延迟期间关闭 fake_enabled
func HandleFakeReport(ctx context.Context, t *asynq.Task) error {
	var p bo.FakeTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	acct, err := cache.GetAccount(ctx, p.AccountID)
	if err != nil || acct == nil {
		return countErr(ctx, "fake", p.AccountID, asynq.SkipRetry)
	}

	// 二次校验开关
	if !acct.FakeEnabled {
		g.Log().Infof(ctx, "fake report skipped: account %d disabled during delay", p.AccountID)
		return nil
	}

	platform, err := cache.GetPlatform(ctx, acct.PlatformId)
	if err != nil {
		return countErr(ctx, "fake", p.AccountID, asynq.SkipRetry)
	}
	ad := adapter.Get(platform.Name)
	if ad == nil {
		return countErr(ctx, "fake", p.AccountID, asynq.SkipRetry)
	}

	payload := &bo.CallbackPayload{
		Account:     acct,
		ClickID:     p.ClickID,
		Devices:     p.Devices,
		MappedType:  p.MappedType,
		TaolianType: p.TaolianType,
		ConvTS:      convTS(p.TS),
		IsFake:      true,
	}
	result, err := ad.ReportConversion(ctx, payload)
	if err != nil || result == nil || !result.Success {
		rawBody := ""
		if result != nil {
			rawBody = result.RawBody
		}
		g.Log().Warningf(ctx, "fake report failed: err=%v body=%s", err, rawBody)
		if err != nil {
			return countErr(ctx, "fake", p.AccountID, err)
		}
		if result != nil {
			return countErr(ctx, "fake", p.AccountID, retryIf5xx(result.HTTPStatus, "fake"))
		}
		return countErr(ctx, "fake", p.AccountID, nil)
	}
	stats.IncrCounter(ctx, "fake:ok", p.AccountID)
	return nil
}

// convTS 升级前已入队的任务没有 ts, 退回当前时间
func convTS(ts int64) int64 {
	if ts > 0 {
		return ts
	}
	return time.Now().UnixMilli()
}

// HandleArchiveDaily 每日归档 (阶段 8 完善)
func HandleArchiveDaily(ctx context.Context, t *asynq.Task) error {
	return stats.ArchiveYesterday(ctx)
}
