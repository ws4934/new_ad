package test_test

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"os"
	"strconv"
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
	// 点击ID每次运行唯一: 测试不起 asynq 消费者, 上次运行留下的待处理任务不能影响本次计数
	run := strconv.FormatInt(time.Now().UnixNano(), 36)
	X, Y, Z := "DEDUP_X_"+run, "DEDUP_Y_"+run, "DEDUP_Z_"+run
	hit := func(imp string) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/m/%d?IMP=%s&imei_md5=dedup_dev", s.GetListenedPort(), aid, imp))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}
	hit(X)
	hit(X) // 重复
	hit(Y)
	time.Sleep(300 * time.Millisecond)
	stats.FlushCounters(ctx)
	stats.FlushRequestLog(ctx)

	click, _ := g.Redis().Get(ctx, keys[0])
	dup, _ := g.Redis().Get(ctx, keys[1])
	logs, _ := g.Redis().LLen(ctx, keys[3])
	if click.Int() != 2 || dup.Int() != 1 || logs != 3 {
		t.Fatalf("click=%d dup=%d reqlog=%d, want 2/1/3", click.Int(), dup.Int(), logs)
	}
	if n := taolianTasks(X); n != 1 {
		t.Fatalf("taolian tasks for %s = %d, want 1 (duplicate must not be reported)", X, n)
	}

	// 跨零点: 昨天集合里已有的点击ID今天重发也算重复
	yKey := consts.KeyStatsClickSet(aid, time.Now().AddDate(0, 0, -1).Format("20060102"))
	defer g.Redis().Del(ctx, yKey)
	h := fnv.New64a()
	h.Write([]byte(Z))
	g.Redis().SAdd(ctx, yKey, strconv.FormatUint(h.Sum64(), 36))
	hit(Z)
	time.Sleep(200 * time.Millisecond)
	stats.FlushCounters(ctx)
	click, _ = g.Redis().Get(ctx, keys[0])
	dup, _ = g.Redis().Get(ctx, keys[1])
	if click.Int() != 2 || dup.Int() != 2 || taolianTasks(Z) != 0 {
		t.Fatalf("yesterday dup: click=%d dup=%d tasks=%d, want 2/2/0", click.Int(), dup.Int(), taolianTasks(Z))
	}

	// 清零今日统计不清去重集合: 清零后重发的老点击仍是重复, 不再上报
	if os.Getenv("IT_ALLOW_RESET") == "1" {
		if _, err := stats.ResetToday(ctx); err != nil {
			t.Fatal(err)
		}
		hit(X)
		time.Sleep(200 * time.Millisecond)
		stats.FlushCounters(ctx)
		click, _ = g.Redis().Get(ctx, keys[0])
		dup, _ = g.Redis().Get(ctx, keys[1])
		if click.Int() != 0 || dup.Int() != 1 || taolianTasks(X) != 1 {
			t.Fatalf("after reset: click=%d dup=%d tasks=%d, want 0/1/1", click.Int(), dup.Int(), taolianTasks(X))
		}
	}
}
