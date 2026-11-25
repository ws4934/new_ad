// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ReportStats is the golang structure for table report_stats.
type ReportStats struct {
	Id          int64       `json:"id"          orm:"id"           description:""`                  //
	AccountId   int         `json:"accountId"   orm:"account_id"   description:""`                  //
	PlatformId  int         `json:"platformId"  orm:"platform_id"  description:""`                  //
	StatDate    *gtime.Time `json:"statDate"    orm:"stat_date"    description:""`                  //
	ClickCount  int         `json:"clickCount"  orm:"click_count"  description:""`                  //
	TaolianOk   int         `json:"taolianOk"   orm:"taolian_ok"   description:""`                  //
	TaolianErr  int         `json:"taolianErr"  orm:"taolian_err"  description:""`                  //
	CallbackOk  int         `json:"callbackOk"  orm:"callback_ok"  description:""`                  //
	CallbackErr int         `json:"callbackErr" orm:"callback_err" description:""`                  //
	FakeOk      int         `json:"fakeOk"      orm:"fake_ok"      description:""`                  //
	FakeErr     int         `json:"fakeErr"     orm:"fake_err"     description:""`                  //
	DropCount   int         `json:"dropCount"   orm:"drop_count"   description:"监测点击丢弃"`            // 监测点击丢弃
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""`                  //
	DebugOk     int         `json:"debugOk"     orm:"debug_ok"     description:"联调模拟回传成功"`          // 联调模拟回传成功
	DebugErr    int         `json:"debugErr"    orm:"debug_err"    description:"联调模拟回传失败"`          // 联调模拟回传失败
	CbRecv      int         `json:"cbRecv"      orm:"cb_recv"      description:"收到淘联转化回调"`          // 收到淘联转化回调
	CbDrop      int         `json:"cbDrop"      orm:"cb_drop"      description:"淘联回调丢弃"`            // 淘联回调丢弃
	DupCount    int         `json:"dupCount"    orm:"dup_count"    description:"重复点击 (同点击ID, 已丢弃)"` // 重复点击 (同点击ID, 已丢弃)
}
