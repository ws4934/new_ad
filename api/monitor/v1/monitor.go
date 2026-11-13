package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MonitorReq 广告平台监测回调入口
// 所有 query 参数由 adapter.ParseClickRequest 从 ghttp.Request 解析, 不在此声明
type MonitorReq struct {
	g.Meta    `path:"/m/{accountId}" tags:"Monitor" method:"get" summary:"广告平台监测链接回调"`
	AccountID int `json:"accountId" v:"required"`
}

type MonitorRes struct {
	g.Meta `mime:"text/plain"`
}
