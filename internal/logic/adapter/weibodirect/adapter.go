package weibodirect

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"net/url"
	"strconv"
	"sync"
	"time"

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

	fullURL := reportURL + "?" + buildQuery(p, cbHost(), creds, time.Now())
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

// cbHost 回传参数 host = 我方回调域名 (配置不会热变, 只解析一次)
var cbHost = sync.OnceValue(func() string {
	v, _ := g.Cfg().Get(context.Background(), "adMonitor.callbackBaseUrl")
	u, err := url.Parse(v.String())
	if err != nil {
		return ""
	}
	return u.Hostname()
})

// buildQuery time 为调用接口时的毫秒时间戳 (文档要求); mark_id / item_order_pay 由 Encode
// 做一次 urlencode, 微博服务端会 decode 一次 (多次 encode 会报 mark_id may be expired)
func buildQuery(p *bo.CallbackPayload, host string, c credentials, now time.Time) string {
	v := url.Values{}
	v.Set("time", strconv.FormatInt(now.UnixMilli(), 10))
	v.Set("behavior", p.MappedType)
	v.Set("mark_id", p.ClickID)
	if host != "" {
		v.Set("host", host)
	}
	if c.TechnicalProvider != "" {
		v.Set("technical_provider", c.TechnicalProvider)
	}
	if p.MappedType == behaviorPayOrder {
		// click_time 用监测入口收到点击的时间 (经淘联 callbackUrl 透传回来)
		clickMs, _ := strconv.ParseInt(p.Devices[consts.FieldTS], 10, 64)
		b, _ := json.Marshal(genOrder(c.Order, c.ServiceProvider, clickMs, p.ConvTS, orderRand(p)))
		v.Set("item_order_pay", string(b))
	}
	return v.Encode()
}

// orderRand 同一转化 (点击+行为+转化时间) 的种子固定, 重试时订单号等不变, 避免微博重复计数
func orderRand(p *bo.CallbackPayload) *rand.Rand {
	h := fnv.New64a()
	fmt.Fprintf(h, "%s|%s|%d", p.ClickID, p.MappedType, p.ConvTS)
	return rand.New(rand.NewPCG(h.Sum64(), 0))
}

// isOK 成功应答: {"code":0,"message":"OK"}
func isOK(body string) bool {
	var r struct {
		Code *int `json:"code"`
	}
	return json.Unmarshal([]byte(body), &r) == nil && r.Code != nil && *r.Code == 0
}
