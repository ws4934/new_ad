package xiaohongshu

import (
	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/adapter"
)

// Spec 小红书监测链接宏元数据 (__XXX__ 格式)
// 对接文档: docs/platform-docs/小红书监测链接说明文档.md
// 注意: 小红书 __IDFA__ / __IMEI__ 字段值本身就是 MD5 (命名迷惑),
//       映射到内部 FieldIDFAMD5 / FieldIMEIMD5
var Spec = adapter.MacroSpec{
	PlatformName: "xiaohongshu",
	Fields: []adapter.MacroField{
		{Name: "click_id", Macro: "__CLICK_ID__", Internal: consts.FieldClickID, Required: true},
		{Name: "idfa", Macro: "__IDFA__", Internal: consts.FieldIDFAMD5}, // idfa md5
		{Name: "imei", Macro: "__IMEI__", Internal: consts.FieldIMEIMD5}, // imei md5
		{Name: "oaid_md5", Macro: "__OAID_MD5__", Internal: consts.FieldOAIDMD5},
		{Name: "caid_md5", Macro: "__CAID_MD5__", Internal: consts.FieldCAIDMD5},
		{Name: "caid", Macro: "__CAID__", Internal: consts.FieldCAID},
		{Name: "os", Macro: "__OS__", Internal: consts.FieldOS},
		{Name: "ts", Macro: "__TS__", Internal: consts.FieldTS},
		{Name: "ip", Macro: "__IP__", Internal: consts.FieldIP},
		{Name: "ua", Macro: "__UA__", Internal: consts.FieldUA},
		{Name: "advertiser_id", Macro: "__ADVERTISER_ID__", Internal: consts.FieldAdvertiserID},
	},
}
