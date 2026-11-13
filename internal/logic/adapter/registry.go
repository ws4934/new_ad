package adapter

import "sync"

var (
	registry   = map[string]Adapter{}
	registryMu sync.RWMutex
)

// Register 注册平台适配器 (在各 adapter 包的 init() 中调用)
func Register(a Adapter) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[a.Name()] = a
}

// Get 按平台名获取适配器
func Get(name string) Adapter {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[name]
}

// All 获取所有已注册适配器
func All() map[string]Adapter {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make(map[string]Adapter, len(registry))
	for k, v := range registry {
		out[k] = v
	}
	return out
}
