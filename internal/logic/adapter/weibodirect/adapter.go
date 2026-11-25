package weibodirect

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/url"
	"strconv"
	"time"

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

func (a *Adapter) Name() string            { return "weibo_direct" }
func (a *Adapter) Spec() adapter.MacroSpec { return Spec }

const defaultReportURL = "https://api.biz.weibo.com/v4/track/activate"

// ReportConversion 向微博回传直达转化 (免鉴权)
// 接口: GET https://api.biz.weibo.com/v4/track/activate?time=&behavior=&mark_id=
// 对接文档: docs/platform-docs/微博直达资产联调手册_副本.md 3.2
func (a *Adapter) ReportConversion(ctx context.Context, p *bo.CallbackPayload) (*bo.ReportResult, error) {
	platform, err := cache.GetPlatform(ctx, p.Account.PlatformId)
	if err != nil {
		return nil, err
	}
	var pcfg map[string]any
	_ = json.Unmarshal([]byte(platform.Config), &pcfg)
	reportURL, _ := pcfg["reportUrl"].(string)
	if reportURL == "" {
		reportURL = defaultReportURL
	}

	var creds credentials
	_ = json.Unmarshal([]byte(p.Account.Credentials), &creds)
	cbBase, _ := g.Cfg().Get(ctx, "adMonitor.callbackBaseUrl")
	host := ""
	if u, err := url.Parse(cbBase.String()); err == nil {
		host = u.Hostname()
	}

	now := time.Now()
	fullURL := reportURL + "?" + buildQuery(p.ClickID, p.MappedType, host, creds, now, rand.New(rand.NewSource(now.UnixNano())))
	g.Log().Debugf(ctx, "weibo_direct report: %s", fullURL)

	resp, err := httpx.New().Get(ctx, fullURL)
	if err != nil {
		return &bo.ReportResult{Success: false}, err
	}
	defer resp.Close()

	raw := resp.ReadAllString()
	result := &bo.ReportResult{
		HTTPStatus: resp.StatusCode,
		RawBody:    raw,
		Success:    resp.StatusCode == 200 && isOK(raw),
	}
	// 文档建议记录 trace_id, 报错时反馈给微博排查
	if tid := resp.Header.Get("X-OPEN-TRACEID"); tid != "" {
		result.RawBody += " trace_id=" + tid
	}
	return result, nil
}

// behaviorPayOrder 付款订单商品数, 必须带 item_order_pay
const behaviorPayOrder = "2001"

// credentials 微博直达账户凭证 (ad_account.credentials)
type credentials struct {
	ServiceProvider   string      `json:"service_provider"`
	TechnicalProvider string      `json:"technical_provider"`
	Order             OrderConfig `json:"order"`
}

// buildQuery time 为调用接口时的毫秒时间戳; mark_id / item_order_pay 由 Encode
// 做一次 urlencode, 微博服务端会 decode 一次 (多次 encode 会报 mark_id may be expired)
func buildQuery(markID, behavior, host string, c credentials, now time.Time, r *rand.Rand) string {
	v := url.Values{}
	v.Set("time", strconv.FormatInt(now.UnixMilli(), 10))
	v.Set("behavior", behavior)
	v.Set("mark_id", markID)
	if host != "" {
		v.Set("host", host)
	}
	if c.TechnicalProvider != "" {
		v.Set("technical_provider", c.TechnicalProvider)
	}
	if behavior == behaviorPayOrder {
		b, _ := json.Marshal(genOrder(c.Order, c.ServiceProvider, now, r))
		v.Set("item_order_pay", string(b))
	}
	return v.Encode()
}

// isOK 成功应答: {"code":0,"message":"OK"}
func isOK(body string) bool {
	var r struct {
		Code *int `json:"code"`
	}
	return json.Unmarshal([]byte(body), &r) == nil && r.Code != nil && *r.Code == 0
}
