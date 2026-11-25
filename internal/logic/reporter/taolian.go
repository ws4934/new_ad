package reporter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/httpx"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

// CaidItem 一条 caid 版本项
type CaidItem struct {
	Version string
	Value   string
}

// SplitCaid 解析 caid 字符串为带版本的多条记录。
//
// 广告平台 (微博/小红书) 给我方的 CAID 格式统一为:
//
//	"20250325_md5xxx,20230330_md5yyy"
//
// 版本号就是值前面的 8 位数字 + 下划线前缀。我方完全按此自解析, 不依赖任何
// 后台配置。不符合 "8位数字+下划线" 前缀的条目直接丢弃 (而不是编一个默认版本号
// 上报给淘联 — 那会导致归因错误)。
//
// 对接文档依据:
//   - docs/platform-docs/微博监测链接文档.md: 2025/6/30 起 CAID 切换为
//     "20250325_xxx,20230330_xxx" 格式
//   - docs/platform-docs/小红书监测链接说明文档.md: CAID 格式
//     "version1_caid1,version2_caid2"
func SplitCaid(raw string) []CaidItem {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]CaidItem, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		idx := strings.Index(p, "_")
		if idx != 8 || !isAllDigits(p[:idx]) {
			// 不符合 "8位数字_值" 格式, 丢弃 (不能编版本号)
			continue
		}
		out = append(out, CaidItem{Version: p[:idx], Value: p[idx+1:]})
	}
	return out
}

func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isMacroValue 轻量版宏判断 (避免循环引用 adapter 包)
func isMacroValue(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return true
	}
	if len(v) >= 4 && strings.HasPrefix(v, "__") && strings.HasSuffix(v, "__") {
		return true
	}
	if len(v) >= 2 && strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
		return true
	}
	return false
}

// taolianOS 各平台系统宏统一为淘联枚举 (0 Android / 1 iOS / 3 鸿蒙)。
// 微博 __OS__ 是 0/1, {devicetype} 是 android/ios; 小红书 __OS__ 0/1/2 (2=鸿蒙, 淘联鸿蒙是 3);
// 其他值一律不传
func taolianOS(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "0", "android":
		return "0"
	case "1", "ios":
		return "1"
	case "2":
		return "3"
	}
	return ""
}

// TaolianReportReq 淘联上报请求参数
type TaolianReportReq struct {
	Channel      string
	TaskID       string
	TbkID        string
	SourcePack   int
	Devices      map[consts.InternalField]string
	AccountID    int
	PlatformName string
	ClickID      string
	ActionTime   int64
}

// TaolianResult 淘联上报结果
type TaolianResult struct {
	Success    bool
	HTTPStatus int
	RawBody    string
}

// Reporter 淘联上报器
type Reporter struct {
	BaseURL         string // 如 https://tfc.alimama.com/m/dsp
	CallbackBaseURL string // 如 https://our-domain.com
}

// NewReporter 构造
func NewReporter(baseURL, cbBaseURL string) *Reporter {
	return &Reporter{BaseURL: baseURL, CallbackBaseURL: cbBaseURL}
}

var (
	defaultReporter     *Reporter
	defaultReporterOnce sync.Once
)

// NewReporterFromConfig 从全局配置构造, 单例复用避免每次任务都读配置
func NewReporterFromConfig(ctx context.Context) *Reporter {
	defaultReporterOnce.Do(func() {
		cfg := g.Cfg()
		defaultReporter = NewReporter(
			cfg.MustGet(ctx, "adMonitor.taolian.reportUrl").String(),
			cfg.MustGet(ctx, "adMonitor.callbackBaseUrl").String(),
		)
	})
	return defaultReporter
}

// Report 上报点击到淘联 (action=2)
func (r *Reporter) Report(ctx context.Context, req *TaolianReportReq) (*TaolianResult, error) {
	// 1. 构造 callbackUrl
	// query: a=accountId, p=platformName, c=clickId, d=设备号(base64 JSON)
	// 设备号透传到回传入口, 用于按事件类型写 CSV
	devJSON, _ := json.Marshal(req.Devices)
	devB64 := base64.RawURLEncoding.EncodeToString(devJSON)
	cbURL := fmt.Sprintf("%s/cb?a=%d&p=%s&c=%s&d=%s",
		r.CallbackBaseURL, req.AccountID, req.PlatformName, url.QueryEscape(req.ClickID), devB64)

	params := url.Values{}
	params.Set("action", "2")
	params.Set("channel", req.Channel)
	params.Set("taskId", req.TaskID)
	params.Set("tbkId", req.TbkID)
	params.Set("sourcePack", strconv.Itoa(req.SourcePack))
	params.Set("callbackUrl", cbURL) // url.Values.Encode() 会自动 urlencode
	params.Set("actionTime", strconv.FormatInt(req.ActionTime, 10))

	// 2. 填充设备号 (按淘联字段名)
	setIf := func(k, v string) {
		if v != "" && !isMacroValue(v) {
			params.Set(k, v)
		}
	}
	setIf("imeiMd5", req.Devices[consts.FieldIMEIMD5])
	setIf("idfaMd5", req.Devices[consts.FieldIDFAMD5])
	setIf("oaidMd5", req.Devices[consts.FieldOAIDMD5])
	// 微博 androidid_MD5 是 md5 值, 对应淘联 androididMd5 (不是原值 androidid)
	setIf("androididMd5", req.Devices[consts.FieldAndroidID])

	// 广告投放字段: 淘联文档要求务必填写, 以免转化归因遗漏
	setIf("adid", req.Devices[consts.FieldAdID])
	setIf("cid", req.Devices[consts.FieldCreativeID])
	setIf("os", taolianOS(req.Devices[consts.FieldOS]))

	// 3. CAID 多版本拆分 (版本号自动从值前缀解析, 不依赖配置)
	if raw := req.Devices[consts.FieldCAIDMD5]; raw != "" && !isMacroValue(raw) {
		caids := SplitCaid(raw)
		if len(caids) >= 1 {
			params.Set("caidMd5", caids[0].Value)
			params.Set("caidVersion", caids[0].Version)
		}
		if len(caids) >= 2 {
			params.Set("caid2Md5", caids[1].Value)
			params.Set("caid2Version", caids[1].Version)
		}
	}

	// 4. 调用淘联 (POST form, 对接文档支持 GET/POST)
	g.Log().Debugf(ctx, "taolian report POST: %s params=%s", r.BaseURL, params.Encode())

	// 必须传 Encode() 后的字符串: gclient 会把 url.Values 的 []string 值序列化成 `["x"]`
	resp, err := httpx.New().ContentType("application/x-www-form-urlencoded").Post(ctx, r.BaseURL, params.Encode())
	if err != nil {
		return &TaolianResult{Success: false}, err
	}
	defer resp.Close()

	body := strings.TrimSpace(resp.ReadAllString())
	// 淘联成功判定 (兼容两种响应格式):
	// 1. 旧格式: 返回纯字符串 "success"
	// 2. 新格式: 返回 JSON {"code":0,"message":""}
	success := false
	if resp.StatusCode == 200 {
		success = body == "success" ||
			strings.Contains(body, `"code":0`) ||
			strings.Contains(body, `"code": 0`)
	}
	return &TaolianResult{
		Success:    success,
		HTTPStatus: resp.StatusCode,
		RawBody:    body,
	}, nil
}

// BuildReqFromAccount 从账户配置构造淘联请求
// 供 queue handler 使用, 封装 JSON 解析
func BuildReqFromAccount(
	ctx context.Context,
	acct *entity.AdAccount,
	platform *entity.AdPlatform,
	clickID string,
	devices map[consts.InternalField]string,
	ts int64,
) *TaolianReportReq {
	_ = ctx

	var tcfg map[string]any
	_ = json.Unmarshal([]byte(acct.TaolianConfig), &tcfg)

	if ts == 0 {
		ts = time.Now().Unix()
	}

	return &TaolianReportReq{
		Channel:      gconv.String(tcfg["channel"]),
		TaskID:       gconv.String(tcfg["taskId"]),
		TbkID:        gconv.String(tcfg["tbkId"]),
		SourcePack:   gconv.Int(tcfg["sourcePack"]),
		Devices:      devices,
		AccountID:    acct.Id,
		PlatformName: platform.Name,
		ClickID:      clickID,
		ActionTime:   ts,
	}
}
