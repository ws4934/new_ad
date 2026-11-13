package test_test

import (
	"context"
	"testing"
	"time"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/dao"
	"ad-monitor/internal/logic/stats"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/test/gtest"
)

// TestIntegration_StatsReport 报表查询: 历史归档 + 今日实时合并
func TestIntegration_StatsReport(t *testing.T) {
	ctx := context.Background()
	ensureTestData(ctx, t)
	aid := getTestAccountID(ctx, t)

	var acct entity.AdAccount
	if err := dao.AdAccount.Ctx(ctx).WherePri(aid).Scan(&acct); err != nil {
		t.Fatal(err)
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	// ── 准备数据 ──
	// 1. 历史: 直接写一条昨天的归档记录
	rc := dao.ReportStats.Columns()
	_, err := dao.ReportStats.Ctx(ctx).Data(g.Map{
		rc.AccountId:  aid,
		rc.PlatformId: acct.PlatformId,
		rc.StatDate:   yesterday,
		rc.ClickCount: 100,
		rc.TaolianOk:  90,
		rc.CallbackOk: 10,
	}).OnConflict(rc.AccountId, rc.PlatformId, rc.StatDate).Save()
	gtest.AssertNil(err)
	defer dao.ReportStats.Ctx(ctx).
		Where(rc.AccountId, aid).Where(rc.StatDate, yesterday).Delete()

	// 2. 今日实时: 写 Redis 计数器
	dayKey := stats.Today()
	clickKey := consts.KeyStatsClick(aid, dayKey)
	before, _ := g.Redis().Get(ctx, clickKey)
	_, _ = g.Redis().IncrBy(ctx, clickKey, 7)
	defer func() {
		if before == nil || before.IsNil() {
			g.Redis().Del(ctx, clickKey)
		} else {
			g.Redis().Set(ctx, clickKey, before.Int())
		}
	}()

	// ── 查询: 昨天~今天 ──
	res, err := stats.QueryReport(ctx, yesterday, today, aid, 0)
	gtest.AssertNil(err)

	var histRow, todayRow bool
	for _, r := range res.Rows {
		if r.StatDate == yesterday && r.AccountId == aid {
			histRow = true
			gtest.Assert(r.ClickCount, 100)
			gtest.Assert(r.TaolianOk, 90)
			gtest.Assert(r.IsToday, false)
			gtest.AssertNE(r.PlatformName, "")
		}
		if r.StatDate == today && r.AccountId == aid {
			todayRow = true
			gtest.Assert(r.IsToday, true)
			gtest.AssertGE(r.ClickCount, 7)
		}
	}
	if !histRow {
		t.Error("history row (yesterday) missing")
	}
	if !todayRow {
		t.Error("today live row missing")
	}

	// 汇总应包含两部分
	gtest.AssertGE(res.Summary.ClickCount, 107)

	// ── 边界: 区间不含今天 → 只有历史 ──
	res2, err := stats.QueryReport(ctx, yesterday, yesterday, aid, 0)
	gtest.AssertNil(err)
	for _, r := range res2.Rows {
		gtest.Assert(r.IsToday, false)
	}

	// ── 参数校验 ──
	_, err = stats.QueryReport(ctx, "bad-date", today, 0, 0)
	gtest.AssertNE(err, nil)
	_, err = stats.QueryReport(ctx, today, yesterday, 0, 0)
	gtest.AssertNE(err, nil)
	_, err = stats.QueryReport(ctx, "2020-01-01", today, 0, 0) // 超 92 天
	gtest.AssertNE(err, nil)
}
