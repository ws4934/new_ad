// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ReportStats is the golang structure for table report_stats.
type ReportStats struct {
	Id          int64       `json:"id"          orm:"id"           description:""` //
	AccountId   int         `json:"accountId"   orm:"account_id"   description:""` //
	PlatformId  int         `json:"platformId"  orm:"platform_id"  description:""` //
	StatDate    *gtime.Time `json:"statDate"    orm:"stat_date"    description:""` //
	ClickCount  int         `json:"clickCount"  orm:"click_count"  description:""` //
	TaolianOk   int         `json:"taolianOk"   orm:"taolian_ok"   description:""` //
	TaolianErr  int         `json:"taolianErr"  orm:"taolian_err"  description:""` //
	CallbackOk  int         `json:"callbackOk"  orm:"callback_ok"  description:""` //
	CallbackErr int         `json:"callbackErr" orm:"callback_err" description:""` //
	FakeOk      int         `json:"fakeOk"      orm:"fake_ok"      description:""` //
	FakeErr     int         `json:"fakeErr"     orm:"fake_err"     description:""` //
	DropCount   int         `json:"dropCount"   orm:"drop_count"   description:""` //
	CreatedAt   *gtime.Time `json:"createdAt"   orm:"created_at"   description:""` //
}
