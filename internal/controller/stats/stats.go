package stats

import (
	"context"
	"os"
	"path/filepath"

	v1 "ad-monitor/api/stats/v1"
	"ad-monitor/internal/consts"
	"ad-monitor/internal/dao"
	"ad-monitor/internal/logic/dedup"
	logicstats "ad-monitor/internal/logic/stats"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct {
	csvWriter *dedup.CSVWriter
}

func New(csvWriter *dedup.CSVWriter) *Controller {
	return &Controller{csvWriter: csvWriter}
}

// ResetToday 清零所有账户今日统计
func (c *Controller) ResetToday(ctx context.Context, _ *v1.ResetTodayReq) (*v1.ResetTodayRes, error) {
	n, err := logicstats.ResetToday(ctx)
	if err != nil {
		return nil, err
	}
	g.Log().Infof(ctx, "stats: reset today, %d keys deleted", n)
	return &v1.ResetTodayRes{Deleted: n}, nil
}

// Overview 今日统计总览
func (c *Controller) Overview(ctx context.Context, req *v1.OverviewReq) (res *v1.OverviewRes, err error) {
	// 查所有启用账户
	var accounts []*entity.AdAccount
	if err := dao.AdAccount.Ctx(ctx).Where(dao.AdAccount.Columns().Status, 1).Scan(&accounts); err != nil {
		return nil, err
	}

	day := logicstats.Today()
	res = &v1.OverviewRes{}

	if len(accounts) > 0 {
		// 批量 MGet 4 个计数器
		keys := make([]string, 0, len(accounts)*4)
		for _, a := range accounts {
			keys = append(keys,
				consts.KeyStatsClick(a.Id, day),
				consts.KeyStatsTaolianOk(a.Id, day),
				consts.KeyStatsCallbackOk(a.Id, day),
				consts.KeyStatsFakeOk(a.Id, day),
			)
		}
		counters, mErr := g.Redis().MGet(ctx, keys...)
		if mErr != nil {
			g.Log().Warningf(ctx, "overview mget failed: %v", mErr)
		}
		if counters != nil {
			for i := range accounts {
				base := i * 4
				if v, ok := counters[keys[base]]; ok && v != nil {
					res.TotalClick += v.Int()
				}
				if v, ok := counters[keys[base+1]]; ok && v != nil {
					res.TotalTaolian += v.Int()
				}
				if v, ok := counters[keys[base+2]]; ok && v != nil {
					res.TotalCallback += v.Int()
				}
				if v, ok := counters[keys[base+3]]; ok && v != nil {
					res.TotalFake += v.Int()
				}
			}
		}
	}

	// CSV 新增数不再从 Redis bitmap 读 (去重改为本地 bloom filter)
	// 前端 CSV 导出页可看文件大小

	return res, nil
}

// Report 历史统计报表 (report_stats 归档 + 今日实时)
func (c *Controller) Report(ctx context.Context, req *v1.ReportReq) (res *v1.ReportRes, err error) {
	data, err := logicstats.QueryReport(ctx, req.StartDate, req.EndDate, req.AccountID, req.PlatformID)
	if err != nil {
		return nil, err
	}
	return &v1.ReportRes{ReportData: data}, nil
}

// Qps 全局 QPS
func (c *Controller) Qps(ctx context.Context, req *v1.QpsReq) (res *v1.QpsRes, err error) {
	metric, err := logicstats.GetGlobalQPS(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.QpsRes{QPSMetric: metric}, nil
}

// ExportFiles CSV 文件列表 (带大小)
func (c *Controller) ExportFiles(ctx context.Context, req *v1.ExportFilesReq) (res *v1.ExportFilesRes, err error) {
	files, err := c.csvWriter.ListFiles(req.Platform, req.Field, "")
	if err != nil {
		return nil, err
	}
	items := make([]*v1.ExportFileItem, 0, len(files))
	for _, name := range files {
		info, err := os.Stat(filepath.Join(c.csvWriter.BaseDir(), name))
		size := int64(0)
		if err == nil {
			size = info.Size()
		}
		items = append(items, &v1.ExportFileItem{Name: name, Size: size})
	}
	return &v1.ExportFilesRes{Files: items}, nil
}

// ExportDownload CSV 文件下载
func (c *Controller) ExportDownload(ctx context.Context, req *v1.ExportDownloadReq) (res *v1.ExportDownloadRes, err error) {
	safeName := filepath.Base(req.Filename)
	if safeName != req.Filename || safeName == "." || safeName == ".." {
		return nil, gerror.New("invalid filename")
	}
	c.csvWriter.Flush() // 先把写缓冲落盘, 保证下载到完整内容
	fullPath := filepath.Join(c.csvWriter.BaseDir(), safeName)
	r := g.RequestFromCtx(ctx)
	r.Response.ServeFile(fullPath)
	return nil, nil
}

// ExportDelete 删除 CSV 文件
// 走 CSVWriter.RemoveFile: 若删除的是活跃分片, 会先关闭写句柄再删, 避免写进已删除的 inode
func (c *Controller) ExportDelete(ctx context.Context, req *v1.ExportDeleteReq) (res *v1.ExportDeleteRes, err error) {
	safeName := filepath.Base(req.Filename)
	if safeName != req.Filename || safeName == "." || safeName == ".." {
		return nil, gerror.New("invalid filename")
	}
	if err := c.csvWriter.RemoveFile(safeName); err != nil {
		return nil, gerror.Newf("删除失败: %v", err)
	}
	return &v1.ExportDeleteRes{}, nil
}
