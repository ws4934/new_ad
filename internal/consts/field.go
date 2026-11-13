package consts

// InternalField 内部标准字段枚举
// 目的: 解耦各平台的字段名差异, 业务逻辑只认内部字段
type InternalField string

const (
	FieldClickID    InternalField = "click_id"    // 平台点击标识 (微博=IMP / 小红书=click_id)
	FieldIDFA       InternalField = "idfa"        // iOS 原值
	FieldIDFAMD5    InternalField = "idfa_md5"    // iOS md5
	FieldIMEI       InternalField = "imei"        // Android 原值
	FieldIMEIMD5    InternalField = "imei_md5"    // Android md5
	FieldOAID       InternalField = "oaid"
	FieldOAIDMD5    InternalField = "oaid_md5"
	FieldCAID       InternalField = "caid"        // 逗号分隔多版本原值
	FieldCAIDMD5    InternalField = "caid_md5"    // 逗号分隔多版本 md5
	FieldAndroidID  InternalField = "androidid"
	FieldOS         InternalField = "os"
	FieldIP         InternalField = "ip"
	FieldUA         InternalField = "ua"
	FieldTS         InternalField = "ts"
	FieldAdID         InternalField = "ad_id"
	FieldCreativeID   InternalField = "creative_id"
	FieldAdvertiserID InternalField = "advertiser_id" // 广告主ID (小红书从 __ADVERTISER_ID__ 宏获取)
)

// DeviceFields 所有用于设备号去重的字段
var DeviceFields = []InternalField{
	FieldIDFA, FieldIDFAMD5,
	FieldIMEI, FieldIMEIMD5,
	FieldOAID, FieldOAIDMD5,
	FieldCAID, FieldCAIDMD5,
}
