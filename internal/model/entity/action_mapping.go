// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ActionMapping is the golang structure for table action_mapping.
type ActionMapping struct {
	Id           int         `json:"id"           orm:"id"            description:""` //
	AccountId    int         `json:"accountId"    orm:"account_id"    description:""` //
	TaolianType  int         `json:"taolianType"  orm:"taolian_type"  description:""` //
	TaolianDesc  string      `json:"taolianDesc"  orm:"taolian_desc"  description:""` //
	PlatformType string      `json:"platformType" orm:"platform_type" description:""` //
	PlatformDesc string      `json:"platformDesc" orm:"platform_desc" description:""` //
	Enabled      bool        `json:"enabled"      orm:"enabled"       description:""` //
	CreatedAt    *gtime.Time `json:"createdAt"    orm:"created_at"    description:""` //
	UpdatedAt    *gtime.Time `json:"updatedAt"    orm:"updated_at"    description:""` //
}
