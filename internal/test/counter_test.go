package test_test

import (
	"context"
	"testing"
	"time"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/logic/stats"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/test/gtest"
)

// TestIntegration_CounterBatchFlush 批量计数器: 内存聚合 → FlushCounters → Redis
func TestIntegration_CounterBatchFlush(t *testing.T) {
	ctx := context.Background()
	const aid = 987654 // 不存在的账户 ID, 保证 key 不与其他测试共享
	day := stats.Today()

	keys := []string{
		consts.KeyStatsClick(aid, day),
		consts.KeyStatsTaolianOk(aid, day),
		consts.KeyStatsDrop(aid, day),
	}
	defer func() {
		for _, k := range keys {
			g.Redis().Del(ctx, k)
		}
	}()
	for _, k := range keys {
		g.Redis().Del(ctx, k)
	}

	// 聚合阶段: 未 flush 前 Redis 不应有值
	for i := 0; i < 5; i++ {
		stats.IncrCounter(ctx, "click", aid)
	}
	stats.IncrCounter(ctx, "taolian:ok", aid)
	stats.IncrCounter(ctx, "taolian:ok", aid)
	stats.IncrCounter(ctx, "drop", aid)

	// flush 后批量写入
	stats.FlushCounters(ctx)

	v, _ := g.Redis().Get(ctx, keys[0])
	gtest.Assert(v.Int(), 5)
	v, _ = g.Redis().Get(ctx, keys[1])
	gtest.Assert(v.Int(), 2)
	v, _ = g.Redis().Get(ctx, keys[2])
	gtest.Assert(v.Int(), 1)

	// TTL 已设置 (35 天)
	ttl, _ := g.Redis().TTL(ctx, keys[0])
	gtest.Assert(ttl > 0, true)

	// 再次累加再 flush: 增量叠加而非覆盖
	stats.IncrCounter(ctx, "click", aid)
	stats.FlushCounters(ctx)
	v, _ = g.Redis().Get(ctx, keys[0])
	gtest.Assert(v.Int(), 6)

	// 空 pending 时 flush 是 no-op, 不报错
	stats.FlushCounters(ctx)
	v, _ = g.Redis().Get(ctx, keys[0])
	gtest.Assert(v.Int(), 6)

	// 未知 kind 不写入任何 key
	stats.IncrCounter(ctx, "bogus-kind", aid)
	stats.FlushCounters(ctx)
	v, _ = g.Redis().Get(ctx, keys[0])
	gtest.Assert(v.Int(), 6)
}

// TestIntegration_ResetToday 清零今日计数: Redis 已落盘 + 内存未落盘的都清掉; 昨天的计数和非统计 key 不动
func TestIntegration_ResetToday(t *testing.T) {
	ctx := context.Background()
	const aid = 987656
	day := stats.Today()
	yesterday := time.Now().AddDate(0, 0, -1).Format("20060102")
	keep := []string{consts.KeyStatsClick(aid, yesterday), consts.KeyReqLog(aid)}
	defer func() {
		g.Redis().Del(ctx, append(keep, consts.KeyStatsClick(aid, day), consts.KeyStatsCbRecv(aid, day))...)
	}()

	stats.IncrCounter(ctx, "click", aid)
	stats.IncrCounter(ctx, "cb:recv", aid)
	stats.FlushCounters(ctx)             // 已落盘
	stats.IncrCounter(ctx, "click", aid) // 内存未落盘
	g.Redis().Set(ctx, keep[0], 5)       // 昨天
	g.Redis().LPush(ctx, keep[1], "x")   // 请求记录

	n, err := stats.ResetToday(ctx)
	if err != nil || n < 2 {
		t.Fatalf("reset: n=%d err=%v", n, err)
	}
	stats.FlushCounters(ctx) // 清零前攒的增量不能被写回
	for _, k := range []string{consts.KeyStatsClick(aid, day), consts.KeyStatsCbRecv(aid, day)} {
		if c, _ := g.Redis().Exists(ctx, k); c != 0 {
			t.Errorf("%s survived reset", k)
		}
	}
	for _, k := range keep {
		if c, _ := g.Redis().Exists(ctx, k); c != 1 {
			t.Errorf("%s should be kept", k)
		}
	}
	// 清零后新计数正常
	stats.IncrCounter(ctx, "click", aid)
	stats.FlushCounters(ctx)
	if v, _ := g.Redis().Get(ctx, consts.KeyStatsClick(aid, day)); v.Int() != 1 {
		t.Errorf("click after reset = %d, want 1", v.Int())
	}
}
