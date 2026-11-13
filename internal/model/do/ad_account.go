// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// AdAccount is the golang structure of table ad_account for DAO operations like Where/Data.
type AdAccount struct {
	g.Meta        `orm:"table:ad_account, do:true"`
	Id            any         //
	PlatformId    any         //
	Name          any         //
	Credentials   any         // 账户凭证(company/advertiser_id/platform等)
	TaolianConfig any         // 淘联关联配置(channel/taskId/tbkId/sourcePack)
	FakeEnabled   any         // 虚假上报开关
	FakeRate      any         // 虚假上报概率: 1-100
	FakeTypes     any         // 虚假回传指定 taolian_type 列表,空则从已启用 mapping 随机
	FakeDelayMin  any         // 虚假回传延迟下限(秒)
	FakeDelayMax  any         // 虚假回传延迟上限(秒)
	Status        any         //
	CreatedAt     *gtime.Time //
	UpdatedAt     *gtime.Time //
	DebugMode     any         // 联调模式开关
	DebugEvents   any         // 联调关注的 transformType 列表
}
