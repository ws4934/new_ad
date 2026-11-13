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
}

// FakeTaskPayload 虚假回传任务 payload
// FakeTypes/Devices 都在入队时就决定好, worker 执行时不再查配置
type FakeTaskPayload struct {
	AccountID   int                             `json:"aid"`
	ClickID     string                          `json:"cid"`
	Devices     map[consts.InternalField]string `json:"dev"`
	TaolianType int                             `json:"tt"`
	MappedType  string                          `json:"mt"`
}
