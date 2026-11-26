package adapter

import "ad-monitor/internal/consts"

// MacroField 单个字段的元数据
type MacroField struct {
	Name     string               // 平台接口的字段名 (如 "idfa_md5" / "IMP")
	Macro    string               // 平台宏占位符 (如 "{idfa_MD5}" / "__IDFA_MD5__")
	Internal consts.InternalField // 映射到的内部标准字段
	Required bool                 // 是否必传
}

// MacroSpec 平台宏元数据声明
type MacroSpec struct {
	PlatformName string
	Fields       []MacroField
	NoTaolian    bool // 监测链接不带设备号 (淘联点击上报至少要一个): 点击不上报淘联, 只走虚假上报/联调
}

// ClickData 解析后的标准化点击数据
type ClickData struct {
	AccountID int
	Platform  string
	Values    map[consts.InternalField]string // 只包含成功替换 (非宏占位符/非空) 的字段
	RawQuery  string                          // 原始 query string (排障用)
}
