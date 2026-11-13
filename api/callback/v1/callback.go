package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CallbackReq 淘联转化回传接收入口
// 淘联会把我方填入的 callbackUrl 原样请求过来, 并在 query 后追加 &transformType=xxx
// query: a={accountId}&p={platformName}&c={clickId}&transformType={x}&adid={...}&...
type CallbackReq struct {
	g.Meta        `path:"/cb" tags:"Callback" method:"get" summary:"淘联转化回传"`
	A             int    `json:"a"             v:"required"` // accountId
	P             string `json:"p"             v:"required"` // platformName
	C             string `json:"c"             v:"required"` // clickId
	D             string `json:"d"`                           // 设备号 (base64 JSON, 由淘联上报时透传)
	TransformType int    `json:"transformType" v:"required"` // 淘联 transformType
}

type CallbackRes struct {
	g.Meta `mime:"text/plain"`
}
