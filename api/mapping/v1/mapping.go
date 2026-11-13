package v1

import (
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// --- 列表 (按账户) ---

type MappingListReq struct {
	g.Meta    `path:"/api/v2/mappings" tags:"Mapping" method:"get" summary:"行为映射列表"`
	AccountId int `json:"accountId" v:"required"`
}

type MappingListRes struct {
	List []*entity.ActionMapping `json:"list"`
}

// --- 创建/更新 (批量保存) ---

type MappingSaveReq struct {
	g.Meta    `path:"/api/v2/mappings" tags:"Mapping" method:"post" summary:"批量保存行为映射"`
	AccountId int              `json:"accountId" v:"required"`
	Items     []*MappingSaveItem `json:"items"`
}

type MappingSaveItem struct {
	TaolianType  int    `json:"taolianType"  v:"required"`
	TaolianDesc  string `json:"taolianDesc"`
	PlatformType string `json:"platformType" v:"required"`
	PlatformDesc string `json:"platformDesc"`
	Enabled      bool   `json:"enabled"`
}

type MappingSaveRes struct{}

// --- 删除 ---

type MappingDeleteReq struct {
	g.Meta `path:"/api/v2/mappings/{id}" tags:"Mapping" method:"delete" summary:"删除行为映射"`
	Id     int `json:"id" v:"required"`
}

type MappingDeleteRes struct{}

// --- 枚举选项 ---

type MappingOptionsReq struct {
	g.Meta `path:"/api/v2/mapping-options" tags:"Mapping" method:"get" summary:"行为映射枚举选项"`
}

type TypeOption struct {
	Type interface{} `json:"type"` // int (淘联) 或 string (平台)
	Desc string      `json:"desc"`
}

type MappingOptionsRes struct {
	TaolianTypes  []*TypeOption              `json:"taolianTypes"`
	PlatformTypes map[string][]*TypeOption    `json:"platformTypes"`
}
