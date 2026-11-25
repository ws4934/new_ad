package weibodirect

import (
	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/adapter"
)

// Spec 微博直达监测链接宏元数据 (__XXX__ 格式)
// 对接文档: docs/platform-docs/微博直达资产联调手册_副本.md 3.1
// 直达只下发 MARKID/TS/IP, 没有设备号 (与微博 IMP 下载类监测不同)
var Spec = adapter.MacroSpec{
	PlatformName: "weibo_direct",
	Fields: []adapter.MacroField{
		{Name: "markid", Macro: "__MARKID__", Internal: consts.FieldClickID, Required: true},
		{Name: "ip", Macro: "__IP__", Internal: consts.FieldIP},
		{Name: "clicktime", Macro: "__TS__", Internal: consts.FieldTS},
	},
}
