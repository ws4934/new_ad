package xiaohongshu

import (
	"context"
	"encoding/json"
	"strings"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/adapter"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/logic/httpx"
	"ad-monitor/internal/model/bo"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

type Adapter struct{}

func init() {
	adapter.Register(&Adapter{})
}

func (a *Adapter) Name() string            { return "xiaohongshu" }
func (a *Adapter) Spec() adapter.MacroSpec { return Spec }

// ReportConversion 向小红书回传转化
// 接口: POST JSON https://adapi.xiaohongshu.com/api/open/conversion
// 对接文档: docs/platform-docs/小红书api上报.md
func (a *Adapter) ReportConversion(ctx context.Context, p *bo.CallbackPayload) (*bo.ReportResult, error) {
	platform, err := cache.GetPlatform(ctx, p.Account.PlatformId)
	if err != nil {
		return nil, err
	}
	var pcfg map[string]any
	if err := json.Unmarshal([]byte(platform.Config), &pcfg); err != nil {
		return nil, err
	}
	reportURL, _ := pcfg["reportUrl"].(string)
	if reportURL == "" {
		reportURL = "https://adapi.xiaohongshu.com/api/open/conversion"
	}

	var creds map[string]any
	_ = json.Unmarshal([]byte(p.Account.Credentials), &creds)

	body := map[string]any{
		"click_id":   p.ClickID,
		"event_type": gconv.Int(p.MappedType),
		"timestamp":  p.ConvTS, // 事件发生时间(毫秒), 重试不变
		"platform":   gconv.String(creds["platform"]),
		"scene":      "700", // 固定 唤端项目
	}

	// advertiser_id 从监测链接宏 __ADVERTISER_ID__ 获取 (透传到 Devices map),
	// 不从后台配置读。对接文档里该字段选填, 有值就传。
	if aid := p.Devices[consts.FieldAdvertiserID]; aid != "" && !adapter.IsMacroValue(aid) {
		body["advertiser_id"] = aid
	}

	g.Log().Debugf(ctx, "xhs report: %s body=%v", reportURL, body)

	resp, err := httpx.New().ContentJson().Post(ctx, reportURL, body)
	if err != nil {
		return &bo.ReportResult{Success: false}, err
	}
	defer resp.Close()

	rawBody := resp.ReadAllString()
	result := &bo.ReportResult{
		HTTPStatus: resp.StatusCode,
		RawBody:    rawBody,
	}
	if resp.StatusCode == 200 && (strings.Contains(rawBody, `"code":0`) || strings.Contains(rawBody, `"success":true`)) {
		result.Success = true
	}
	return result, nil
}
