package issue

import (
	"testing"
	"time"
)

func TestReportFresh(t *testing.T) {
	now := time.Date(2026, 11, 13, 14, 0, 0, 0, cst)
	cases := []struct {
		date    string
		fetched time.Time
		want    bool
	}{
		{"20261113", now.Add(-time.Minute), true},                      // 今天, 5 分钟内
		{"20261113", now.Add(-6 * time.Minute), false},                 // 今天, 过期
		{"20261112", now.Add(-30 * time.Minute), true},                 // 昨天, 1 小时内
		{"20261112", now.Add(-2 * time.Hour), false},                   // 昨天, 过期
		{"20261110", time.Date(2026, 11, 12, 0, 0, 1, 0, cst), true},   // 定稿后抓取 → 永久有效
		{"20261110", time.Date(2026, 11, 11, 12, 0, 0, 0, cst), false}, // 定稿前抓取 → 补抓一次
		{"20261101", time.Date(2026, 11, 13, 9, 0, 0, 0, cst), true},
	}
	for _, c := range cases {
		if got := fresh(c.date, c.fetched, now); got != c.want {
			t.Errorf("fresh(%s, %s)=%v want %v", c.date, c.fetched.Format(timeLayout), got, c.want)
		}
	}
}

func TestDateRange(t *testing.T) {
	today := TodayCST()
	tomorrow := time.Now().In(cst).AddDate(0, 0, 1).Format("20060102")
	if d, err := dateRange("20261101", "20261103"); err != nil || len(d) != 3 || d[0] != "20261101" || d[2] != "20261103" {
		t.Fatalf("3 days: %v %v", d, err)
	}
	if d, err := dateRange(today, tomorrow); err != nil || len(d) != 1 || d[0] != today {
		t.Fatalf("future end should clamp to today: %v %v", d, err)
	}
	if _, err := dateRange("20261103", "20261101"); err == nil {
		t.Fatal("reversed range should fail")
	}
	if _, err := dateRange("20261001", "20261031"); err != nil {
		t.Fatalf("31 days should pass: %v", err)
	}
	if _, err := dateRange("20261001", "20261101"); err == nil {
		t.Fatal("32 days should fail")
	}
	if _, err := dateRange("2026-11-01", "20261101"); err == nil {
		t.Fatal("bad format should fail")
	}
}
