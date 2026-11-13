package httpx

import (
	"net"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/net/gclient"
)

// sharedTransport 全局共享连接池。
// g.Client() 每次调用都会新建 client + 新建 Transport (且 gf 默认禁用 keep-alive),
// 高频上报场景下每个请求都要重新 TCP+TLS 握手。这里统一复用一个开启
// keep-alive 的 Transport, 上报与回传共享连接池。
var sharedTransport = &http.Transport{
	DialContext: (&net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	MaxIdleConns:        128,
	MaxIdleConnsPerHost: 32,
	IdleConnTimeout:     90 * time.Second,
	TLSHandshakeTimeout: 10 * time.Second,
}

// New 返回带总超时 + 共享连接池的 HTTP 客户端。
// 注意: gclient 的链式方法 (ContentJson 等) 会修改 client 自身,
// 因此 client 实例不能跨 goroutine 共享 — 每次 New, 仅复用底层 Transport。
//
// 总超时 10s: 防止平台接口挂起时无限占住 asynq worker
// (asynq 任务默认没有 deadline, gclient 默认 Timeout=0 也不限制)。
func New() *gclient.Client {
	c := gclient.New()
	c.SetTimeout(10 * time.Second)
	c.Transport = sharedTransport
	return c
}
