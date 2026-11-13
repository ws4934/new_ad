package cmd

import (
	"context"
	"net/http"

	authctl "ad-monitor/internal/controller/auth"
	"ad-monitor/internal/controller/account"
	"ad-monitor/internal/controller/callback"
	"ad-monitor/internal/controller/mapping"
	"ad-monitor/internal/controller/monitor"
	issuectl "ad-monitor/internal/controller/issue"
	"ad-monitor/internal/controller/platform"
	systemctl "ad-monitor/internal/controller/system"
	statsctl "ad-monitor/internal/controller/stats"
	"ad-monitor/internal/logic/auth"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/logic/dedup"
	"ad-monitor/internal/logic/issue"
	"ad-monitor/internal/logic/queue"
	"ad-monitor/internal/logic/stats"
	"ad-monitor/internal/logic/updater"
	"ad-monitor/resource/public"

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

			g.Log().Infof(ctx, "ad-monitor %s", updater.Version)

			// 0.1 直发监控建表 (幂等) + 每 5 秒扫描到期配置 (各配置按自身间隔查询)
			issue.EnsureSchema(ctx)
			_, _ = gcron.AddSingleton(ctx, "*/5 * * * * *", issue.Tick, "issue-monitor")

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

			// 前端静态文件 (编译进二进制)
			static := ghttp.WrapH(http.FileServer(http.FS(public.FS)))
			s.BindHandler("GET:/", static)
			s.BindHandler("GET:/js/*", static)

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
				group.Bind(issuectl.New())
				group.Bind(systemctl.New())
			})

			s.Run()
			// 在线更新: 新二进制已就位时 exec 替换当前进程
			return updater.RestartIfPending()
		},
	}
)
