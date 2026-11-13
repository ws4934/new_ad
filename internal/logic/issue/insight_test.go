package issue

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// 构造: 今天 00:00 前为手动领取; 今天 now-50m 起手动领取 100 → now-10m 降到 60, 当前 60;
// 今天 now-3h ~ now-2h 有一段直发。
func TestInsightAndStats(t *testing.T) {
	link := os.Getenv("ISSUE_TEST_DB")
	if link == "" {
		t.Skip("ISSUE_TEST_DB not set")
	}
	if err := gdb.SetConfig(gdb.Config{"default": gdb.ConfigGroup{{Link: link}}}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	EnsureSchema(ctx)
	_, _ = g.DB().Exec(ctx, "TRUNCATE issue_monitor, issue_monitor_log RESTART IDENTITY CASCADE")

	now := time.Now()
	y, mo, d := now.In(cst).Date()
	dayStart := time.Date(y, mo, d, 0, 0, 0, 0, cst)
	if now.Sub(dayStart) < 4*time.Hour {
		t.Skip("needs ≥4h elapsed today (Beijing) for deterministic windows")
	}
	f, tr := false, true
	id, err := g.Model(TableMonitor).Ctx(ctx).Data(g.Map{
		"name": "t", "app_key": "k", "app_secret": "s", "activity_category": 1,
		"can_issue": f, "manual_claim_2_go": 60, "checked_at": gtime.New(now),
		"changed_at": gtime.New(now.Add(-2 * time.Hour)),
	}).InsertAndGetId()
	if err != nil {
		t.Fatal(err)
	}
	add := func(at time.Time, ci *bool, manual any, auto any, changed bool) {
		_, err := g.Model(TableLog).Ctx(ctx).Data(g.Map{
			"monitor_id": id, "can_issue": *ci, "manual_claim_2_go": manual, "auto_payout_cnt": auto,
			"status_changed": changed, "created_at": gtime.New(at),
		}).Insert()
		if err != nil {
			t.Fatal(err)
		}
	}
	add(dayStart.Add(-time.Hour), &f, 500, nil, false)
	add(now.Add(-3*time.Hour), &tr, nil, 20, true)
	add(now.Add(-2*time.Hour), &f, 200, nil, true)
	add(now.Add(-50*time.Minute), &f, 100, nil, false)
	add(now.Add(-10*time.Minute), &f, 60, nil, false)

	list, err := List(ctx)
	if err != nil || len(list) != 1 {
		t.Fatal(err, len(list))
	}
	ins := computeInsight(ctx, list[0], now)
	if ins.TodaySwitches != 2 {
		t.Errorf("switches=%d want 2", ins.TodaySwitches)
	}
	if ins.TodayDirectSec < 3590 || ins.TodayDirectSec > 3610 {
		t.Errorf("directSec=%d want ~3600", ins.TodayDirectSec)
	}
	if want := int64(now.Sub(dayStart).Seconds()); ins.TodaySec < want-5 || ins.TodaySec > want+5 {
		t.Errorf("todaySec=%d want ~%d", ins.TodaySec, want)
	}
	// 速度: 窗口 = 最近 1 小时 → 首点 now-50m 的 100, 当前 60 → 40 / (50/60h) = 48/h; 预计 60/48h = 75 分钟后
	if ins.RatePerHour == nil || *ins.RatePerHour < 47.9 || *ins.RatePerHour > 48.1 {
		t.Fatalf("rate=%v want 48", ins.RatePerHour)
	}
	eta, _ := time.ParseInLocation(timeLayout, *ins.EtaAt, cst)
	if diff := eta.Sub(now.Add(75 * time.Minute)); diff < -2*time.Second || diff > 2*time.Second {
		t.Errorf("eta=%s off by %v", *ins.EtaAt, diff)
	}

	series, rounds, err := Stats(ctx, int(id), 6)
	if err != nil {
		t.Fatal(err)
	}
	if len(series) < 36 || len(series) > 38 {
		t.Errorf("series len=%d want ~37", len(series))
	}
	last := series[len(series)-1]
	if last.Manual == nil || *last.Manual != 60 || *last.CanIssue {
		t.Errorf("last point %+v", last)
	}
	// 新 → 旧: 手动(进行中) / 直发 1h / 手动(窗口起点起)
	if len(rounds) != 3 || rounds[0].State || rounds[0].End != nil || !rounds[1].State {
		t.Fatalf("rounds %+v", rounds)
	}
	if rounds[1].Sec < 3590 || rounds[1].Sec > 3610 {
		t.Errorf("direct round sec=%d", rounds[1].Sec)
	}
}
