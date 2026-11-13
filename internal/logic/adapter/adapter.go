package adapter

import (
	"context"

	"ad-monitor/internal/model/bo"
)

// Adapter 平台适配器接口
// 每个广告平台实现此接口
type Adapter interface {
	// Name 平台标识 (与 ad_platform.name 一致)
	Name() string

	// Spec 返回平台宏元数据
	Spec() MacroSpec

	// ReportConversion 向广告平台回传转化数据
	// 这是唯一真正有差异的方法: 不同平台 GET/POST/JSON/字段名均不同
	ReportConversion(ctx context.Context, payload *bo.CallbackPayload) (*bo.ReportResult, error)
}
