package cmd

import (
	"context"

	authctl "ad-monitor/internal/controller/auth"
	"ad-monitor/internal/controller/account"
	"ad-monitor/internal/controller/callback"
	"ad-monitor/internal/controller/mapping"
	"ad-monitor/internal/controller/monitor"
	"ad-monitor/internal/controller/platform"
	statsctl "ad-monitor/internal/controller/stats"
	"ad-monitor/internal/logic/auth"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/logic/dedup"
	"ad-monitor/internal/logic/queue"
	"ad-monitor/internal/logic/stats"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gcron"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "广告监测上报平台",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 0. 安全前置校验: jwtSecret/adminPassword 均未配置时拒绝启动
			auth.EnsureSecretConfigured()

			// 1. 缓存预热
			cache.WarmupAll(ctx)

			// 2. 启动 asynq 消费者 + 调度器 + QPS 打点 / 业务计数 worker
			asynqSrv := queue.StartServer(ctx)
			asynqSch := queue.StartScheduler(ctx)
			stats.StartClickWorker(ctx)
			stats.StartCounterWorker(ctx)

			// 3. 初始化 CSV writer + dedup
			csvPath := g.Cfg().MustGet(ctx, "adMonitor.csvExportPath").String()
			csvWriter := dedup.NewCSVWriter(csvPath)
			dedupSvc := dedup.New(csvWriter)

			// 3.1 定时持久化 Bloom Filter (每 5 分钟, 无新增的 filter 跳过)
			// 必须先 flush CSV 再 dump bloom: 若 bloom 先落盘而 CSV 还在缓冲,
			// 硬崩溃后设备在 bloom 中"已见过"但 CSV 行永久丢失 (静默缺数据)。
			// 反向顺序最坏情况只是 bloom 比 CSV 旧 → 重复行, 可下游去重。
			_, _ = gcron.Add(ctx, "0 */5 * * * *", func(ctx context.Context) {
				csvWriter.Flush()
				dedupSvc.DumpAll(ctx)
			}, "bloom-dump")

			// 3.2 定时刷 CSV 写缓冲 (每 30 秒; 进程退出由 ShutdownPlugin 兜底)
			_, _ = gcron.Add(ctx, "*/30 * * * * *", func(ctx context.Context) {
				csvWriter.Flush()
			}, "csv-flush")

			// 4. HTTP Server
			s := g.Server()

			// 4.1 注册 Plugin: 服务关闭时优雅停止各组件
			s.Plugin(&queue.ShutdownPlugin{Server: asynqSrv, Scheduler: asynqSch})
			s.Plugin(&dedup.ShutdownPlugin{Dedup: dedupSvc, CSV: csvWriter})
			s.Plugin(&stats.ShutdownPlugin{})

			// 外部接收路由 (无鉴权)
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Bind(monitor.New(dedupSvc))
				group.Bind(callback.New(dedupSvc))
			})

			// 登录接口 (无鉴权, 用 MiddlewareHandlerResponse 包装响应)
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(authctl.New())
			})

			// 后台管理 API (JWT 鉴权 + 响应包装)
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(auth.MiddlewareJWT)
				group.Middleware(ghttp.MiddlewareHandlerResponse)
				group.Bind(platform.New())
				group.Bind(account.New())
				group.Bind(mapping.New())
				group.Bind(statsctl.New(csvWriter))
			})

			s.Run()
			return nil
		},
	}
)
