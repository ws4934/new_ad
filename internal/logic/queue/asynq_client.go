package queue

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/model/bo"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hibiken/asynq"
)

var (
	clientOnce sync.Once
	client     *asynq.Client
)

// redisOpt 从 config 读取 asynq Redis 连接参数
// config 路径: adMonitor.queue.redisAddr / adMonitor.queue.redisDb
func redisOpt(ctx context.Context) asynq.RedisClientOpt {
	addr := g.Cfg().MustGet(ctx, "adMonitor.queue.redisAddr").String()
	db := g.Cfg().MustGet(ctx, "adMonitor.queue.redisDb").Int()
	return asynq.RedisClientOpt{Addr: addr, DB: db}
}

// GetClient 单例 asynq 客户端
func GetClient(ctx context.Context) *asynq.Client {
	clientOnce.Do(func() {
		client = asynq.NewClient(redisOpt(ctx))
	})
	return client
}

// EnqueueTaolianReport 入队淘联上报任务 (优先级 default 队列)
func EnqueueTaolianReport(ctx context.Context, p *bo.TaolianTaskPayload) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	task := asynq.NewTask(consts.TaskTaolianReport, data)
	_, err = GetClient(ctx).EnqueueContext(ctx, task,
		asynq.Queue(consts.QueueDefault),
		asynq.MaxRetry(3),
		asynq.Retention(24*time.Hour),
	)
	return err
}

// EnqueueCallbackReport 入队真实回传任务
func EnqueueCallbackReport(ctx context.Context, p *bo.CallbackTaskPayload) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	task := asynq.NewTask(consts.TaskCallbackReport, data)
	_, err = GetClient(ctx).EnqueueContext(ctx, task,
		asynq.Queue(consts.QueueCallback),
		asynq.MaxRetry(3),
		asynq.Retention(24*time.Hour),
	)
	return err
}

// EnqueueFakeReport 入队虚假回传任务 (延迟执行)
func EnqueueFakeReport(ctx context.Context, p *bo.FakeTaskPayload, delay time.Duration) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	task := asynq.NewTask(consts.TaskFakeReport, data)
	_, err = GetClient(ctx).EnqueueContext(ctx, task,
		asynq.Queue(consts.QueueFake),
		asynq.ProcessIn(delay),
		asynq.MaxRetry(3),
		asynq.Retention(24*time.Hour),
	)
	return err
}
