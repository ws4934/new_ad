package stats

import (
	"context"
	"sync"
	"time"

	"ad-monitor/internal/consts"

	"github.com/gogf/gf/v2/frame/g"
)

// 每账户最近 reqLogMax 条监测请求 (reqlog:<aid>) 和淘联回调 (cblog:<aid>), 调试观察用
// (宏有没有被替换、淘联回调带了什么)。
// 热路径只做进程内追加, 随计数器 worker 每秒落 Redis (一条脚本: LPUSH + LTRIM + 续期);
// 写失败直接丢 (调试数据, 不重试)。7 天无请求自动过期, 删掉的账户不会残留。
// 查看: redis-cli LRANGE reqlog:<账户ID> 0 -1 / cblog:<账户ID>  (表头最新)
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
	reqPending = map[string][]any{} // redis key → 待落盘记录
)

// RecordRequest 记录一条监测请求
func RecordRequest(accountID int, rawQuery string) { record(consts.KeyReqLog(accountID), rawQuery) }

// RecordCallback 记录一条淘联回调
func RecordCallback(accountID int, rawQuery string) { record(consts.KeyCbLog(accountID), rawQuery) }

// record 每个 key 只攒最新 reqLogMax 条
func record(key, rawQuery string) {
	line := time.Now().Format("2006-01-02 15:04:05.000") + " " + rawQuery
	reqMu.Lock()
	l := append(reqPending[key], line)
	if len(l) > reqLogMax {
		l = l[1:]
	}
	reqPending[key] = l
	reqMu.Unlock()
}

// RequestLog 读取最近的记录 (最新在前), key 为 consts.KeyReqLog / consts.KeyCbLog
func RequestLog(ctx context.Context, key string) ([]string, error) {
	vs, err := g.Redis().LRange(ctx, key, 0, reqLogMax-1)
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
	reqPending = make(map[string][]any, len(batch))
	reqMu.Unlock()

	for key, lines := range batch {
		args := append([]any{reqLogLua, 1, key, reqLogMax, reqLogTTL}, lines...)
		_, _ = g.Redis().Do(ctx, "EVAL", args...)
	}
}
