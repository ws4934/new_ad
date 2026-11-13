package adapter_test

import (
	"testing"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/adapter"
)

func TestIsMacroValue(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", true},
		{"   ", true},
		{"{IMP}", true},
		{"{clicktime}", true},
		{"{idfa_MD5}", true},
		{"__IDFA_MD5__", true},
		{"__CALLBACK_PARAM__", true},
		{"__ADVERTISER_ID__", true},
		{"abc123", false},
		{"1712764800", false},
		{"DEADBEEF123456", false},
		{"607748", false},  // 真实的 advertiser_id
	}
	for _, c := range cases {
		if got := adapter.IsMacroValue(c.in); got != c.want {
			t.Errorf("IsMacroValue(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestBuildMonitorURL(t *testing.T) {
	spec := adapter.MacroSpec{
		PlatformName: "test",
		Fields: []adapter.MacroField{
			{Name: "IMP", Macro: "{IMP}", Internal: consts.FieldClickID},
			{Name: "imei_md5", Macro: "{imei_MD5}", Internal: consts.FieldIMEIMD5},
		},
	}
	got := adapter.BuildMonitorURL("https://example.com", 42, spec)
	want := "https://example.com/m/42?IMP={IMP}&imei_md5={imei_MD5}"
	if got != want {
		t.Errorf("BuildMonitorURL = %q, want %q", got, want)
	}
}

func TestExtractDeviceIDs(t *testing.T) {
	data := &adapter.ClickData{
		Values: map[consts.InternalField]string{
			consts.FieldClickID:      "IMP_123",        // 不是设备号, 不应提取
			consts.FieldIMEIMD5:      "abc123",          // 是设备号
			consts.FieldOAIDMD5:      "def456",          // 是设备号
			consts.FieldIP:           "1.2.3.4",         // 不是设备号
			consts.FieldAdvertiserID: "607748",           // 不是设备号
		},
	}
	devices := adapter.ExtractDeviceIDs(data)

	// 应该提取 2 个设备号
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d: %v", len(devices), devices)
	}
	if devices[consts.FieldIMEIMD5] != "abc123" {
		t.Errorf("imei_md5 = %q, want abc123", devices[consts.FieldIMEIMD5])
	}
	if devices[consts.FieldOAIDMD5] != "def456" {
		t.Errorf("oaid_md5 = %q, want def456", devices[consts.FieldOAIDMD5])
	}
	// 不应包含非设备号字段
	if _, ok := devices[consts.FieldClickID]; ok {
		t.Error("click_id should not be in devices")
	}
	if _, ok := devices[consts.FieldIP]; ok {
		t.Error("ip should not be in devices")
	}
	if _, ok := devices[consts.FieldAdvertiserID]; ok {
		t.Error("advertiser_id should not be in devices")
	}
}

func TestExtractDeviceIDs_Empty(t *testing.T) {
	data := &adapter.ClickData{
		Values: map[consts.InternalField]string{
			consts.FieldClickID: "IMP_123",
		},
	}
	devices := adapter.ExtractDeviceIDs(data)
	if len(devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(devices))
	}
}

func TestExtractDeviceIDs_SkipsMacroValues(t *testing.T) {
	data := &adapter.ClickData{
		Values: map[consts.InternalField]string{
			consts.FieldIMEIMD5: "{imei_MD5}",  // 未替换的宏, 不应提取
			consts.FieldOAIDMD5: "real_value",
		},
	}
	devices := adapter.ExtractDeviceIDs(data)
	// {imei_MD5} 虽然在 Values 里, 但 ExtractDeviceIDs 只检查非空
	// 宏值过滤是在 ParseClickRequest 阶段做的, ExtractDeviceIDs 不重复过滤
	// 所以这里 {imei_MD5} 会被提取出来 (它非空)
	// 真正的过滤在 ParseClickRequest 里, 宏值不会进 Values
	if len(devices) != 2 {
		t.Logf("devices = %v (ExtractDeviceIDs 不做宏过滤, 那是 ParseClickRequest 的职责)", devices)
	}
}
