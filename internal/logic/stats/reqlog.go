package stats

import (
	"context"
	"sync"
	"time"

	"ad-monitor/internal/consts"

	"github.com/gogf/gf/v2/frame/g"
)

// 每账户最近 reqLogMax 条监测请求, 调试观察用 (宏有没有被替换、参数长什么样)。
// 热路径只做进程内追加, 随计数器 worker 每秒落 Redis (一条脚本: LPUSH + LTRIM + 续期);
// 写失败直接丢 (调试数据, 不重试)。7 天无请求自动过期, 删掉的账户不会残留。
// 查看: redis-cli LRANGE reqlog:<账户ID> 0 -1  (表头最新)
const (
	reqLogMax = 50
	reqLogTTL = 7 * 24 * 3600
)

// reqLogLua KEYS[1]=列表, ARGV[1]=保留条数, ARGV[2]=TTL, ARGV[3..]=按时间顺序的新记录
const reqLogLua = `redis.call('LPUSH', KEYS[1], unpack(ARGV, 3))
redis.call('LTRIM', KEYS[1], 0, tonumber(ARGV[1]) - 1)
redis.call('EXPIRE', KEYS[1], ARGV[2])
return 1`

var (
	reqMu      sync.Mutex
	reqPending = map[int][]any{}
)

// RecordRequest 记录一条监测请求 (每账户只攒最新 reqLogMax 条)
func RecordRequest(accountID int, rawQuery string) {
	line := time.Now().Format("2006-01-02 15:04:05.000") + " " + rawQuery
	reqMu.Lock()
	l := append(reqPending[accountID], line)
	if len(l) > reqLogMax {
		l = l[1:]
	}
	reqPending[accountID] = l
	reqMu.Unlock()
}

// RequestLog 读取账户最近的监测请求 (最新在前)
func RequestLog(ctx context.Context, accountID int) ([]string, error) {
	vs, err := g.Redis().LRange(ctx, consts.KeyReqLog(accountID), 0, reqLogMax-1)
	if err != nil {
		return nil, err
	}
	return vs.Strings(), nil
}

// FlushRequestLog 按时间顺序 LPUSH (表头为最新), 再裁到 reqLogMax 条
func FlushRequestLog(ctx context.Context) {
	reqMu.Lock()
	if len(reqPending) == 0 {
		reqMu.Unlock()
		return
	}
	batch := reqPending
	reqPending = make(map[int][]any, len(batch))
	reqMu.Unlock()

	for aid, lines := range batch {
		args := append([]any{reqLogLua, 1, consts.KeyReqLog(aid), reqLogMax, reqLogTTL}, lines...)
		_, _ = g.Redis().Do(ctx, "EVAL", args...)
	}
}
