package bo

import (
	"ad-monitor/internal/consts"
	"ad-monitor/internal/model/entity"
)

// CallbackPayload 回传广告平台时的上下文
type CallbackPayload struct {
	Account     *entity.AdAccount
	ClickID     string                          // 平台点击标识 (IMP / click_id)
	Devices     map[consts.InternalField]string // 归因设备号 (已过滤宏)
	MappedType  string                          // 映射后的 action_type / event_type
	TaolianType int                             // 原始 transformType
	ConvTS      int64                           // 转化时间(毫秒), 同一任务重试时不变
	IsFake      bool
}

// ReportResult 向广告平台回传的结果
type ReportResult struct {
	HTTPStatus int
	RawBody    string
	Success    bool
}
