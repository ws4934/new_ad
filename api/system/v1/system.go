package v1

import (
	"ad-monitor/internal/logic/updater"

	"github.com/gogf/gf/v2/frame/g"
)

type UpdateStatusReq struct {
	g.Meta `path:"/api/v2/system/update" tags:"System" method:"get" summary:"在线更新状态"`
}
type UpdateStatusRes struct{ updater.Status }

type UpdateCheckReq struct {
	g.Meta `path:"/api/v2/system/update/check" tags:"System" method:"post" summary:"检查更新"`
}
type UpdateCheckRes struct{ updater.Status }

type UpdateStartReq struct {
	g.Meta `path:"/api/v2/system/update/start" tags:"System" method:"post" summary:"开始更新"`
}
type UpdateStartRes struct{ updater.Status }
