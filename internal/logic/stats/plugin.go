package stats

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ShutdownPlugin 实现 ghttp.Plugin 接口
// 服务关闭时 flush 内存中累积的业务计数器, 保证正常重启不丢计数
type ShutdownPlugin struct{}

var _ ghttp.Plugin = (*ShutdownPlugin)(nil)

func (p *ShutdownPlugin) Name() string                   { return "stats-shutdown" }
func (p *ShutdownPlugin) Author() string                 { return "" }
func (p *ShutdownPlugin) Version() string                { return "1.0.0" }
func (p *ShutdownPlugin) Description() string            { return "优雅退出时 flush 业务计数器到 Redis" }
func (p *ShutdownPlugin) Install(_ *ghttp.Server) error { return nil }

// Remove 在 server shutdown 时由框架调用
func (p *ShutdownPlugin) Remove() error {
	ctx := context.Background()
	g.Log().Infof(ctx, "shutdown: flushing stats counters...")
	FlushCounters(ctx)
	return nil
}
