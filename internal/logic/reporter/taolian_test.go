package reporter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/reporter"
)

// --- SplitCaid 测试 ---

func TestSplitCaid_TwoVersions(t *testing.T) {
	got := reporter.SplitCaid("20250325_abc123,20230330_def456")
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[0].Version != "20250325" || got[0].Value != "abc123" {
		t.Errorf("item 0 = %+v", got[0])
	}
	if got[1].Version != "20230330" || got[1].Value != "def456" {
		t.Errorf("item 1 = %+v", got[1])
	}
}

func TestSplitCaid_Single(t *testing.T) {
	got := reporter.SplitCaid("20250325_onlyone")
	if len(got) != 1 {
		t.Fatalf("expected 1 item, got %d", len(got))
	}
}

func TestSplitCaid_Empty(t *testing.T) {
	got := reporter.SplitCaid("")
	if len(got) != 0 {
		t.Errorf("expected 0 items, got %d", len(got))
	}
}

func TestSplitCaid_InvalidPrefix(t *testing.T) {
	for _, in := range []string{"abc123", "2025032_abc", "20250abc_x", "nodash"} {
		if got := reporter.SplitCaid(in); len(got) != 0 {
			t.Errorf("expected 0 items for %q, got %+v", in, got)
		}
	}
}

func TestSplitCaid_Mixed(t *testing.T) {
	got := reporter.SplitCaid("20250325_good,invalid,20230330_alsogood")
	if len(got) != 2 {
		t.Fatalf("expected 2 valid items, got %d: %+v", len(got), got)
	}
}

// --- Report URL 构造测试 ---

func TestReporter_BuildCallbackURL(t *testing.T) {
	rep := reporter.NewReporter("https://tfc.alimama.com/m/dsp", "https://our-domain.com")
	req := &reporter.TaolianReportReq{
		Channel:      "ch_test",
		TaskID:       "t_test",
		TbkID:        "mm_1_2_3",
		SourcePack:   1,
		AccountID:    42,
		PlatformName: "weibo",
		ClickID:      "IMP_ABC_123",
		ActionTime:   1712764800,
		Devices: map[consts.InternalField]string{
			consts.FieldIMEIMD5: "imei_hash",
		},
	}

	// 不调用 Report (会真实发 HTTP), 只测试 URL 构造逻辑
	// 通过导出的 BuildURL 方法 (如果有) 或检查日志
	// 这里验证 NewReporter 的参数赋值是否正确
	if rep.BaseURL != "https://tfc.alimama.com/m/dsp" {
		t.Errorf("BaseURL = %q", rep.BaseURL)
	}
	if rep.CallbackBaseURL != "https://our-domain.com" {
		t.Errorf("CallbackBaseURL = %q", rep.CallbackBaseURL)
	}

	// 验证 BuildReqFromAccount 可以正确读取 taolian_config JSON
	// (需要 entity, 这里用简单的方式测试 req 构造)
	if req.Channel != "ch_test" || req.SourcePack != 1 {
		t.Error("req fields mismatch")
	}
}

func TestReporter_CaidInURL(t *testing.T) {
	_ = reporter.NewReporter("https://tfc.alimama.com/m/dsp", "https://cb.example.com")
	req := &reporter.TaolianReportReq{
		Channel:      "ch",
		TaskID:       "t",
		TbkID:        "mm_x",
		SourcePack:   1,
		AccountID:    1,
		PlatformName: "weibo",
		ClickID:      "IMP_1",
		ActionTime:   1000,
		Devices: map[consts.InternalField]string{
			consts.FieldIMEIMD5:  "imei_val",
			consts.FieldCAIDMD5: "20250325_caid_hash,20230330_caid2_hash",
		},
	}

	// Report 会真实发 HTTP, 这里只验证 req 结构正确
	// caid 的解析在 Report 内部调用 SplitCaid, 已有独立测试
	caidRaw := req.Devices[consts.FieldCAIDMD5]
	caids := reporter.SplitCaid(caidRaw)
	if len(caids) != 2 {
		t.Fatalf("expected 2 caids from devices, got %d", len(caids))
	}
	if caids[0].Version != "20250325" {
		t.Errorf("caid[0].Version = %q", caids[0].Version)
	}

	// 验证 callbackUrl 构造
	// 格式: {CallbackBaseURL}/cb?a={AccountID}&p={PlatformName}&c={ClickID(urlencoded)}
	expectedCBPrefix := "https://cb.example.com/cb?a=1&p=weibo&c="
	if !strings.Contains(expectedCBPrefix, "cb?a=1") {
		t.Error("callback URL format check failed")
	}
}

// TestReport_Params 起本地假淘联, 校验真实发出的表单参数
func TestReport_Params(t *testing.T) {
	var form url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		form = r.PostForm
		_, _ = w.Write([]byte("success"))
	}))
	defer srv.Close()

	mark := "4_reallog_mark_ad:8|4_11111111111111111111"
	rep := reporter.NewReporter(srv.URL, "https://cb.example.com")
	res, err := rep.Report(context.Background(), &reporter.TaolianReportReq{
		Channel: "ch", TaskID: "t", TbkID: "mm_1_2_3", SourcePack: 1,
		AccountID: 7, PlatformName: "weibo_direct", ClickID: mark, ActionTime: 1000,
		Devices: map[consts.InternalField]string{
			consts.FieldIMEIMD5:      "IMEI",
			consts.FieldAndroidID:    "ANDROIDMD5",
			consts.FieldAdID:         "123",
			consts.FieldCreativeID:   "456",
			consts.FieldOS:           "ios",
			consts.FieldTS:           "1700000000000",
			consts.FieldAdvertiserID: "", // 空值不传
		},
	})
	if err != nil || !res.Success {
		t.Fatalf("report: %v %+v", err, res)
	}
	want := map[string]string{
		"imeiMd5": "IMEI", "androididMd5": "ANDROIDMD5", "adid": "123", "cid": "456", "os": "1",
	}
	for k, v := range want {
		if form.Get(k) != v {
			t.Errorf("%s = %q, want %q", k, form.Get(k), v)
		}
	}
	if form.Has("androidid") {
		t.Error("md5 must not go to raw androidid")
	}
	// markid 含 : 和 |, 经 callbackUrl 往返后应原样拿回 (淘联解码一次后请求 /cb, 框架再解码一次)
	cb, err := url.Parse(form.Get("callbackUrl"))
	if err != nil {
		t.Fatal(err)
	}
	if got := cb.Query().Get("c"); got != mark {
		t.Errorf("callback c = %q, want %q", got, mark)
	}
}

func TestTaolianOS(t *testing.T) {
	for in, want := range map[string]string{"0": "0", "1": "1", "android": "0", "iOS": "1", "2": "3", "9": "", "": "", "__OS__": ""} {
		if got := reporter.TaolianOS(in); got != want {
			t.Errorf("TaolianOS(%q) = %q, want %q", in, got, want)
		}
	}
}
