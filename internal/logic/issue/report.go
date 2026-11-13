package issue

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
	_ "time/tzdata" // 裸机部署缺 zoneinfo 时也能加载 Asia/Shanghai

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
	reportTodayTTL = 5 * time.Minute
	reportPastTTL  = time.Hour
	reportErrTTL   = time.Minute
)

type ReportRow struct {
	Pid     string `json:"pid"`
	FundNum int64  `json:"fundNum"`
	UseNum  int64  `json:"useNum"`
}

// Report 某 AppKey 某天某类红包的发放/使用数 (pid 维度)。
type Report struct {
	Date      string      `json:"date"`
	Category  int         `json:"category"`
	TotalFund int64       `json:"totalFund"`
	TotalUse  int64       `json:"totalUse"`
	List      []ReportRow `json:"list"`
}

type reportEntry struct {
	at  time.Time
	rep *Report
	err error
}

var reportCache sync.Map // appKey|date|category → *reportEntry

// TodayCST 北京时间今天 (YYYYMMDD)。
func TodayCST() string { return time.Now().In(cst).Format("20060102") }

func reportKey(appKey, date string, category int) string {
	return fmt.Sprintf("%s|%s|%d", appKey, date, category)
}

// cachedReport 只读缓存, 不发请求。
func cachedReport(appKey, date string, category int) *reportEntry {
	if v, ok := reportCache.Load(reportKey(appKey, date, category)); ok {
		return v.(*reportEntry)
	}
	return nil
}

// GetReport 带缓存查询 (今天 5 分钟, 历史 1 小时, 失败 1 分钟)。
func GetReport(appKey, secret, date string, category int) (*Report, error) {
	if e := cachedReport(appKey, date, category); e != nil {
		ttl := reportPastTTL
		if e.err != nil {
			ttl = reportErrTTL
		} else if date == TodayCST() {
			ttl = reportTodayTTL
		}
		if time.Since(e.at) < ttl {
			return e.rep, e.err
		}
	}
	rep, err := fetchReport(appKey, secret, date, category)
	reportCache.Store(reportKey(appKey, date, category), &reportEntry{at: time.Now(), rep: rep, err: err})
	return rep, err
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
	return GetReport(appKey, secret, date, category)
}

type TrendPoint struct {
	Date string `json:"date"`
	Fund int64  `json:"fund"`
	Use  int64  `json:"use"`
}

// ReportTrend 最近 days 天 (含今天) 的发放/使用合计, 旧 → 新。
func ReportTrend(ctx context.Context, appKey string, category, days int) ([]TrendPoint, error) {
	secret, err := secretOf(ctx, appKey)
	if err != nil {
		return nil, err
	}
	today := time.Now().In(cst)
	out := make([]TrendPoint, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := today.AddDate(0, 0, -i).Format("20060102")
		rep, err := GetReport(appKey, secret, d, category)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", d, err)
		}
		out = append(out, TrendPoint{Date: d, Fund: rep.TotalFund, Use: rep.TotalUse})
	}
	return out, nil
}

// refreshTodayReports 由 Tick 调用: 为启用的配置刷新今日报表缓存 (过期才真正请求)。
func refreshTodayReports(list []*Monitor) {
	today := TodayCST()
	for _, m := range list {
		if m.Enabled {
			_, _ = GetReport(m.AppKey, m.AppSecret, today, m.ActivityCategory)
		}
	}
}
