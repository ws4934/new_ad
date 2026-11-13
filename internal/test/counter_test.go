package test_test

import (
	"context"
	"testing"

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
