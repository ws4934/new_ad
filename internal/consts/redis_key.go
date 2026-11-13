package consts

import "fmt"

// Redis key 模板 (使用函数封装, 避免散落在各处的 fmt.Sprintf 拼接错误)

// --- 缓存 ---
func KeyCacheAccount(id int) string  { return fmt.Sprintf("cache:account:%d", id) }
func KeyCacheMappings(id int) string { return fmt.Sprintf("cache:mappings:%d", id) }
func KeyCachePlatform(id int) string { return fmt.Sprintf("cache:platform:%d", id) }

// --- 计数器 (day 格式: 20060102) ---
func KeyStatsClick(aid int, day string) string       { return fmt.Sprintf("stats:click:%d:%s", aid, day) }
func KeyStatsTaolianOk(aid int, day string) string   { return fmt.Sprintf("stats:taolian:ok:%d:%s", aid, day) }
func KeyStatsTaolianErr(aid int, day string) string  { return fmt.Sprintf("stats:taolian:err:%d:%s", aid, day) }
func KeyStatsCallbackOk(aid int, day string) string  { return fmt.Sprintf("stats:callback:ok:%d:%s", aid, day) }
func KeyStatsCallbackErr(aid int, day string) string { return fmt.Sprintf("stats:callback:err:%d:%s", aid, day) }
func KeyStatsFakeOk(aid int, day string) string      { return fmt.Sprintf("stats:fake:ok:%d:%s", aid, day) }
func KeyStatsFakeErr(aid int, day string) string     { return fmt.Sprintf("stats:fake:err:%d:%s", aid, day) }
func KeyStatsDrop(aid int, day string) string        { return fmt.Sprintf("stats:drop:%d:%s", aid, day) }

// --- QPS (秒级, TTL 120s) ---
func KeyQpsGlobal(sec int64) string { return fmt.Sprintf("qps:g:%d", sec) }

// --- 登录失败计数 (IP 级, TTL 10 分钟) ---
func KeyLoginFail(ip string) string { return "auth:loginfail:" + ip }
