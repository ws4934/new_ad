package queue

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/hibiken/asynq"
)

// ShutdownPlugin 实现 ghttp.Plugin, 服务关闭时优雅停止 asynq
type ShutdownPlugin struct {
	Server    *asynq.Server
	Scheduler *asynq.Scheduler
}

var _ ghttp.Plugin = (*ShutdownPlugin)(nil)

func (p *ShutdownPlugin) Name() string                    { return "asynq-shutdown" }
func (p *ShutdownPlugin) Author() string                  { return "" }
func (p *ShutdownPlugin) Version() string                 { return "1.0.0" }
func (p *ShutdownPlugin) Description() string             { return "优雅退出时关停 asynq server 和 scheduler" }
func (p *ShutdownPlugin) Install(_ *ghttp.Server) error   { return nil }

// Remove 在 server shutdown 时由框架调用
func (p *ShutdownPlugin) Remove() error {
	ctx := context.Background()
	if p.Server != nil {
		g.Log().Infof(ctx, "shutdown: stopping asynq server...")
		p.Server.Shutdown()
	}
	if p.Scheduler != nil {
		g.Log().Infof(ctx, "shutdown: stopping asynq scheduler...")
		p.Scheduler.Shutdown()
	}
	g.Log().Infof(ctx, "shutdown: asynq cleanup done")
	return nil
}
