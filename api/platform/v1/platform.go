package v1

import (
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// --- 列表 ---

type PlatformListReq struct {
	g.Meta `path:"/api/v2/platforms" tags:"Platform" method:"get" summary:"平台列表"`
}

type PlatformListRes struct {
	List []*entity.AdPlatform `json:"list"`
}

// --- 创建 ---

type PlatformCreateReq struct {
	g.Meta      `path:"/api/v2/platforms" tags:"Platform" method:"post" summary:"创建平台"`
	Name        string `json:"name"        v:"required"`
	DisplayName string `json:"displayName" v:"required"`
	Config      string `json:"config"`
}

type PlatformCreateRes struct {
	Id int `json:"id"`
}

// --- 更新 ---

type PlatformUpdateReq struct {
	g.Meta      `path:"/api/v2/platforms/{id}" tags:"Platform" method:"put" summary:"更新平台"`
	Id          int    `json:"id"          v:"required"`
	DisplayName string `json:"displayName"`
	Config      string `json:"config"`
	Status      *int   `json:"status"`
}

type PlatformUpdateRes struct{}

// PlatformDeleteReq 删除平台。
type PlatformDeleteReq struct {
	g.Meta `path:"/api/v2/platforms/{id}" tags:"Platform" method:"delete" summary:"删除平台"`
	Id     int `json:"id" v:"required|min:1"`
}

type PlatformDeleteRes struct{}
