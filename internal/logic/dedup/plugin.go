package dedup

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// ShutdownPlugin 实现 ghttp.Plugin 接口
// 服务关闭时持久化 Bloom Filter + 关闭 CSV 文件句柄
type ShutdownPlugin struct {
	Dedup *Dedup
	CSV   *CSVWriter
}

var _ ghttp.Plugin = (*ShutdownPlugin)(nil)

func (p *ShutdownPlugin) Name() string        { return "dedup-shutdown" }
func (p *ShutdownPlugin) Author() string       { return "" }
func (p *ShutdownPlugin) Version() string      { return "1.0.0" }
func (p *ShutdownPlugin) Description() string  { return "优雅退出时持久化 Bloom Filter 并关闭 CSV 文件" }
func (p *ShutdownPlugin) Install(_ *ghttp.Server) error { return nil }

// Remove 在 server shutdown 时由框架调用
// 顺序: CSV flush → bloom dump → CSV close。
// CSV 必须先于 bloom 落盘: 若 bloom 先持久化而进程在 CSV flush 前被强杀,
// 设备会永久缺失于 CSV (bloom 认为已见过, 不再重发)。
func (p *ShutdownPlugin) Remove() error {
	ctx := context.Background()
	if p.CSV != nil {
		g.Log().Infof(ctx, "shutdown: flushing csv writers...")
		p.CSV.Flush()
	}
	if p.Dedup != nil {
		g.Log().Infof(ctx, "shutdown: dumping bloom filters...")
		p.Dedup.DumpAll(ctx)
	}
	if p.CSV != nil {
		g.Log().Infof(ctx, "shutdown: closing csv writers...")
		p.CSV.Close()
	}
	g.Log().Infof(ctx, "shutdown: dedup cleanup done")
	return nil
}
