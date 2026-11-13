// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// AdPlatform is the golang structure of table ad_platform for DAO operations like Where/Data.
type AdPlatform struct {
	g.Meta      `orm:"table:ad_platform, do:true"`
	Id          any         //
	Name        any         // 平台标识: weibo / xiaohongshu
	DisplayName any         //
	Config      any         // 平台级配置(reportUrl/caidVersions等)
	Status      any         //
	CreatedAt   *gtime.Time //
	UpdatedAt   *gtime.Time //
}
