package issue

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func i64(v int64) *int64 { return &v }

func TestNormalizeEmails(t *testing.T) {
	got := normalizeEmails(" a@x.com，b@x.com;\nc@x.com  ")
	if got != "a@x.com,b@x.com,c@x.com" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildMail(t *testing.T) {
	f := false
	m := &Monitor{Name: "主号", AppKey: "34070427", ActivityCategory: 1, CanIssue: &f}
	subj, body := buildMail(m, &Result{CanIssue: true, AutoPayoutCnt: i64(10), ExtraMsg: "当前红包剩余10个"})
	if subj != "[直发提醒] 主号: 手动领取 → 直发中" {
		t.Fatalf("subject %q", subj)
	}
	for _, want := range []string{"剩余直发红包: 10 个", "距下轮直发目标还需发放: 已超展示门槛（存在指定直发）"} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "当前红包剩余") {
		t.Fatalf("extra_msg should not be shown:\n%s", body)
	}
	tr := true
	m.CanIssue = &tr
	if _, body = buildMail(m, &Result{CanIssue: false, ManualClaim2Go: i64(5)}); !strings.Contains(body, "还需手抽红包 (完成后切换回直发): 5 个") {
		t.Fatalf("manual body:\n%s", body)
	}
	if _, body = buildMail(m, &Result{CanIssue: false}); !strings.Contains(body, "还需手抽红包 (完成后切换回直发): 未达展示门槛") {
		t.Fatalf("manual none body:\n%s", body)
	}
}

// 集成测试: ISSUE_TEST_DB=pgsql:postgres:test@tcp(127.0.0.1:55432)/ad_monitor
// 可选 ISSUE_TEST_SMTP=127.0.0.1:51025 (mailpit) + ISSUE_TEST_MAILPIT=http://127.0.0.1:58025
func TestMonitorFlow(t *testing.T) {
	link := os.Getenv("ISSUE_TEST_DB")
	if link == "" {
		t.Skip("ISSUE_TEST_DB not set")
	}
	if err := gdb.SetConfig(gdb.Config{"default": gdb.ConfigGroup{{Link: link}}}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	EnsureSchema(ctx)
	EnsureSchema(ctx) // 幂等
	_, _ = g.DB().Exec(ctx, "TRUNCATE issue_monitor, issue_monitor_log, app_setting RESTART IDENTITY CASCADE")

	if smtpAddr := os.Getenv("ISSUE_TEST_SMTP"); smtpAddr != "" {
		host, port, _ := strings.Cut(smtpAddr, ":")
		p := 0
		_ = json.Unmarshal([]byte(port), &p)
		if err := SaveSMTP(ctx, SMTPConfig{Host: host, Port: p, Username: "u", Password: "p", From: "直发提醒 <bot@test.local>"}); err != nil {
			t.Fatal(err)
		}
		// 空密码保存应保留原密码
		if err := SaveSMTP(ctx, SMTPConfig{Host: host, Port: p, Username: "u", From: "直发提醒 <bot@test.local>"}); err != nil {
			t.Fatal(err)
		}
		if cfg, _ := GetSMTP(ctx); cfg.Password != "p" {
			t.Fatalf("password not kept: %+v", cfg)
		}
	}

	// 模拟淘宝接口返回序列
	seq := []*Result{
		{CanIssue: false, ManualClaim2Go: i64(5)},
		{CanIssue: false, ManualClaim2Go: i64(5)}, // 无变化 → 不写日志
		{CanIssue: false, ManualClaim2Go: i64(2)}, // 数值变化 → 写日志, 不发邮件
		{CanIssue: true, AutoPayoutCnt: i64(10)},  // 状态翻转 → 写日志 + 发邮件
	}
	calls := 0
	queryFn = func(string, string, int) (*Result, error) {
		if calls >= len(seq) {
			return nil, errors.New("boom")
		}
		r := seq[calls]
		calls++
		return r, nil
	}
	defer func() { queryFn = Query }()

	if _, err := Create(ctx, Input{Name: "x", AppKey: "k", ActivityCategory: 1}); err == nil {
		t.Fatal("create without secret should fail")
	}
	id, err := Create(ctx, Input{Name: "主号", AppKey: "34070427", AppSecret: "s", ActivityCategory: 1,
		IntervalSec: 1, NotifyEmails: "ops@test.local", Enabled: true}) // Create 内部查询第 1 次
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if _, err := CheckNow(ctx, id); err != nil {
			t.Fatal(err)
		}
	}
	m, err := CheckNow(ctx, id) // 第 5 次: 查询失败, 保留上次结果
	if err != nil {
		t.Fatal(err)
	}
	if m.LastError != "boom" || m.CanIssue == nil || !*m.CanIssue || *m.AutoPayoutCnt != 10 || m.ChangedAt == nil {
		t.Fatalf("unexpected monitor %+v", m)
	}
	if m.IntervalSec != minIntervalSec {
		t.Fatalf("interval should be clamped to %d, got %d", minIntervalSec, m.IntervalSec)
	}

	logs, err := Logs(ctx, id, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 3 {
		t.Fatalf("want 3 logs, got %d", len(logs))
	}
	if !logs[0].StatusChanged || logs[1].StatusChanged || logs[2].StatusChanged {
		t.Fatalf("status_changed flags wrong: %+v %+v %+v", logs[0], logs[1], logs[2])
	}
	if os.Getenv("ISSUE_TEST_SMTP") != "" && logs[0].MailResult != "已发送" {
		t.Fatalf("mail result %q", logs[0].MailResult)
	}
	if mp := os.Getenv("ISSUE_TEST_MAILPIT"); mp != "" {
		resp, err := http.Get(mp + "/api/v1/messages")
		if err != nil {
			t.Fatal(err)
		}
		var out struct {
			Messages []struct{ Subject string } `json:"messages"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&out)
		resp.Body.Close()
		if len(out.Messages) != 1 || out.Messages[0].Subject != "[直发提醒] 主号: 手动领取 → 直发中" {
			t.Fatalf("mailpit messages: %+v", out.Messages)
		}
	}

	// 更新 appKey → 清空历史状态, 并立即重新查询 (queryFn 已耗尽 → last_error)
	if err := Update(ctx, id, Input{Name: "主号", AppKey: "other", ActivityCategory: 2, IntervalSec: 30, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	var list []*Monitor
	if list, err = List(ctx); err != nil || len(list) != 1 {
		t.Fatalf("list %v %d", err, len(list))
	}
	if m := list[0]; m.CanIssue != nil || m.AppSecret != "s" || m.ActivityCategory != 2 {
		t.Fatalf("update result %+v", m)
	}

	// AppSecret 留空 → 复用同 AppKey 的密钥; 无同 AppKey 记录 → 报错
	id2, err := Create(ctx, Input{Name: "福利直降", AppKey: "other", ActivityCategory: 4, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if m2, _ := get(ctx, id2); m2.AppSecret != "s" {
		t.Fatalf("secret not reused: %q", m2.AppSecret)
	}
	if _, err := Create(ctx, Input{Name: "x", AppKey: "nobody", ActivityCategory: 1}); err == nil {
		t.Fatal("create without secret for unknown appKey should fail")
	}
	_ = Delete(ctx, id2)

	if err := Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	if logs, _ := Logs(ctx, id, 50); len(logs) != 0 {
		t.Fatal("logs should cascade delete")
	}
}

// 真实网关冒烟: 错误密钥应得到淘宝的签名错误而不是 panic / 解析失败。
func TestQueryRealGateway(t *testing.T) {
	if os.Getenv("ISSUE_TEST_REAL") == "" {
		t.Skip("ISSUE_TEST_REAL not set")
	}
	_, err := Query("34070427", "wrong-secret", 1)
	if err == nil {
		t.Fatal("expected error with wrong secret")
	}
	t.Logf("gateway error: %v", err)
}
