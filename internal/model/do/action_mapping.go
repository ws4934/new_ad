// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ActionMapping is the golang structure of table action_mapping for DAO operations like Where/Data.
type ActionMapping struct {
	g.Meta       `orm:"table:action_mapping, do:true"`
	Id           any         //
	AccountId    any         //
	TaolianType  any         //
	TaolianDesc  any         //
	PlatformType any         //
	PlatformDesc any         //
	Enabled      any         //
	CreatedAt    *gtime.Time //
	UpdatedAt    *gtime.Time //
}
