// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// AdPlatform is the golang structure for table ad_platform.
type AdPlatform struct {
	Id          int         `json:"id"          orm:"id"           description:""`                               //
	Name        string      `json:"name"        orm:"name"         description:"平台标识: weibo / xiaohongshu"`      // 平台标识: weibo / xiaohongshu
	DisplayName string      `json:"displayName" orm:"display_name" description:""`                               //
	Config      string      `json:"config"      orm:"config"       description:"平台级配置(reportUrl/caidVersions等)"` // 平台级配置(reportUrl/caidVersions等)
	Status      int         `json:"status"      orm:"status"       description:""`                               //
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""`                               //
	UpdatedAt   *gtime.Time `json:"updatedAt"   orm:"updated_at"   description:""`                               //
}
