package consts

// asynq 任务类型常量
const (
	TaskTaolianReport  = "taolian:report"  // 淘联点击上报
	TaskCallbackReport = "callback:report" // 广告平台真实回传
	TaskFakeReport     = "fake:report"     // 广告平台虚假回传
	TaskArchiveDaily   = "archive:daily"   // 每日归档定时任务
)

// asynq 队列名
const (
	QueueDefault  = "default"  // 淘联上报
	QueueCallback = "callback" // 广告平台真实回传
	QueueFake     = "fake"     // 虚假回传
)
