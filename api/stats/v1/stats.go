package v1

import (
	logicstats "ad-monitor/internal/logic/stats"

	"github.com/gogf/gf/v2/frame/g"
)

// --- 总览 ---

type OverviewReq struct {
	g.Meta `path:"/api/v2/stats/overview" tags:"Stats" method:"get" summary:"今日统计总览"`
}

type OverviewRes struct {
	TotalClick    int `json:"totalClick"`
	TotalTaolian  int `json:"totalTaolian"`
	TotalCallback int `json:"totalCallback"`
	TotalFake     int `json:"totalFake"`
}

// --- QPS ---

type ResetTodayReq struct {
	g.Meta `path:"/api/v2/stats/reset-today" tags:"Stats" method:"post" summary:"清零所有账户今日统计"`
}

type ResetTodayRes struct {
	Deleted int `json:"deleted"`
}

type QpsReq struct {
	g.Meta `path:"/api/v2/stats/qps" tags:"Stats" method:"get" summary:"全局 QPS"`
}

type QpsRes struct {
	*logicstats.QPSMetric
}

// --- 历史报表 ---

type ReportReq struct {
	g.Meta     `path:"/api/v2/stats/report" tags:"Stats" method:"get" summary:"历史统计报表"`
	StartDate  string `json:"startDate"`  // YYYY-MM-DD, 默认 6 天前
	EndDate    string `json:"endDate"`    // YYYY-MM-DD, 默认今天
	AccountID  int    `json:"accountId"`  // 0 = 全部账户
	PlatformID int    `json:"platformId"` // 0 = 全部平台
}

// 行/汇总结构定义在 logicstats (api → logic 单向依赖, 同 QpsRes)
type ReportRes struct {
	*logicstats.ReportData
}

// --- CSV 文件列表 (带大小) ---

type ExportFilesReq struct {
	g.Meta   `path:"/api/v2/stats/export-files" tags:"Stats" method:"get" summary:"CSV 文件列表"`
	Platform string `json:"platform"`
	Field    string `json:"field"`
}

type ExportFilesRes struct {
	Files []*ExportFileItem `json:"files"`
}

type ExportFileItem struct {
	Name string `json:"name"`
	Size int64  `json:"size"` // 字节
}

// --- CSV 文件下载 ---

type ExportDownloadReq struct {
	g.Meta   `path:"/api/v2/stats/export-download" tags:"Stats" method:"get" summary:"CSV 文件下载"`
	Filename string `json:"filename" v:"required"`
}

type ExportDownloadRes struct {
	g.Meta `mime:"application/octet-stream"`
}

// --- CSV 文件删除 ---

type ExportDeleteReq struct {
	g.Meta   `path:"/api/v2/stats/export-delete" tags:"Stats" method:"post" summary:"删除 CSV 文件"`
	Filename string `json:"filename" v:"required"`
}

type ExportDeleteRes struct{}
