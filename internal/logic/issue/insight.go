package issue

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

const (
	rateWindow  = time.Hour       // 速度按最近 1 小时 (且限于当前状态) 计算
	rateMinSpan = 5 * time.Minute // 数据跨度不足则不估算
	timeLayout  = "2006-01-02 15:04:05"
)

// Insight 由变化记录推算的附加数据, 随列表返回。
type Insight struct {
	RatePerHour    *float64 `json:"ratePerHour"`
	EtaAt          *string  `json:"etaAt"`
	TodaySwitches  int      `json:"todaySwitches"`
	TodayDirectSec int64    `json:"todayDirectSec"`
	TodaySec       int64    `json:"todaySec"`
	TodayFund      *int64   `json:"todayFund"`
	TodayUse       *int64   `json:"todayUse"`
	ReportError    string   `json:"reportError"`
}

// point 一条变化记录的精简形式。
type point struct {
	T              time.Time
	CanIssue       bool
	AutoPayoutCnt  *int64
	ManualClaim2Go *int64
	StatusChanged  bool
}

func loadPoints(ctx context.Context, id int, since time.Time) ([]point, *point, error) {
	var rows []*LogItem
	err := g.Model(TableLog).Ctx(ctx).Where("monitor_id", id).WhereGTE("created_at", since).OrderAsc("id").Scan(&rows)
	if err != nil {
		return nil, nil, err
	}
	var before *LogItem
	err = g.Model(TableLog).Ctx(ctx).Where("monitor_id", id).WhereLT("created_at", since).OrderDesc("id").Limit(1).Scan(&before)
	if err != nil {
		return nil, nil, err
	}
	conv := func(l *LogItem) point {
		return point{T: l.CreatedAt.Time, CanIssue: l.CanIssue, AutoPayoutCnt: l.AutoPayoutCnt,
			ManualClaim2Go: l.ManualClaim2Go, StatusChanged: l.StatusChanged}
	}
	pts := make([]point, len(rows))
	for i, l := range rows {
		pts[i] = conv(l)
	}
	if before == nil {
		return pts, nil, nil
	}
	b := conv(before)
	return pts, &b, nil
}

func metricOf(canIssue bool, auto, manual *int64) *int64 {
	if canIssue {
		return auto
	}
	return manual
}

func fmtCST(t time.Time) string { return t.In(cst).Format(timeLayout) }

func computeInsight(ctx context.Context, m *Monitor, now time.Time) *Insight {
	ins := &Insight{}
	if day, err := todayTotals(ctx, m.AppKey, m.ActivityCategory); err == nil && day != nil {
		ins.TodayFund, ins.TodayUse = &day.TotalFund, &day.TotalUse
	}
	if v, ok := reportErrs.Load(reportKey(m.AppKey, TodayCST(), m.ActivityCategory)); ok {
		ins.ReportError = v.(reportErr).err.Error()
	}

	y, mo, d := now.In(cst).Date()
	dayStart := time.Date(y, mo, d, 0, 0, 0, 0, cst)
	since := dayStart
	if w := now.Add(-rateWindow); w.Before(since) {
		since = w
	}
	pts, before, err := loadPoints(ctx, m.Id, since)
	if err != nil {
		g.Log().Warningf(ctx, "issue: insight monitor=%d: %v", m.Id, err)
		return ins
	}

	// 今日: 切换次数 + 直发时长
	var cur *point
	var curStart time.Time
	if before != nil && !before.T.After(dayStart) {
		cur, curStart = before, dayStart
	}
	for i := range pts {
		p := &pts[i]
		if p.T.Before(dayStart) {
			cur, curStart = p, dayStart
			continue
		}
		if p.StatusChanged {
			ins.TodaySwitches++
		}
		if cur == nil {
			cur, curStart = p, p.T
			ins.TodaySec -= int64(p.T.Sub(dayStart).Seconds()) // 今天首条记录之前的时间不计入
			continue
		}
		if cur.CanIssue {
			ins.TodayDirectSec += int64(p.T.Sub(curStart).Seconds())
		}
		cur, curStart = p, p.T
	}
	if cur != nil {
		if cur.CanIssue {
			ins.TodayDirectSec += int64(now.Sub(curStart).Seconds())
		}
		ins.TodaySec += int64(now.Sub(dayStart).Seconds())
	} else {
		ins.TodaySec = 0
	}

	// 速度 / 预计: 当前状态下, 当前指标最近 1 小时的下降速度
	if m.CanIssue == nil {
		return ins
	}
	value := metricOf(*m.CanIssue, m.AutoPayoutCnt, m.ManualClaim2Go)
	if value == nil {
		return ins
	}
	from := now.Add(-rateWindow)
	if m.ChangedAt != nil && m.ChangedAt.Time.After(from) {
		from = m.ChangedAt.Time
	}
	for _, p := range pts {
		if p.T.Before(from) || p.CanIssue != *m.CanIssue {
			continue
		}
		v := metricOf(p.CanIssue, p.AutoPayoutCnt, p.ManualClaim2Go)
		if v == nil {
			continue
		}
		span := now.Sub(p.T)
		if span < rateMinSpan || *v <= *value {
			break
		}
		rate := float64(*v-*value) / span.Hours()
		eta := fmtCST(now.Add(time.Duration(float64(*value) / rate * float64(time.Hour))))
		ins.RatePerHour, ins.EtaAt = &rate, &eta
		break
	}
	return ins
}

// AttachInsights 为列表中每条配置附加 Insight。
func AttachInsights(ctx context.Context, list []*Monitor) {
	now := time.Now()
	for _, m := range list {
		m.Insight = computeInsight(ctx, m, now)
	}
}

type SeriesPoint struct {
	T        string `json:"t"`
	Full     string `json:"full"`
	CanIssue *bool  `json:"canIssue"`
	Auto     *int64 `json:"auto"`
	Manual   *int64 `json:"manual"`
}

type Round struct {
	State bool    `json:"state"`
	Start string  `json:"start"`
	End   *string `json:"end"`
	Sec   int64   `json:"sec"`
}

// Stats 趋势 (按固定间隔重采样, 取该时刻最后一条记录的值) + 状态周期 (新 → 旧)。
func Stats(ctx context.Context, id, hours int) ([]SeriesPoint, []Round, error) {
	if _, err := get(ctx, id); err != nil {
		return nil, nil, err
	}
	now := time.Now()
	step, label := 10*time.Minute, "15:04"
	if hours > 24 {
		step, label = time.Hour, "01-02 15:04"
	}
	start := now.Add(-time.Duration(hours) * time.Hour).Truncate(step)
	pts, before, err := loadPoints(ctx, id, start)
	if err != nil {
		return nil, nil, err
	}
	all := pts
	if before != nil {
		all = append([]point{*before}, pts...)
	}

	series := []SeriesPoint{}
	j := -1
	for t := start; !t.After(now); t = t.Add(step) {
		for j+1 < len(all) && !all[j+1].T.After(t) {
			j++
		}
		sp := SeriesPoint{T: t.In(cst).Format(label), Full: fmtCST(t)}
		if j >= 0 {
			p := all[j]
			ci := p.CanIssue
			sp.CanIssue, sp.Auto, sp.Manual = &ci, p.AutoPayoutCnt, p.ManualClaim2Go
		}
		series = append(series, sp)
	}

	rounds := []Round{}
	for i, p := range all {
		if i > 0 && all[i-1].CanIssue == p.CanIssue {
			continue
		}
		s := p.T
		if s.Before(start) {
			s = start
		}
		rounds = append(rounds, Round{State: p.CanIssue, Start: fmtCST(s)})
	}
	for i := range rounds {
		st, _ := time.ParseInLocation(timeLayout, rounds[i].Start, cst)
		end := now
		if i+1 < len(rounds) {
			e := rounds[i+1].Start
			rounds[i].End = &e
			end, _ = time.ParseInLocation(timeLayout, e, cst)
		}
		rounds[i].Sec = int64(end.Sub(st).Seconds())
	}
	for l, r := 0, len(rounds)-1; l < r; l, r = l+1, r-1 {
		rounds[l], rounds[r] = rounds[r], rounds[l]
	}
	return series, rounds, nil
}
