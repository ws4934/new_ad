package test_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/stats"
	"ad-monitor/internal/model/bo"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hibiken/asynq"
)

// TestIntegration_ClickDedup 同一点击ID的重复请求: 只计「重复点击」, 不计点击, 不上报淘联; 请求记录照常保留
func TestIntegration_ClickDedup(t *testing.T) {
	s := setupTestServer(t)
	defer s.Shutdown()
	ctx := context.Background()
	aid := getTestAccountID(ctx, t)
	day := stats.Today()
	stats.FlushCounters(ctx)
	keys := []string{consts.KeyStatsClick(aid, day), consts.KeyStatsDup(aid, day), consts.KeyStatsClickSet(aid, day), consts.KeyReqLog(aid)}
	g.Redis().Del(ctx, keys...)
	defer g.Redis().Del(ctx, keys...)

	ins := asynq.NewInspector(asynq.RedisClientOpt{
		Addr: g.Cfg().MustGet(ctx, "adMonitor.queue.redisAddr").String(),
		DB:   g.Cfg().MustGet(ctx, "adMonitor.queue.redisDb").Int(),
	})
	taolianTasks := func(cid string) int {
		n := 0
		tasks, _ := ins.ListPendingTasks(consts.QueueDefault, asynq.PageSize(1000))
		for _, ti := range tasks {
			var p bo.TaolianTaskPayload
			if json.Unmarshal(ti.Payload, &p) == nil && p.AccountID == aid && p.ClickID == cid {
				n++
			}
		}
		return n
	}
	hit := func(imp string) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/m/%d?IMP=%s&imei_md5=dedup_dev", s.GetListenedPort(), aid, imp))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}
	hit("DEDUP_X")
	hit("DEDUP_X") // 重复
	hit("DEDUP_Y")
	time.Sleep(300 * time.Millisecond)
	stats.FlushCounters(ctx)
	stats.FlushRequestLog(ctx)

	click, _ := g.Redis().Get(ctx, keys[0])
	dup, _ := g.Redis().Get(ctx, keys[1])
	logs, _ := g.Redis().LLen(ctx, keys[3])
	if click.Int() != 2 || dup.Int() != 1 || logs != 3 {
		t.Fatalf("click=%d dup=%d reqlog=%d, want 2/1/3", click.Int(), dup.Int(), logs)
	}
	if n := taolianTasks("DEDUP_X"); n != 1 {
		t.Fatalf("taolian tasks for DEDUP_X = %d, want 1 (duplicate must not be reported)", n)
	}

	// 清零今日统计后去重集合也清掉, 同一个点击ID重新计数
	if _, err := stats.ResetToday(ctx); err != nil {
		t.Fatal(err)
	}
	hit("DEDUP_X")
	time.Sleep(200 * time.Millisecond)
	stats.FlushCounters(ctx)
	if v, _ := g.Redis().Get(ctx, keys[0]); v.Int() != 1 {
		t.Fatalf("click after reset = %d, want 1", v.Int())
	}
}
