package queue

import (
	"context"

	"ad-monitor/internal/consts"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hibiken/asynq"
)

// StartServer 启动 asynq 消费者 server, 返回实例用于优雅关停
func StartServer(ctx context.Context) *asynq.Server {
	cfg := g.Cfg()
	concurrency := cfg.MustGet(ctx, "adMonitor.queue.concurrency").Int()
	qDefault := cfg.MustGet(ctx, "adMonitor.queue.queues.default").Int()
	qCallback := cfg.MustGet(ctx, "adMonitor.queue.queues.callback").Int()
	qFake := cfg.MustGet(ctx, "adMonitor.queue.queues.fake").Int()

	srv := asynq.NewServer(
		redisOpt(ctx),
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				consts.QueueDefault:  qDefault,
				consts.QueueCallback: qCallback,
				consts.QueueFake:     qFake,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(consts.TaskTaolianReport, HandleTaolianReport)
	mux.HandleFunc(consts.TaskCallbackReport, HandleCallbackReport)
	mux.HandleFunc(consts.TaskFakeReport, HandleFakeReport)
	mux.HandleFunc(consts.TaskArchiveDaily, HandleArchiveDaily)

	go func() {
		if err := srv.Run(mux); err != nil {
			g.Log().Fatalf(ctx, "asynq server run failed: %v", err)
		}
	}()
	g.Log().Infof(ctx, "asynq server started (concurrency=%d)", concurrency)
	return srv
}

// StartScheduler 启动 asynq 定时任务调度器 (每日归档), 返回实例用于优雅关停
func StartScheduler(ctx context.Context) *asynq.Scheduler {
	scheduler := asynq.NewScheduler(redisOpt(ctx), &asynq.SchedulerOpts{})

	task := asynq.NewTask(consts.TaskArchiveDaily, nil)
	if _, err := scheduler.Register("5 0 * * *", task, asynq.Queue(consts.QueueDefault)); err != nil {
		g.Log().Errorf(ctx, "register scheduler task failed: %v", err)
		return nil
	}

	go func() {
		if err := scheduler.Run(); err != nil {
			g.Log().Fatalf(ctx, "asynq scheduler run failed: %v", err)
		}
	}()
	g.Log().Infof(ctx, "asynq scheduler started")
	return scheduler
}
