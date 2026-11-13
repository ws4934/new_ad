package stats

import (
	"context"
	"time"

	"ad-monitor/internal/consts"

	"github.com/gogf/gf/v2/frame/g"
)

// clickCh QPS 打点队列, 满了直接丢弃 (统计数据, 可容忍少量丢失)
var clickCh = make(chan struct{}, 8192)

// RecordClickAsync 非阻塞打点, 由常驻 worker 批量消费
func RecordClickAsync() {
	select {
	case clickCh <- struct{}{}:
	default: // 队列满, 丢弃
	}
}

// StartClickWorker 启动 QPS 打点常驻 worker, 进程启动时调用一次
// 批量聚合: 每 200ms 用一次 INCRBY 把累积的点击数写进当前秒的 key,
// 替代逐条 INCR — 高 QPS 下单条写法 (1 RTT/点击) 跟不上会丢打点。
// 代价: 秒边界上 ≤200ms 的归桶偏移, 对 QPS 观测可忽略。
func StartClickWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		pending := int64(0)
		for {
			select {
			case <-clickCh:
				pending++
			case <-ticker.C:
				if pending == 0 {
					continue
				}
				key := consts.KeyQpsGlobal(time.Now().Unix())
				if _, err := g.Redis().IncrBy(ctx, key, pending); err != nil {
					continue // Redis 暂不可用, 保留 pending 下个周期重试
				}
				_, _ = g.Redis().Expire(ctx, key, 120)
				pending = 0
			}
		}
	}()
}

// QPSMetric QPS 多维度指标
type QPSMetric struct {
	Current int     `json:"current"` // 最近 1 秒
	Avg10s  float64 `json:"avg10s"`  // 最近 10 秒均值
	Avg60s  float64 `json:"avg60s"`  // 最近 60 秒均值
	Peak5m  int     `json:"peak5m"`  // 最近 5 分钟内任意 1 秒峰值
}

// GetGlobalQPS 读取全局 QPS 指标
// 实现: 构造最近 300 秒的 key, MGET 批量读取, 在内存中聚合
func GetGlobalQPS(ctx context.Context) (*QPSMetric, error) {
	now := time.Now().Unix()
	// 构造最近 300 秒的 key (按秒逆序)
	keys := make([]string, 300)
	for i := 0; i < 300; i++ {
		keys[i] = consts.KeyQpsGlobal(now - int64(i) - 1)
	}

	// MGet 返回 map[string]*gvar.Var
	result, err := g.Redis().MGet(ctx, keys...)
	if err != nil {
		return nil, err
	}

	m := &QPSMetric{}
	// 最近 1 秒 (keys[0])
	if v, ok := result[keys[0]]; ok && v != nil {
		m.Current = v.Int()
	}
	var sum10, sum60 int
	for i, k := range keys {
		v := result[k]
		if v == nil {
			continue
		}
		n := v.Int()
		if i < 10 {
			sum10 += n
		}
		if i < 60 {
			sum60 += n
		}
		if n > m.Peak5m {
			m.Peak5m = n
		}
	}
	m.Avg10s = float64(sum10) / 10.0
	m.Avg60s = float64(sum60) / 60.0
	return m, nil
}
