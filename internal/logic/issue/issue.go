// Package issue 淘宝客超级红包媒体直发状态监控 (taobao.tbk.dg.rta.sr.issue.match)。
package issue

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"topsdk"
	"topsdk/defaultability"
	"topsdk/defaultability/request"
)

const (
	gatewayURL     = "https://eco.taobao.com/router/rest"
	minIntervalSec = 10
	logKeepDays    = 30
)

// ActivityNames 活动序号 → 名称
var ActivityNames = map[int]string{1: "超级红包", 2: "福利购", 4: "福利直降"}

// Monitor 一条直发查询配置 + 最近一次结果。
type Monitor struct {
	Id               int         `json:"id"               orm:"id"`
	Name             string      `json:"name"             orm:"name"`
	AppKey           string      `json:"appKey"           orm:"app_key"`
	AppSecret        string      `json:"-"                orm:"app_secret"`
	ActivityCategory int         `json:"activityCategory" orm:"activity_category"`
	IntervalSec      int         `json:"intervalSec"      orm:"interval_sec"`
	NotifyEmails     string      `json:"notifyEmails"     orm:"notify_emails"`
	Enabled          bool        `json:"enabled"          orm:"enabled"`
	CanIssue         *bool       `json:"canIssue"         orm:"can_issue"`
	AutoPayoutCnt    *int64      `json:"autoPayoutCnt"    orm:"auto_payout_cnt"`
	ManualClaim2Go   *int64      `json:"manualClaim2Go"   orm:"manual_claim_2_go"`
	ExtraMsg         string      `json:"extraMsg"         orm:"extra_msg"`
	LastError        string      `json:"lastError"        orm:"last_error"`
	CheckedAt        *gtime.Time `json:"checkedAt"        orm:"checked_at"`
	ChangedAt        *gtime.Time `json:"changedAt"        orm:"changed_at"`
	CreatedAt        *gtime.Time `json:"createdAt"        orm:"created_at"`
	Insight          *Insight    `json:"insight"          orm:"-"`
}

// LogItem 数值/状态变化记录。
type LogItem struct {
	Id             int64       `json:"id"             orm:"id"`
	MonitorId      int         `json:"monitorId"      orm:"monitor_id"`
	CanIssue       bool        `json:"canIssue"       orm:"can_issue"`
	AutoPayoutCnt  *int64      `json:"autoPayoutCnt"  orm:"auto_payout_cnt"`
	ManualClaim2Go *int64      `json:"manualClaim2Go" orm:"manual_claim_2_go"`
	StatusChanged  bool        `json:"statusChanged"  orm:"status_changed"`
	MailResult     string      `json:"mailResult"     orm:"mail_result"`
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"`
}

// Input 创建/更新入参。AppSecret 为空表示更新时保持不变。
type Input struct {
	Name             string
	AppKey           string
	AppSecret        string
	ActivityCategory int
	IntervalSec      int
	NotifyEmails     string
	Enabled          bool
}

// Result 一次接口查询结果。
type Result struct {
	CanIssue       bool
	AutoPayoutCnt  *int64
	ManualClaim2Go *int64
	ExtraMsg       string
}

// 每个 monitor 一把锁: 定时轮询与手动刷新互斥, 避免同一状态变化重复发邮件。
var locks sync.Map

func lockOf(id int) *sync.Mutex {
	m, _ := locks.LoadOrStore(id, &sync.Mutex{})
	return m.(*sync.Mutex)
}

func validate(in *Input, creating bool) error {
	in.Name = strings.TrimSpace(in.Name)
	in.AppKey = strings.TrimSpace(in.AppKey)
	in.AppSecret = strings.TrimSpace(in.AppSecret)
	in.NotifyEmails = normalizeEmails(in.NotifyEmails)
	if in.Name == "" || in.AppKey == "" {
		return gerror.New("名称和 AppKey 必填")
	}
	if creating && in.AppSecret == "" {
		return gerror.New("AppSecret 必填")
	}
	if _, ok := ActivityNames[in.ActivityCategory]; !ok {
		return gerror.New("活动类型只能是 1/2/4")
	}
	if in.IntervalSec < minIntervalSec {
		in.IntervalSec = minIntervalSec
	}
	return nil
}

// normalizeEmails 支持逗号/分号/空白/换行分隔, 统一为逗号分隔。
func normalizeEmails(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ';' || r == '，' || r == '；' || r == ' ' || r == '\n' || r == '\t' || r == '\r'
	})
	return strings.Join(parts, ",")
}

// List 全部配置 (按 id 升序)。
func List(ctx context.Context) ([]*Monitor, error) {
	var list []*Monitor
	err := g.Model(TableMonitor).Ctx(ctx).OrderAsc("id").Scan(&list)
	return list, err
}

func get(ctx context.Context, id int) (*Monitor, error) {
	var m *Monitor
	if err := g.Model(TableMonitor).Ctx(ctx).WherePri(id).Scan(&m); err != nil {
		return nil, err
	}
	if m == nil {
		return nil, gerror.Newf("配置 %d 不存在", id)
	}
	return m, nil
}

// Create 新建配置并立即查询一次 (查询失败不影响创建)。
func Create(ctx context.Context, in Input) (int, error) {
	// AppSecret 留空: 复用同 AppKey 已保存的密钥, 方便同一账号快速添加不同活动类型
	if strings.TrimSpace(in.AppSecret) == "" && strings.TrimSpace(in.AppKey) != "" {
		v, err := g.Model(TableMonitor).Ctx(ctx).Where("app_key", strings.TrimSpace(in.AppKey)).
			OrderDesc("id").Value("app_secret")
		if err != nil {
			return 0, err
		}
		in.AppSecret = v.String()
	}
	if err := validate(&in, true); err != nil {
		return 0, err
	}
	id, err := g.Model(TableMonitor).Ctx(ctx).Data(g.Map{
		"name":              in.Name,
		"app_key":           in.AppKey,
		"app_secret":        in.AppSecret,
		"activity_category": in.ActivityCategory,
		"interval_sec":      in.IntervalSec,
		"notify_emails":     in.NotifyEmails,
		"enabled":           in.Enabled,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	_, _ = CheckNow(ctx, int(id))
	return int(id), nil
}

// Update 更新配置。appKey/secret/活动类型变化时清空历史状态, 避免把不同数据源的变化当成状态翻转发邮件。
func Update(ctx context.Context, id int, in Input) error {
	if err := validate(&in, false); err != nil {
		return err
	}
	old, err := get(ctx, id)
	if err != nil {
		return err
	}
	data := g.Map{
		"name":              in.Name,
		"app_key":           in.AppKey,
		"activity_category": in.ActivityCategory,
		"interval_sec":      in.IntervalSec,
		"notify_emails":     in.NotifyEmails,
		"enabled":           in.Enabled,
		"updated_at":        gtime.Now(),
	}
	if in.AppSecret != "" {
		data["app_secret"] = in.AppSecret
	}
	sourceChanged := old.AppKey != in.AppKey || old.ActivityCategory != in.ActivityCategory ||
		(in.AppSecret != "" && in.AppSecret != old.AppSecret)
	if sourceChanged {
		data["can_issue"] = nil
		data["auto_payout_cnt"] = nil
		data["manual_claim_2_go"] = nil
		data["extra_msg"] = ""
		data["last_error"] = ""
		data["checked_at"] = nil
		data["changed_at"] = nil
	}
	mu := lockOf(id)
	mu.Lock()
	_, err = g.Model(TableMonitor).Ctx(ctx).WherePri(id).Data(data).Update()
	mu.Unlock()
	if err == nil && sourceChanged {
		_, _ = CheckNow(ctx, id)
	}
	return err
}

// Delete 删除配置 (日志级联删除)。
func Delete(ctx context.Context, id int) error {
	_, err := g.Model(TableMonitor).Ctx(ctx).WherePri(id).Delete()
	locks.Delete(id)
	return err
}

// Logs 最近 limit 条变化记录。
func Logs(ctx context.Context, id, limit int) ([]*LogItem, error) {
	var list []*LogItem
	err := g.Model(TableLog).Ctx(ctx).Where("monitor_id", id).OrderDesc("id").Limit(limit).Scan(&list)
	return list, err
}

// queryFn 测试替换点。
var queryFn = Query

// Query 调用淘宝接口查询一次。
func Query(appKey, appSecret string, activity int) (res *Result, err error) {
	// SDK 的 HandleJsonResponse 对短响应体直接切片 jsonStr[0:20], 会 panic
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("解析淘宝接口响应失败: %v", r)
		}
	}()
	client := topsdk.NewDefaultTopClient(appKey, appSecret, gatewayURL, 15000, 15000)
	req := request.TaobaoTbkDgRtaSrIssueMatchRequest{}
	req.SetActivityCategory(int64(activity))
	resp, err := defaultability.NewDefaultability(&client).TaobaoTbkDgRtaSrIssueMatch(&req)
	if err != nil {
		return nil, err
	}
	d := resp.Data
	if d.CanIssue == nil {
		return nil, fmt.Errorf("响应缺少 can_issue: %s", resp.Body)
	}
	res = &Result{CanIssue: *d.CanIssue, AutoPayoutCnt: d.AutoPayoutCnt, ManualClaim2Go: d.ManualClaim2Go}
	if d.ExtraMsg != nil {
		res.ExtraMsg = *d.ExtraMsg
	}
	return res, nil
}

// CheckNow 立即查询并落库; 状态翻转时记录日志并发邮件。
func CheckNow(ctx context.Context, id int) (*Monitor, error) {
	mu := lockOf(id)
	mu.Lock()
	defer mu.Unlock()

	m, err := get(ctx, id)
	if err != nil {
		return nil, err
	}
	now := gtime.Now()
	res, qerr := queryFn(m.AppKey, m.AppSecret, m.ActivityCategory)
	if qerr != nil {
		_, err = g.Model(TableMonitor).Ctx(ctx).WherePri(id).Data(g.Map{
			"last_error": qerr.Error(), "checked_at": now,
		}).Update()
		if err != nil {
			return nil, err
		}
		return get(ctx, id)
	}

	statusChanged := m.CanIssue != nil && *m.CanIssue != res.CanIssue
	valuesChanged := m.CanIssue == nil || statusChanged ||
		!eqPtr(m.AutoPayoutCnt, res.AutoPayoutCnt) || !eqPtr(m.ManualClaim2Go, res.ManualClaim2Go)

	data := g.Map{
		"can_issue":         res.CanIssue,
		"auto_payout_cnt":   res.AutoPayoutCnt,
		"manual_claim_2_go": res.ManualClaim2Go,
		"extra_msg":         res.ExtraMsg,
		"last_error":        "",
		"checked_at":        now,
	}
	if statusChanged {
		data["changed_at"] = now
	}
	if _, err = g.Model(TableMonitor).Ctx(ctx).WherePri(id).Data(data).Update(); err != nil {
		return nil, err
	}

	if valuesChanged {
		mailResult := ""
		if statusChanged {
			mailResult = notifyStatusChange(ctx, m, res)
		}
		_, err = g.Model(TableLog).Ctx(ctx).Data(g.Map{
			"monitor_id":        id,
			"can_issue":         res.CanIssue,
			"auto_payout_cnt":   res.AutoPayoutCnt,
			"manual_claim_2_go": res.ManualClaim2Go,
			"status_changed":    statusChanged,
			"mail_result":       mailResult,
		}).Insert()
		if err != nil {
			g.Log().Warningf(ctx, "issue: write log monitor=%d: %v", id, err)
		}
	}
	return get(ctx, id)
}

func eqPtr(a, b *int64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// notifyStatusChange 发状态变化邮件, 返回写入日志的结果描述。
func notifyStatusChange(ctx context.Context, m *Monitor, res *Result) string {
	if m.NotifyEmails == "" {
		return "未配置收件人"
	}
	subject, body := buildMail(m, res)
	if err := SendMail(ctx, strings.Split(m.NotifyEmails, ","), subject, body); err != nil {
		g.Log().Warningf(ctx, "issue: send mail monitor=%d: %v", m.Id, err)
		return "发送失败: " + err.Error()
	}
	return "已发送"
}

// StatusText 状态文案。
func StatusText(canIssue bool) string {
	if canIssue {
		return "直发中"
	}
	return "手动领取"
}

func countText(v *int64, none string) string {
	if v == nil {
		return none
	}
	return fmt.Sprintf("%d 个", *v)
}

func buildMail(m *Monitor, res *Result) (string, string) {
	from := "未知"
	if m.CanIssue != nil {
		from = StatusText(*m.CanIssue)
	}
	to := StatusText(res.CanIssue)
	subject := fmt.Sprintf("[直发提醒] %s: %s → %s", m.Name, from, to)

	var b strings.Builder
	fmt.Fprintf(&b, "配置: %s (AppKey %s, %s)\n", m.Name, m.AppKey, ActivityNames[m.ActivityCategory])
	fmt.Fprintf(&b, "状态: %s → %s\n", from, to)
	fmt.Fprintf(&b, "时间: %s\n\n", gtime.Now().String())
	if res.CanIssue {
		fmt.Fprintf(&b, "剩余可直发红包: %s\n", countText(res.AutoPayoutCnt, "超过展示门槛"))
		fmt.Fprintf(&b, "距下轮直发目标还需发放: %s\n", countText(res.ManualClaim2Go, "超过展示门槛"))
	} else {
		fmt.Fprintf(&b, "还需手抽红包达标: %s\n", countText(res.ManualClaim2Go, "超过展示门槛"))
	}
	if res.ExtraMsg != "" {
		fmt.Fprintf(&b, "提示: %s\n", res.ExtraMsg)
	}
	return subject, b.String()
}

// Tick 定时任务: 查询所有到期的启用配置, 并清理过期日志。
func Tick(ctx context.Context) {
	list, err := List(ctx)
	if err != nil {
		g.Log().Warningf(ctx, "issue: list monitors: %v", err)
		return
	}
	now := time.Now()
	for _, m := range list {
		if !m.Enabled {
			continue
		}
		if m.CheckedAt != nil && now.Sub(m.CheckedAt.Time) < time.Duration(m.IntervalSec)*time.Second {
			continue
		}
		if _, err := CheckNow(ctx, m.Id); err != nil {
			g.Log().Warningf(ctx, "issue: check monitor=%d: %v", m.Id, err)
		}
	}
	refreshTodayReports(list)
	_, _ = g.Model(TableLog).Ctx(ctx).
		WhereLT("created_at", gtime.Now().AddDate(0, 0, -logKeepDays)).Delete()
}
