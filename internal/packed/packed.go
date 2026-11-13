package packed

import (
	// 注册 PostgreSQL 驱动到 GoFrame ORM
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	// 注册 Redis 驱动到 GoFrame gredis
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	// 平台适配器 blank import
	// 各 adapter 包在 init() 里调用 adapter.Register 自注册, 必须在这里 blank import
	// 触发 init, 否则 registry 是空的, adapter.Get("weibo") 会返回 nil。
	//
	// 新增平台时, 在这里加一行 blank import 即可。
	_ "ad-monitor/internal/logic/adapter/weibo"
	_ "ad-monitor/internal/logic/adapter/xiaohongshu"
)
