// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// PlatformCsvStats is the golang structure of table platform_csv_stats for DAO operations like Where/Data.
type PlatformCsvStats struct {
	g.Meta        `orm:"table:platform_csv_stats, do:true"`
	Id            any         //
	PlatformId    any         //
	StatDate      *gtime.Time //
	CsvNewImeiMd5 any         //
	CsvNewIdfaMd5 any         //
	CsvNewOaidMd5 any         //
	CsvNewCaidMd5 any         //
	CreatedAt     *gtime.Time //
}
