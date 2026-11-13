package stats

import (
	"context"
	"time"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/dao"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const maxReportRangeDays = 92 // 报表查询最大区间

// ReportData 报表查询结果
type ReportData struct {
	Rows    []*ReportRow  `json:"rows"`
	Summary *ReportTotals `json:"summary"`
}

// ReportRow 单行: 日期 × 账户
type ReportRow struct {
	StatDate     string `json:"statDate"`
	AccountId    int    `json:"accountId"`
	AccountName  string `json:"accountName"`
	PlatformId   int    `json:"platformId"`
	PlatformName string `json:"platformName"`
	IsToday      bool   `json:"isToday"` // 今日实时数据 (未归档, 来自 Redis)
	ReportTotals
}

// ReportTotals 8 个计数器
type ReportTotals struct {
	ClickCount  int `json:"clickCount"`
	TaolianOk   int `json:"taolianOk"`
	TaolianErr  int `json:"taolianErr"`
	CallbackOk  int `json:"callbackOk"`
	CallbackErr int `json:"callbackErr"`
	FakeOk      int `json:"fakeOk"`
	FakeErr     int `json:"fakeErr"`
	DropCount   int `json:"dropCount"`
}

// IsEmpty 全部计数为 0
func (t *ReportTotals) IsEmpty() bool {
	return *t == ReportTotals{}
}

// Add 逐字段累加
func (t *ReportTotals) Add(o *ReportTotals) {
	t.ClickCount += o.ClickCount
	t.TaolianOk += o.TaolianOk
	t.TaolianErr += o.TaolianErr
	t.CallbackOk += o.CallbackOk
	t.CallbackErr += o.CallbackErr
	t.FakeOk += o.FakeOk
	t.FakeErr += o.FakeErr
	t.DropCount += o.DropCount
}

// QueryReport 历史统计报表
// 数据源: report_stats 表 (每日 00:05 归档的历史数据) + 今日实时 Redis 计数器
// (今天尚未归档, 若查询区间包含今天则从 Redis 现读合并进结果)
func QueryReport(ctx context.Context, startDate, endDate string, accountID, platformID int) (*ReportData, error) {
	now := time.Now()
	// 默认区间: 最近 7 天 (含今天)
	if startDate == "" {
		startDate = now.AddDate(0, 0, -6).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = now.Format("2006-01-02")
	}
	start, err := time.ParseInLocation("2006-01-02", startDate, time.Local)
	if err != nil {
		return nil, gerror.Newf("startDate 格式错误: %s", startDate)
	}
	end, err := time.ParseInLocation("2006-01-02", endDate, time.Local)
	if err != nil {
		return nil, gerror.Newf("endDate 格式错误: %s", endDate)
	}
	if end.Before(start) {
		return nil, gerror.New("endDate 不能早于 startDate")
	}
	if end.Sub(start) > maxReportRangeDays*24*time.Hour {
		return nil, gerror.Newf("查询区间最大 %d 天", maxReportRangeDays)
	}

	// 账户/平台名称映射 (报表是低频查询, 直接查库)
	var accts []*entity.AdAccount
	if err := dao.AdAccount.Ctx(ctx).Scan(&accts); err != nil {
		return nil, err
	}
	var plats []*entity.AdPlatform
	if err := dao.AdPlatform.Ctx(ctx).Scan(&plats); err != nil {
		return nil, err
	}
	acctByID := make(map[int]*entity.AdAccount, len(accts))
	for _, a := range accts {
		acctByID[a.Id] = a
	}
	platName := make(map[int]string, len(plats))
	for _, p := range plats {
		platName[p.Id] = p.DisplayName
	}

	res := &ReportData{Rows: []*ReportRow{}, Summary: &ReportTotals{}}

	// ── 1. 今日实时 (区间包含今天时): Redis MGet 批量读 8 个计数器 ──
	todayStr := now.Format("2006-01-02")
	if startDate <= todayStr && todayStr <= endDate {
		rows, err := todayLiveRows(ctx, accts, platName, accountID, platformID)
		if err != nil {
			g.Log().Warningf(ctx, "report: read today live counters failed: %v", err)
		} else {
			res.Rows = append(res.Rows, rows...)
		}
	}

	// ── 2. 历史归档: report_stats ──
	rc := dao.ReportStats.Columns()
	m := dao.ReportStats.Ctx(ctx).
		WhereGTE(rc.StatDate, startDate).
		WhereLTE(rc.StatDate, endDate)
	if accountID > 0 {
		m = m.Where(rc.AccountId, accountID)
	}
	if platformID > 0 {
		m = m.Where(rc.PlatformId, platformID)
	}
	var hist []*entity.ReportStats
	if err := m.OrderDesc(rc.StatDate).OrderAsc(rc.AccountId).Scan(&hist); err != nil {
		return nil, err
	}
	for _, h := range hist {
		row := &ReportRow{
			StatDate:     h.StatDate.Format("Y-m-d"),
			AccountId:    h.AccountId,
			PlatformId:   h.PlatformId,
			PlatformName: platName[h.PlatformId],
			ReportTotals: ReportTotals{
				ClickCount:  h.ClickCount,
				TaolianOk:   h.TaolianOk,
				TaolianErr:  h.TaolianErr,
				CallbackOk:  h.CallbackOk,
				CallbackErr: h.CallbackErr,
				FakeOk:      h.FakeOk,
				FakeErr:     h.FakeErr,
				DropCount:   h.DropCount,
			},
		}
		if a, ok := acctByID[h.AccountId]; ok {
			row.AccountName = a.Name
		}
		res.Rows = append(res.Rows, row)
	}

	// ── 3. 汇总 ──
	for _, r := range res.Rows {
		res.Summary.Add(&r.ReportTotals)
	}
	return res, nil
}

// todayLiveRows 从 Redis 读今天的实时计数器, 按账户生成行 (全 0 的账户跳过)
func todayLiveRows(ctx context.Context, accts []*entity.AdAccount, platName map[int]string, accountID, platformID int) ([]*ReportRow, error) {
	day := Today()
	todayStr := time.Now().Format("2006-01-02")

	// 过滤账户
	filtered := make([]*entity.AdAccount, 0, len(accts))
	for _, a := range accts {
		if accountID > 0 && a.Id != accountID {
			continue
		}
		if platformID > 0 && a.PlatformId != platformID {
			continue
		}
		filtered = append(filtered, a)
	}
	if len(filtered) == 0 {
		return nil, nil
	}

	// 每账户 8 个 key, 一次 MGet
	keys := make([]string, 0, len(filtered)*8)
	for _, a := range filtered {
		keys = append(keys,
			consts.KeyStatsClick(a.Id, day),
			consts.KeyStatsTaolianOk(a.Id, day),
			consts.KeyStatsTaolianErr(a.Id, day),
			consts.KeyStatsCallbackOk(a.Id, day),
			consts.KeyStatsCallbackErr(a.Id, day),
			consts.KeyStatsFakeOk(a.Id, day),
			consts.KeyStatsFakeErr(a.Id, day),
			consts.KeyStatsDrop(a.Id, day),
		)
	}
	counters, err := g.Redis().MGet(ctx, keys...)
	if err != nil {
		return nil, err
	}

	get := func(k string) int {
		if v, ok := counters[k]; ok && v != nil {
			return v.Int()
		}
		return 0
	}

	rows := make([]*ReportRow, 0, len(filtered))
	for i, a := range filtered {
		base := i * 8
		t := ReportTotals{
			ClickCount:  get(keys[base]),
			TaolianOk:   get(keys[base+1]),
			TaolianErr:  get(keys[base+2]),
			CallbackOk:  get(keys[base+3]),
			CallbackErr: get(keys[base+4]),
			FakeOk:      get(keys[base+5]),
			FakeErr:     get(keys[base+6]),
			DropCount:   get(keys[base+7]),
		}
		// 全 0 跳过, 和归档逻辑一致
		if t.IsEmpty() {
			continue
		}
		rows = append(rows, &ReportRow{
			StatDate:     todayStr,
			AccountId:    a.Id,
			AccountName:  a.Name,
			PlatformId:   a.PlatformId,
			PlatformName: platName[a.PlatformId],
			IsToday:      true,
			ReportTotals: t,
		})
	}
	return rows, nil
}
