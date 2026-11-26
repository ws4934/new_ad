package test_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/dao"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/logic/stats"
	"ad-monitor/internal/model/bo"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/hibiken/asynq"
)

// queuedTasks 某次点击在淘联队列 (待处理) 和虚假回传队列 (延迟) 里的任务数
func queuedTasks(ctx context.Context, aid int, cid string) (taolian, fake int) {
	ins := asynq.NewInspector(asynq.RedisClientOpt{
		Addr: g.Cfg().MustGet(ctx, "adMonitor.queue.redisAddr").String(),
		DB:   g.Cfg().MustGet(ctx, "adMonitor.queue.redisDb").Int(),
	})
	defer ins.Close()
	count := func(tasks []*asynq.TaskInfo) int {
		n := 0
		for _, ti := range tasks {
			var p bo.TaolianTaskPayload // 与 FakeTaskPayload 的 aid/cid 字段同名
			if json.Unmarshal(ti.Payload, &p) == nil && p.AccountID == aid && p.ClickID == cid {
				n++
			}
		}
		return n
	}
	pending, _ := ins.ListPendingTasks(consts.QueueDefault, asynq.PageSize(1000))
	scheduled, _ := ins.ListScheduledTasks(consts.QueueFake, asynq.PageSize(1000))
	return count(pending), count(scheduled)
}

// uniqueCID 点击ID每次运行唯一: 测试不起 asynq 消费者, 上次运行留下的任务不能影响本次计数
func uniqueCID(prefix string) string {
	return prefix + strconv.FormatInt(time.Now().UnixNano(), 36)
}

func hitMonitor(t *testing.T, s *ghttp.Server, aid int, query string) {
	t.Helper()
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/m/%d?%s", s.GetListenedPort(), aid, query))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	time.Sleep(300 * time.Millisecond)
}

func setFake(ctx context.Context, aid int, on bool) {
	ac := dao.AdAccount.Columns()
	_, _ = dao.AdAccount.Ctx(ctx).WherePri(aid).Data(g.Map{ac.FakeEnabled: on, ac.FakeRate: 100}).Update()
	cache.InvalidateAccount(ctx, aid)
}

// TestIntegration_FakeReport_SkipsTaolian 虚假上报抽中的点击只进虚假回传队列, 不上报淘联
func TestIntegration_FakeReport_SkipsTaolian(t *testing.T) {
	s := setupTestServer(t)
	defer s.Shutdown()
	ctx := context.Background()
	aid := getTestAccountID(ctx, t)
	setFake(ctx, aid, true)
	defer setFake(ctx, aid, false)

	cid := uniqueCID("FAKE_ONLY_")
	hitMonitor(t, s, aid, "IMP="+cid+"&imei_md5=fake_dev")
	if tl, fk := queuedTasks(ctx, aid, cid); tl != 0 || fk != 1 {
		t.Fatalf("taolian tasks=%d fake tasks=%d, want 0/1", tl, fk)
	}
}

// TestIntegration_WeiboDirect_NoTaolian 微博直达没有设备号: 点击从不上报淘联, 开虚假上报时只进虚假回传队列
func TestIntegration_WeiboDirect_NoTaolian(t *testing.T) {
	s := setupTestServer(t)
	defer s.Shutdown()
	ctx := context.Background()

	var plat entity.AdPlatform
	if err := dao.AdPlatform.Ctx(ctx).Where(dao.AdPlatform.Columns().Name, "weibo_direct").Scan(&plat); err != nil {
		t.Fatalf("weibo_direct platform not found: %v", err)
	}
	ac := dao.AdAccount.Columns()
	id, err := dao.AdAccount.Ctx(ctx).Data(g.Map{
		ac.PlatformId:    plat.Id,
		ac.Name:          uniqueCID("it_weibo_direct_"),
		ac.Credentials:   `{"service_provider":"sp"}`,
		ac.TaolianConfig: `{}`,
		ac.FakeTypes:     "[]",
		ac.FakeDelayMin:  30,
		ac.FakeDelayMax:  60,
		ac.Status:        1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatal(err)
	}
	aid := int(id)
	mc := dao.ActionMapping.Columns()
	defer func() {
		dao.ActionMapping.Ctx(ctx).Where(mc.AccountId, aid).Delete()
		dao.AdAccount.Ctx(ctx).WherePri(aid).Delete()
		cache.InvalidateAccount(ctx, aid)
		g.Redis().Del(ctx, consts.KeyStatsClick(aid, stats.Today()), consts.KeyStatsClickSet(aid, stats.Today()), consts.KeyReqLog(aid))
	}()
	if _, err := dao.ActionMapping.Ctx(ctx).Data(g.Map{
		mc.AccountId: aid, mc.TaolianType: 11, mc.TaolianDesc: "淘客订单支付",
		mc.PlatformType: "2001", mc.PlatformDesc: "付款订单数", mc.Enabled: true,
	}).Insert(); err != nil {
		t.Fatal(err)
	}

	cid := uniqueCID("WD_PLAIN_")
	hitMonitor(t, s, aid, "markid="+cid+"&ip=1.2.3.4&clicktime=1700000000000")
	if tl, fk := queuedTasks(ctx, aid, cid); tl != 0 || fk != 0 {
		t.Fatalf("fake off: taolian tasks=%d fake tasks=%d, want 0/0", tl, fk)
	}

	setFake(ctx, aid, true)
	cid = uniqueCID("WD_FAKE_")
	hitMonitor(t, s, aid, "markid="+cid+"&ip=1.2.3.4&clicktime=1700000000000")
	if tl, fk := queuedTasks(ctx, aid, cid); tl != 0 || fk != 1 {
		t.Fatalf("fake on: taolian tasks=%d fake tasks=%d, want 0/1", tl, fk)
	}
	stats.FlushCounters(ctx)
	if v, _ := g.Redis().Get(ctx, consts.KeyStatsClick(aid, stats.Today())); v.Int() != 2 {
		t.Fatalf("clicks=%d, want 2 (both requests accepted)", v.Int())
	}
}
