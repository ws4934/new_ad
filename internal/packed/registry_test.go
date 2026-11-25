package packed_test

import (
	"testing"

	"ad-monitor/internal/logic/adapter"
	// 触发 packed init: pg/redis 驱动 + 平台 adapter 自注册
	_ "ad-monitor/internal/packed"
)

// TestAdaptersRegistered 验证 blank import 触发各 adapter 包的 init(), 注册进 registry
// 这是 packed.go 里 `_ "ad-monitor/internal/logic/adapter/weibo"` 等 blank import 的
// 行为真值验证 — 保证 adapter.Get("weibo") 不会返回 nil。
func TestAdaptersRegistered(t *testing.T) {
	cases := []string{"weibo", "weibo_direct", "xiaohongshu"}
	for _, name := range cases {
		a := adapter.Get(name)
		if a == nil {
			t.Errorf("adapter %q not registered (check packed.go blank imports)", name)
			continue
		}
		if a.Name() != name {
			t.Errorf("adapter %q: Name() = %q, want %q", name, a.Name(), name)
		}
		spec := a.Spec()
		if spec.PlatformName != name {
			t.Errorf("adapter %q: Spec().PlatformName = %q, want %q", name, spec.PlatformName, name)
		}
		if len(spec.Fields) == 0 {
			t.Errorf("adapter %q: Spec().Fields is empty", name)
		}
	}
}

// TestAdapterRegistry_All 验证 adapter.All() 返回完整列表
func TestAdapterRegistry_All(t *testing.T) {
	all := adapter.All()
	if len(all) < 2 {
		t.Errorf("expected >= 2 adapters registered, got %d", len(all))
	}
	if _, ok := all["weibo"]; !ok {
		t.Error("weibo not in All()")
	}
	if _, ok := all["xiaohongshu"]; !ok {
		t.Error("xiaohongshu not in All()")
	}
}
