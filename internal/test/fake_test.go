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
	"ad-monitor/internal/model/bo"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hibiken/asynq"
)

// TestIntegration_FakeReport_SkipsTaolian 虚假上报抽中的点击只进虚假回传队列, 不上报淘联
func TestIntegration_FakeReport_SkipsTaolian(t *testing.T) {
	s := setupTestServer(t)
	defer s.Shutdown()
	ctx := context.Background()
	aid := getTestAccountID(ctx, t)

	ac := dao.AdAccount.Columns()
	setFake := func(on bool) {
		_, _ = dao.AdAccount.Ctx(ctx).WherePri(aid).Data(g.Map{ac.FakeEnabled: on, ac.FakeRate: 100}).Update()
		cache.InvalidateAccount(ctx, aid)
	}
	setFake(true)
	defer setFake(false)

	ins := asynq.NewInspector(asynq.RedisClientOpt{
		Addr: g.Cfg().MustGet(ctx, "adMonitor.queue.redisAddr").String(),
		DB:   g.Cfg().MustGet(ctx, "adMonitor.queue.redisDb").Int(),
	})
	count := func(tasks []*asynq.TaskInfo, cid string) int {
		n := 0
		for _, ti := range tasks {
			var p bo.TaolianTaskPayload // 与 FakeTaskPayload 的 aid/cid 字段同名
			if json.Unmarshal(ti.Payload, &p) == nil && p.AccountID == aid && p.ClickID == cid {
				n++
			}
		}
		return n
	}

	// 点击ID每次运行唯一: 测试不起 asynq 消费者, 上次运行留下的任务不能影响本次计数
	cid := "FAKE_ONLY_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/m/%d?IMP=%s&imei_md5=fake_dev", s.GetListenedPort(), aid, cid))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	time.Sleep(300 * time.Millisecond)

	pending, _ := ins.ListPendingTasks(consts.QueueDefault, asynq.PageSize(1000))
	scheduled, _ := ins.ListScheduledTasks(consts.QueueFake, asynq.PageSize(1000))
	if tl, fk := count(pending, cid), count(scheduled, cid); tl != 0 || fk != 1 {
		t.Fatalf("taolian tasks=%d fake tasks=%d, want 0/1", tl, fk)
	}
}
