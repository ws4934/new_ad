package adapter

import (
	"fmt"
	"regexp"
	"strings"

	"ad-monitor/internal/consts"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

// macroPattern 同时覆盖 {xxx} 和 __XXX__ 两种宏格式
var macroPattern = regexp.MustCompile(`^(\{[a-zA-Z_0-9]+\}|__[A-Z_0-9]+__)$`)

// IsMacroValue 判断值是否仍然是未替换的宏占位符或空
func IsMacroValue(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return true
	}
	return macroPattern.MatchString(v)
}

// BuildMonitorURL 生成监测链接模板
// 输出: {baseURL}/m/{accountId}?{name1}={macro1}&{name2}={macro2}&...
func BuildMonitorURL(baseURL string, accountID int, spec MacroSpec) string {
	parts := make([]string, 0, len(spec.Fields))
	for _, f := range spec.Fields {
		parts = append(parts, f.Name+"="+f.Macro)
	}
	return fmt.Sprintf("%s/m/%d?%s", baseURL, accountID, strings.Join(parts, "&"))
}

// ParseClickRequest 从请求解析为标准化字段
// 自动过滤值为空或仍是宏占位符的字段
// 必填字段缺失时返回 error (由 controller 决定是否返回 204)
func ParseClickRequest(r *ghttp.Request, spec MacroSpec) (*ClickData, error) {
	data := &ClickData{
		Platform: spec.PlatformName,
		Values:   map[consts.InternalField]string{},
		RawQuery: r.URL.RawQuery,
	}
	for _, f := range spec.Fields {
		raw := strings.TrimSpace(r.Get(f.Name).String())
		if raw == "" || IsMacroValue(raw) {
			if f.Required {
				return nil, gerror.Newf("required field %s not replaced (got %q)", f.Name, raw)
			}
			continue
		}
		data.Values[f.Internal] = raw
	}
	return data, nil
}

// ExtractDeviceIDs 从 ClickData 提取设备号 (按内部字段名)
func ExtractDeviceIDs(data *ClickData) map[consts.InternalField]string {
	out := map[consts.InternalField]string{}
	for _, f := range consts.DeviceFields {
		if v, ok := data.Values[f]; ok && v != "" {
			out[f] = v
		}
	}
	return out
}
