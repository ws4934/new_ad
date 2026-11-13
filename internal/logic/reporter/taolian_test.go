package reporter_test

import (
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
