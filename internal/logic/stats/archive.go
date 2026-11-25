package stats

import (
	"context"
	"time"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/dao"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// ArchiveYesterday 每日归档任务: 把前一天的 Redis 计数器归档到 PG 的 report_stats + platform_csv_stats 表。
// 由 asynq scheduler 在每日 00:05 触发调用。
func ArchiveYesterday(ctx context.Context) error {
	yesterday := time.Now().AddDate(0, 0, -1)
	day := yesterday.Format("20060102")      // Redis key 用的日期格式
	statDate := yesterday.Format("2006-01-02") // 数据库 stat_date 用的日期格式

	rc := dao.ReportStats.Columns()

	// ── 1. 账户维度归档: 遍历所有 account, 读 12 个 Redis 计数器, 写 report_stats ──
	var accts []*entity.AdAccount
	if err := dao.AdAccount.Ctx(ctx).Scan(&accts); err != nil {
		return err
	}
	for _, acct := range accts {
		aid := acct.Id
		row := g.Map{
			rc.AccountId:  aid,
			rc.PlatformId: acct.PlatformId,
			rc.StatDate:   statDate,
		}
		// 读 12 个计数器
		keys := map[string]string{
			rc.ClickCount:  consts.KeyStatsClick(aid, day),
			rc.TaolianOk:   consts.KeyStatsTaolianOk(aid, day),
			rc.TaolianErr:  consts.KeyStatsTaolianErr(aid, day),
			rc.CallbackOk:  consts.KeyStatsCallbackOk(aid, day),
			rc.CallbackErr: consts.KeyStatsCallbackErr(aid, day),
			rc.FakeOk:      consts.KeyStatsFakeOk(aid, day),
			rc.FakeErr:     consts.KeyStatsFakeErr(aid, day),
			rc.DropCount:   consts.KeyStatsDrop(aid, day),
			rc.DebugOk:     consts.KeyStatsDebugOk(aid, day),
			rc.DebugErr:    consts.KeyStatsDebugErr(aid, day),
			rc.CbRecv:      consts.KeyStatsCbRecv(aid, day),
			rc.CbDrop:      consts.KeyStatsCbDrop(aid, day),
		}
		hasAny := false
		readErr := false
		var redisKeys []string
		for col, k := range keys {
			v, err := g.Redis().Get(ctx, k)
			if err != nil {
				// Redis 读失败不能当 0: 否则会把残缺数据写库, 然后删掉没读到的 key
				g.Log().Errorf(ctx, "archive: read %s failed: %v", k, err)
				readErr = true
				break
			}
			if v != nil && !v.IsNil() {
				n := v.Int()
				row[col] = n
				if n > 0 {
					hasAny = true
				}
			} else {
				row[col] = 0
			}
			redisKeys = append(redisKeys, k)
		}
		if readErr {
			continue // 跳过该账户: 不写库也不删 key, key 有 35 天 TTL, 留给下次重试
		}
		if !hasAny {
			continue // 全为 0 跳过, 不写数据库
		}
		// PG 下 Save 需要显式指定冲突列 (表主键是自增 id, 业务唯一键是 account_id+platform_id+stat_date)
		_, err := dao.ReportStats.Ctx(ctx).
			Data(row).
			OnConflict(rc.AccountId, rc.PlatformId, rc.StatDate).
			Save()
		if err != nil {
			g.Log().Errorf(ctx, "archive report_stats failed: aid=%d err=%v", aid, err)
			continue // 写库失败不删 key, 留给下次重试
		}
		// 归档成功后清理昨天的 Redis key, 避免内存浪费和重复归档
		for _, k := range redisKeys {
			_, _ = g.Redis().Del(ctx, k)
		}
	}

	// CSV 新增统计不再归档 (去重改为本地 bloom filter, 没有 Redis bitmap 可读)
	// CSV 文件行数即去重后的设备号数, 前端直接展示文件大小

	g.Log().Infof(ctx, "archive done: day=%s accounts=%d", day, len(accts))
	return nil
}
