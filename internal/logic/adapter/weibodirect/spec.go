package weibodirect

import (
	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/adapter"
)

// Spec 微博直达监测链接宏元数据 (__XXX__ 格式)
// 对接文档: docs/platform-docs/微博直达资产联调手册_副本.md 3.1
// 直达文档只列了 MARKID/TS/IP; 淘联点击上报要求至少一个设备号, 所以再带上
// 微博通用宏表 (微博监测链接文档.md) 的设备宏, 未替换的会被自动过滤
var Spec = adapter.MacroSpec{
	PlatformName: "weibo_direct",
	Fields: []adapter.MacroField{
		{Name: "markid", Macro: "__MARKID__", Internal: consts.FieldClickID, Required: true},
		{Name: "ip", Macro: "__IP__", Internal: consts.FieldIP},
		{Name: "clicktime", Macro: "__TS__", Internal: consts.FieldTS},
		{Name: "idfa_md5", Macro: "__IDFA_MD5__", Internal: consts.FieldIDFAMD5},
		{Name: "imei_md5", Macro: "__IMEI__", Internal: consts.FieldIMEIMD5},
		{Name: "oaid_md5", Macro: "__OAID_MD5__", Internal: consts.FieldOAIDMD5},
		{Name: "androidid_MD5", Macro: "__ANDROIDID__", Internal: consts.FieldAndroidID},
		{Name: "GX_CAID", Macro: "__CAID__", Internal: consts.FieldCAIDMD5},
		{Name: "ua", Macro: "__UA__", Internal: consts.FieldUA},
		{Name: "devicetype", Macro: "__OS__", Internal: consts.FieldOS}, // iOS=1, android=0
		{Name: "ad_id", Macro: "__AID__", Internal: consts.FieldAdID},
		{Name: "creative_id", Macro: "__CID__", Internal: consts.FieldCreativeID},
	},
}
