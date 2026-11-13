package main

import (
	"fmt"
	"os"

	_ "ad-monitor/internal/packed"

	// 引入 PostgreSQL 驱动
	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	// 引入 Redis 驱动
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"ad-monitor/internal/cmd"
	"ad-monitor/internal/logic/updater"
)

func main() {
	// 供在线更新自检与 docker-entrypoint.sh 使用, 不加载配置
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version":
			fmt.Println(updater.Version)
			return
		case "--newer-than": // 退出码 0: 本程序版本比参数新
			if len(os.Args) > 2 && updater.Compare(updater.Version, os.Args[2]) > 0 {
				return
			}
			os.Exit(1)
		}
	}
	cmd.Main.Run(gctx.GetInitCtx())
}
