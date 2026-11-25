package stats

import (
	"context"
	"strings"
	"sync"
	"time"

	"ad-monitor/internal/consts"

	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/frame/g"
)

// Today 返回今日日期 (20060102 格式)
func Today() string {
	return time.Now().Format("20060102")
}

// 业务计数器批量聚合:
// IncrCounter 只做进程内 map 累加 (微秒级, 热路径无 Redis RTT),
// 常驻 worker 每秒把累积量用 INCRBY 批量写入 Redis。
// 正常退出由 ShutdownPlugin 调用 FlushCounters, 不丢数据;
// 硬崩溃 (kill -9/OOM) 最多丢 1 秒窗口 — 与 Redis AOF everysec 的持久化等级一致。
var (
	cntMu      sync.Mutex
	cntPending = map[string]int64{} // redis key → 累积增量
	// flushMu 串行化落盘与清零: 避免清零时把 "已从内存取出、还没写进 Redis" 的旧增量写回去
	flushMu sync.Mutex
)

const counterFlushInterval = time.Second

// IncrCounter 通用计数器, 进程内聚合后由 worker 批量落 Redis
// kind: click / taolian:ok / taolian:err / callback:ok / callback:err / fake:ok / fake:err / drop
//
//	/ debug:ok / debug:err (联调模拟回传) / cb:recv / cb:drop (淘联转化回调)
func IncrCounter(ctx context.Context, kind string, accountID int) {
	day := Today()
	var key string
	switch kind {
	case "click":
		key = consts.KeyStatsClick(accountID, day)
	case "taolian:ok":
		key = consts.KeyStatsTaolianOk(accountID, day)
	case "taolian:err":
		key = consts.KeyStatsTaolianErr(accountID, day)
	case "callback:ok":
		key = consts.KeyStatsCallbackOk(accountID, day)
	case "callback:err":
		key = consts.KeyStatsCallbackErr(accountID, day)
	case "fake:ok":
		key = consts.KeyStatsFakeOk(accountID, day)
	case "fake:err":
		key = consts.KeyStatsFakeErr(accountID, day)
	case "drop":
		key = consts.KeyStatsDrop(accountID, day)
	case "debug:ok":
		key = consts.KeyStatsDebugOk(accountID, day)
	case "debug:err":
		key = consts.KeyStatsDebugErr(accountID, day)
	case "cb:recv":
		key = consts.KeyStatsCbRecv(accountID, day)
	case "cb:drop":
		key = consts.KeyStatsCbDrop(accountID, day)
	default:
		g.Log().Warningf(ctx, "unknown counter kind: %s", kind)
		return
	}
	cntMu.Lock()
	cntPending[key]++
	cntMu.Unlock()
}

// FlushCounters 把累积的计数批量写入 Redis (INCRBY + 35 天 TTL)
// 单个 key 写失败则回灌内存, 下个周期重试
func FlushCounters(ctx context.Context) {
	flushMu.Lock()
	defer flushMu.Unlock()
	cntMu.Lock()
	if len(cntPending) == 0 {
		cntMu.Unlock()
		return
	}
	batch := cntPending
	cntPending = make(map[string]int64, len(batch))
	cntMu.Unlock()

	ttl := int64((35 * 24 * time.Hour).Seconds())
	for k, n := range batch {
		if _, err := g.Redis().IncrBy(ctx, k, n); err != nil {
			g.Log().Warningf(ctx, "flush counter failed: %s, err: %v", k, err)
			cntMu.Lock()
			cntPending[k] += n // 回灌重试
			cntMu.Unlock()
			continue
		}
		_, _ = g.Redis().Expire(ctx, k, ttl)
	}
}

// ResetToday 清零所有账户今日业务计数 (含进程内未落盘的增量), 返回删除的 key 数
func ResetToday(ctx context.Context) (int, error) {
	flushMu.Lock()
	defer flushMu.Unlock()

	suffix := ":" + Today()
	cntMu.Lock()
	for k := range cntPending {
		if strings.HasSuffix(k, suffix) {
			delete(cntPending, k)
		}
	}
	cntMu.Unlock()

	deleted := 0
	var cursor uint64
	for {
		next, keys, err := g.Redis().Scan(ctx, cursor, gredis.ScanOption{Match: "stats:*" + suffix, Count: 500})
		if err != nil {
			return deleted, err
		}
		if len(keys) > 0 {
			n, err := g.Redis().Del(ctx, keys...)
			if err != nil {
				return deleted, err
			}
			deleted += int(n)
		}
		if next == 0 {
			return deleted, nil
		}
		cursor = next
	}
}

// StartCounterWorker 启动计数器落盘 worker, 进程启动时调用一次
func StartCounterWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(counterFlushInterval)
		defer ticker.Stop()
		for range ticker.C {
			FlushCounters(ctx)
			FlushRequestLog(ctx)
		}
	}()
}
