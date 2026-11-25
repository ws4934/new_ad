package bo

import "ad-monitor/internal/consts"

// TaolianTaskPayload 淘联上报任务 payload
type TaolianTaskPayload struct {
	AccountID int                             `json:"aid"`
	ClickID   string                          `json:"cid"`
	Devices   map[consts.InternalField]string `json:"dev"`
	TS        int64                           `json:"ts"` // 点击时间戳(秒)
}

// CallbackTaskPayload 广告平台真实回传任务 payload
type CallbackTaskPayload struct {
	AccountID   int                             `json:"aid"`
	ClickID     string                          `json:"cid"`
	TaolianType int                             `json:"tt"`
	MappedType  string                          `json:"mt"`
	Devices     map[consts.InternalField]string `json:"dev"`
	TS          int64                           `json:"ts"`            // 转化时间(毫秒), 入队时确定, 重试不变
	Debug       bool                            `json:"dbg,omitempty"` // 联调模式模拟回传 (单独计数)
}

// FakeTaskPayload 虚假回传任务 payload
// FakeTypes/Devices 都在入队时就决定好, worker 执行时不再查配置
type FakeTaskPayload struct {
	AccountID   int                             `json:"aid"`
	ClickID     string                          `json:"cid"`
	Devices     map[consts.InternalField]string `json:"dev"`
	TaolianType int                             `json:"tt"`
	MappedType  string                          `json:"mt"`
	TS          int64                           `json:"ts"` // 转化时间(毫秒) = 计划执行时刻, 重试不变
}
