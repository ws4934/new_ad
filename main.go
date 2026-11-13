package main

import (
	_ "ad-monitor/internal/packed"

	// 引入 PostgreSQL 驱动
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	// 引入 Redis 驱动
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"ad-monitor/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
