package issue

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
	_ "time/tzdata" // 裸机部署缺 zoneinfo 时也能加载 Asia/Shanghai

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"topsdk"
	"topsdk/ability2474"
	req2474 "topsdk/ability2474/request"
)

// cst 淘宝报表按北京时间日统计
var cst, _ = time.LoadLocation("Asia/Shanghai")

// ReportCategories 红包发放报表支持的类型 (比直发接口多 3/5/6)
var ReportCategories = map[int]string{1: "超级红包", 2: "福利购", 3: "签到红包", 4: "福利直降", 5: "幸运赢免单", 6: "超级金币"}

const (
	reportPageSize = 100
	reportTodayTTL = 5 * time.Minute // 今天: 5 分钟刷新
	reportYdayTTL  = time.Hour       // 昨天: 数据仍可能补录, 1 小时刷新
	reportErrTTL   = time.Minute     // 失败后 1 分钟内不重试
	backfillDays   = 31              // 后台补齐最近 31 天
	backfillPerRun = 2               // 每次 Tick 每个 (AppKey, 类型) 最多补 2 天, 避免集中打接口
)

type ReportRow struct {
	Pid     string `json:"pid"     orm:"pid"`
	FundNum int64  `json:"fundNum" orm:"fund_num"`
	UseNum  int64  `json:"useNum"  orm:"use_num"`
}

// Report 某 AppKey 某天某类红包的发放/使用数 (pid 维度)。
type Report struct {
	Date      string      `json:"date"`
	Category  int         `json:"category"`
	TotalFund int64       `json:"totalFund"`
	TotalUse  int64       `json:"totalUse"`
	FetchedAt string      `json:"fetchedAt"`
	List      []ReportRow `json:"list"`
}

type dayRow struct {
	BizDate   string    `orm:"biz_date"`
	TotalFund int64     `orm:"total_fund"`
	TotalUse  int64     `orm:"total_use"`
	FetchedAt time.Time `orm:"fetched_at"`
}

var reportErrs sync.Map // appKey|date|category → time of last failure + error

type reportErr struct {
	at  time.Time
	err error
}

// TodayCST 北京时间今天 (YYYYMMDD)。
func TodayCST() string { return time.Now().In(cst).Format("20060102") }

func reportKey(appKey, date string, category int) string {
	return fmt.Sprintf("%s|%s|%d", appKey, date, category)
}

// fresh 入库数据是否无需刷新。date 在两天后 0 点之后抓取的数据视为定稿, 永不再请求接口。
func fresh(date string, fetchedAt, now time.Time) bool {
	d, err := time.ParseInLocation("20060102", date, cst)
	if err != nil {
		return true
	}
	final := d.AddDate(0, 0, 2)
	switch {
	case !fetchedAt.Before(final):
		return true
	case !now.Before(final):
		return false // 定稿前抓的旧数据, 补抓一次
	case now.Before(d.AddDate(0, 0, 1)):
		return now.Sub(fetchedAt) < reportTodayTTL
	default:
		return now.Sub(fetchedAt) < reportYdayTTL
	}
}

func loadDays(ctx context.Context, appKey string, category int, from, to string) (map[string]*dayRow, error) {
	var rows []*dayRow
	err := g.Model(TableRptDay).Ctx(ctx).
		Where("app_key", appKey).Where("category", category).WhereBetween("biz_date", from, to).Scan(&rows)
	m := make(map[string]*dayRow, len(rows))
	for _, r := range rows {
		m[r.BizDate] = r
	}
	return m, err
}

func loadReport(ctx context.Context, appKey, date string, category int, day *dayRow) (*Report, error) {
	rep := &Report{Date: date, Category: category, TotalFund: day.TotalFund, TotalUse: day.TotalUse,
		FetchedAt: fmtCST(day.FetchedAt), List: []ReportRow{}}
	err := g.Model(TableRptPid).Ctx(ctx).Fields("pid, fund_num, use_num").
		Where("app_key", appKey).Where("category", category).Where("biz_date", date).
		OrderDesc("fund_num").Scan(&rep.List)
	return rep, err
}

func saveReport(ctx context.Context, appKey string, rep *Report, at time.Time) error {
	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, err := tx.Model(TableRptDay).Ctx(ctx).Data(g.Map{
			"app_key": appKey, "biz_date": rep.Date, "category": rep.Category,
			"total_fund": rep.TotalFund, "total_use": rep.TotalUse, "fetched_at": at,
		}).OnConflict("app_key", "category", "biz_date").Save()
		if err != nil {
			return err
		}
		if _, err = tx.Model(TableRptPid).Ctx(ctx).Where("app_key", appKey).
			Where("category", rep.Category).Where("biz_date", rep.Date).Delete(); err != nil {
			return err
		}
		if len(rep.List) == 0 {
			return nil
		}
		rows := make(g.List, 0, len(rep.List))
		for _, r := range rep.List {
			rows = append(rows, g.Map{"app_key": appKey, "biz_date": rep.Date, "category": rep.Category,
				"pid": r.Pid, "fund_num": r.FundNum, "use_num": r.UseNum})
		}
		_, err = tx.Model(TableRptPid).Ctx(ctx).Data(rows).Insert()
		return err
	})
}

// GetReport 优先读库; 未入库或需要刷新时请求接口并入库。接口失败时有旧数据则返回旧数据。
func GetReport(ctx context.Context, appKey, secret, date string, category int) (*Report, error) {
	days, err := loadDays(ctx, appKey, category, date, date)
	if err != nil {
		return nil, err
	}
	return getReport(ctx, appKey, secret, date, category, days[date])
}

func getReport(ctx context.Context, appKey, secret, date string, category int, day *dayRow) (*Report, error) {
	now := time.Now()
	if day != nil && fresh(date, day.FetchedAt, now) {
		return loadReport(ctx, appKey, date, category, day)
	}
	key := reportKey(appKey, date, category)
	if v, ok := reportErrs.Load(key); ok && now.Sub(v.(reportErr).at) < reportErrTTL {
		if day != nil {
			return loadReport(ctx, appKey, date, category, day)
		}
		return nil, v.(reportErr).err
	}
	rep, err := fetchReport(appKey, secret, date, category)
	if err != nil {
		reportErrs.Store(key, reportErr{at: now, err: err})
		if day != nil {
			return loadReport(ctx, appKey, date, category, day)
		}
		return nil, err
	}
	reportErrs.Delete(key)
	if err := saveReport(ctx, appKey, rep, now); err != nil {
		g.Log().Warningf(ctx, "issue: save report %s: %v", key, err)
	}
	rep.FetchedAt = fmtCST(now)
	return rep, nil
}

func fetchReport(appKey, secret, date string, category int) (rep *Report, err error) {
	defer func() { // SDK 对短响应体切片会 panic
		if r := recover(); r != nil {
			err = fmt.Errorf("解析淘宝接口响应失败: %v", r)
		}
	}()
	client := topsdk.NewDefaultTopClient(appKey, secret, gatewayURL, 15000, 15000)
	ab := ability2474.NewAbility2474(&client)
	rep = &Report{Date: date, Category: category, List: []ReportRow{}}
	for page := int64(1); page <= 50; page++ {
		req := req2474.TaobaoTbkDgVegasSendReportRequest{}
		req.SetBizDate(date)
		req.SetActivityCategory(int64(category))
		req.SetPageNo(page)
		req.SetPageSize(reportPageSize)
		resp, err := ab.TaobaoTbkDgVegasSendReport(&req)
		if err != nil {
			return nil, err
		}
		r := resp.Result
		if r.Success != nil && !*r.Success {
			msg := "查询失败"
			if r.MsgInfo != nil {
				msg = *r.MsgInfo
			}
			return nil, gerror.New(msg)
		}
		if r.Model == nil || r.Model.PidRptList == nil {
			break
		}
		rows := *r.Model.PidRptList
		for _, x := range rows {
			row := ReportRow{}
			if x.Pid != nil {
				row.Pid = *x.Pid
			}
			if x.FundNum != nil {
				row.FundNum = *x.FundNum
			}
			if x.UseNum != nil {
				row.UseNum = *x.UseNum
			}
			rep.TotalFund += row.FundNum
			rep.TotalUse += row.UseNum
			rep.List = append(rep.List, row)
		}
		if len(rows) < reportPageSize {
			break
		}
	}
	sort.Slice(rep.List, func(i, j int) bool { return rep.List[i].FundNum > rep.List[j].FundNum })
	return rep, nil
}

// secretOf 取某 AppKey 已保存的密钥 (最新一条配置)。
func secretOf(ctx context.Context, appKey string) (string, error) {
	v, err := g.Model(TableMonitor).Ctx(ctx).Where("app_key", appKey).OrderDesc("id").Value("app_secret")
	if err != nil {
		return "", err
	}
	if v.IsEmpty() {
		return "", gerror.Newf("AppKey %s 没有已保存的配置", appKey)
	}
	return v.String(), nil
}

// ReportFor 报表页: 按 AppKey 查某天的发放明细, date 为空表示今天。
func ReportFor(ctx context.Context, appKey, date string, category int) (*Report, error) {
	secret, err := secretOf(ctx, appKey)
	if err != nil {
		return nil, err
	}
	if date == "" {
		date = TodayCST()
	}
	return GetReport(ctx, appKey, secret, date, category)
}

type TrendPoint struct {
	Date string `json:"date"`
	Fund int64  `json:"fund"`
	Use  int64  `json:"use"`
}

func lastDays(days int) []string {
	today := time.Now().In(cst)
	out := make([]string, 0, days)
	for i := days - 1; i >= 0; i-- {
		out = append(out, today.AddDate(0, 0, -i).Format("20060102"))
	}
	return out
}

// ReportTrend 最近 days 天 (含今天) 的发放/使用合计, 旧 → 新。已入库且无需刷新的日期直接读库。
func ReportTrend(ctx context.Context, appKey string, category, days int) ([]TrendPoint, error) {
	secret, err := secretOf(ctx, appKey)
	if err != nil {
		return nil, err
	}
	dates := lastDays(days)
	stored, err := loadDays(ctx, appKey, category, dates[0], dates[len(dates)-1])
	if err != nil {
		return nil, err
	}
	now := time.Now()
	out := make([]TrendPoint, 0, days)
	for _, d := range dates {
		if day := stored[d]; day != nil && fresh(d, day.FetchedAt, now) {
			out = append(out, TrendPoint{Date: d, Fund: day.TotalFund, Use: day.TotalUse})
			continue
		}
		rep, err := getReport(ctx, appKey, secret, d, category, stored[d])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", d, err)
		}
		out = append(out, TrendPoint{Date: d, Fund: rep.TotalFund, Use: rep.TotalUse})
	}
	return out, nil
}

// todayTotals 卡片用: 只读库里今天的合计, 不发请求。
func todayTotals(ctx context.Context, appKey string, category int) (*dayRow, error) {
	today := TodayCST()
	days, err := loadDays(ctx, appKey, category, today, today)
	return days[today], err
}

// refreshReports 由 Tick 调用: 刷新今天的报表, 并逐步补齐最近 31 天未入库/未定稿的日期。
func refreshReports(ctx context.Context, list []*Monitor) {
	seen := map[string]bool{}
	dates := lastDays(backfillDays)
	now := time.Now()
	for _, m := range list {
		k := reportKey(m.AppKey, "", m.ActivityCategory)
		if !m.Enabled || seen[k] {
			continue
		}
		seen[k] = true
		stored, err := loadDays(ctx, m.AppKey, m.ActivityCategory, dates[0], dates[len(dates)-1])
		if err != nil {
			continue
		}
		fetched := 0
		for i := len(dates) - 1; i >= 0 && fetched < backfillPerRun; i-- { // 新 → 旧
			d := dates[i]
			if day := stored[d]; day != nil && fresh(d, day.FetchedAt, now) {
				continue
			}
			if v, ok := reportErrs.Load(reportKey(m.AppKey, d, m.ActivityCategory)); ok && now.Sub(v.(reportErr).at) < reportErrTTL {
				continue
			}
			_, _ = getReport(ctx, m.AppKey, m.AppSecret, d, m.ActivityCategory, stored[d])
			fetched++
		}
	}
}
