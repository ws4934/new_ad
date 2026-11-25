package weibo

import (
	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/adapter"
)

// Spec 微博监测链接宏元数据
// 使用 {} 格式宏 (对接文档同时支持 __XXX__ 和 {xxx}, 后者更直观)
// 字段对照来源: docs/platform-docs/微博监测链接文档.md 的宏表
var Spec = adapter.MacroSpec{
	PlatformName: "weibo",
	Fields: []adapter.MacroField{
		{Name: "IMP", Macro: "{IMP}", Internal: consts.FieldClickID, Required: true},
		{Name: "idfa_md5", Macro: "{idfa_MD5}", Internal: consts.FieldIDFAMD5},
		{Name: "imei_md5", Macro: "{imei_MD5}", Internal: consts.FieldIMEIMD5},
		{Name: "oaid_md5", Macro: "{oaid_MD5}", Internal: consts.FieldOAIDMD5},
		{Name: "androidid_MD5", Macro: "{androidid_MD5}", Internal: consts.FieldAndroidID},
		{Name: "GX_CAID", Macro: "{GX_CAID}", Internal: consts.FieldCAID}, // 中广协 caid 原值 (不是 md5), 见文档更新日志 2025/6/30
		{Name: "ip", Macro: "{ip}", Internal: consts.FieldIP},
		{Name: "ua", Macro: "{ua}", Internal: consts.FieldUA},
		{Name: "clicktime", Macro: "{clicktime}", Internal: consts.FieldTS},
		{Name: "devicetype", Macro: "{devicetype}", Internal: consts.FieldOS},
		{Name: "ad_id", Macro: "{ad_id}", Internal: consts.FieldAdID},
		{Name: "creative_id", Macro: "{creative_id}", Internal: consts.FieldCreativeID},
	},
}
