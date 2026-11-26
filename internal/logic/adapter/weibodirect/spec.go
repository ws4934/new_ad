package weibodirect

import (
	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/adapter"
)

// Spec 微博直达监测链接宏元数据 (__XXX__ 格式)
// 对接文档: docs/platform-docs/微博直达资产联调手册_副本.md 3.1
// 直达只替换 MARKID/TS/IP 三个宏 (已确认不支持其他宏), 没有设备号淘联无法归因,
// 所以不上报淘联, 转化只来自虚假上报 (联调模式照常模拟回传)
var Spec = adapter.MacroSpec{
	PlatformName: "weibo_direct",
	NoTaolian:    true,
	Fields: []adapter.MacroField{
		{Name: "markid", Macro: "__MARKID__", Internal: consts.FieldClickID, Required: true},
		{Name: "ip", Macro: "__IP__", Internal: consts.FieldIP},
		{Name: "clicktime", Macro: "__TS__", Internal: consts.FieldTS},
	},
}
