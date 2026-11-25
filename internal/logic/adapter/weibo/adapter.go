package weibo

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/adapter"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/logic/httpx"
	"ad-monitor/internal/model/bo"

	"github.com/gogf/gf/v2/frame/g"
)

type Adapter struct{}

func init() {
	adapter.Register(&Adapter{})
}

func (a *Adapter) Name() string            { return "weibo" }
func (a *Adapter) Spec() adapter.MacroSpec { return Spec }

// ReportConversion 向微博回传转化
// 接口: GET https://appmonitor.biz.weibo.com/sdkserver/active
// 对接文档: docs/platform-docs/微博广告api上报文档.md
func (a *Adapter) ReportConversion(ctx context.Context, p *bo.CallbackPayload) (*bo.ReportResult, error) {
	// 从 cache 读平台配置 (reportUrl 来源)
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
		reportURL = "https://appmonitor.biz.weibo.com/sdkserver/active"
	}

	// 账户凭证 (company 等)
	var creds map[string]any
	_ = json.Unmarshal([]byte(p.Account.Credentials), &creds)

	params := url.Values{}
	company, _ := creds["company"].(string)
	params.Set("company", company)
	params.Set("IMP", p.ClickID)
	params.Set("action_type", p.MappedType)
	// 文档: 重传时 active_time 必须完全一致, 否则按多次激活处理
	params.Set("active_time", strconv.FormatInt(p.ConvTS/1000, 10))

	// 归因设备号 (只填值存在且非宏的)
	setIf := func(k, v string) {
		if v != "" && !adapter.IsMacroValue(v) {
			params.Set(k, v)
		}
	}
	setIf("imei", p.Devices[consts.FieldIMEIMD5])
	setIf("oaid", p.Devices[consts.FieldOAIDMD5])
	setIf("idfa", p.Devices[consts.FieldIDFAMD5])
	// 升级前的点击, 回调 d 参数里原值还放在 caid_md5 键下
	gxCaid := p.Devices[consts.FieldCAID]
	if gxCaid == "" {
		gxCaid = p.Devices[consts.FieldCAIDMD5]
	}
	setIf("GX_CAID", gxCaid)
	setIf("ip", p.Devices[consts.FieldIP])
	setIf("ua", p.Devices[consts.FieldUA])

	fullURL := reportURL + "?" + params.Encode()
	g.Log().Debugf(ctx, "weibo report: %s", fullURL)

	resp, err := httpx.New().Get(ctx, fullURL)
	if err != nil {
		return &bo.ReportResult{Success: false}, err
	}
	defer resp.Close()

	body := resp.ReadAllString()
	result := &bo.ReportResult{
		HTTPStatus: resp.StatusCode,
		RawBody:    body,
	}
	// 成功判定: 对接文档明确 — HTTP 200 或 {"result":"OK", "Code":"0"}
	// 我们宽松一点: HTTP 200 且 body 包含 "OK" 或 "\"Code\":\"0\"" 或 body 为空
	if resp.StatusCode == 200 && (body == "" || strings.Contains(body, "OK") || strings.Contains(body, `"Code":"0"`)) {
		result.Success = true
	}
	return result, nil
}
