// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ReportStats is the golang structure of table report_stats for DAO operations like Where/Data.
type ReportStats struct {
	g.Meta      `orm:"table:report_stats, do:true"`
	Id          any         //
	AccountId   any         //
	PlatformId  any         //
	StatDate    *gtime.Time //
	ClickCount  any         //
	TaolianOk   any         //
	TaolianErr  any         //
	CallbackOk  any         //
	CallbackErr any         //
	FakeOk      any         //
	FakeErr     any         //
	DropCount   any         //
	CreatedAt   *gtime.Time //
}
