// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// PlatformCsvStats is the golang structure for table platform_csv_stats.
type PlatformCsvStats struct {
	Id            int64       `json:"id"            orm:"id"               description:""` //
	PlatformId    int         `json:"platformId"    orm:"platform_id"      description:""` //
	StatDate      *gtime.Time `json:"statDate"      orm:"stat_date"        description:""` //
	CsvNewImeiMd5 int         `json:"csvNewImeiMd5" orm:"csv_new_imei_md5" description:""` //
	CsvNewIdfaMd5 int         `json:"csvNewIdfaMd5" orm:"csv_new_idfa_md5" description:""` //
	CsvNewOaidMd5 int         `json:"csvNewOaidMd5" orm:"csv_new_oaid_md5" description:""` //
	CsvNewCaidMd5 int         `json:"csvNewCaidMd5" orm:"csv_new_caid_md5" description:""` //
	CreatedAt     *gtime.Time `json:"createdAt"     orm:"created_at"       description:""` //
}
