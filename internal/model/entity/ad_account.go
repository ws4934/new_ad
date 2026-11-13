// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// AdAccount is the golang structure for table ad_account.
type AdAccount struct {
	Id            int         `json:"id"            orm:"id"             description:""`                                         //
	PlatformId    int         `json:"platformId"    orm:"platform_id"    description:""`                                         //
	Name          string      `json:"name"          orm:"name"           description:""`                                         //
	Credentials   string      `json:"credentials"   orm:"credentials"    description:"账户凭证(company/advertiser_id/platform等)"`    // 账户凭证(company/advertiser_id/platform等)
	TaolianConfig string      `json:"taolianConfig" orm:"taolian_config" description:"淘联关联配置(channel/taskId/tbkId/sourcePack)"`  // 淘联关联配置(channel/taskId/tbkId/sourcePack)
	FakeEnabled   bool        `json:"fakeEnabled"   orm:"fake_enabled"   description:"虚假上报开关"`                                   // 虚假上报开关
	FakeRate      int         `json:"fakeRate"      orm:"fake_rate"      description:"虚假上报概率: 1-100"`                            // 虚假上报概率: 1-100
	FakeTypes     string      `json:"fakeTypes"     orm:"fake_types"     description:"虚假回传指定 taolian_type 列表,空则从已启用 mapping 随机"` // 虚假回传指定 taolian_type 列表,空则从已启用 mapping 随机
	FakeDelayMin  int         `json:"fakeDelayMin"  orm:"fake_delay_min" description:"虚假回传延迟下限(秒)"`                              // 虚假回传延迟下限(秒)
	FakeDelayMax  int         `json:"fakeDelayMax"  orm:"fake_delay_max" description:"虚假回传延迟上限(秒)"`                              // 虚假回传延迟上限(秒)
	Status        int         `json:"status"        orm:"status"         description:""`                                         //
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"     description:""`                                         //
	UpdatedAt     *gtime.Time `json:"updatedAt"     orm:"updated_at"     description:""`                                         //
	DebugMode     bool        `json:"debugMode"     orm:"debug_mode"     description:"联调模式开关"`                                   // 联调模式开关
	DebugEvents   string      `json:"debugEvents"   orm:"debug_events"   description:"联调关注的 transformType 列表"`                   // 联调关注的 transformType 列表
}
