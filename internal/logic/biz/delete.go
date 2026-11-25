package biz

import (
	"context"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/dao"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// deleteAccounts 在调用方事务中删除账户及依赖；账户行锁阻止并发新增关联数据。
func deleteAccounts(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	if _, err := dao.ReportStats.Ctx(ctx).WhereIn(dao.ReportStats.Columns().AccountId, ids).Delete(); err != nil {
		return err
	}
	if _, err := dao.ActionMapping.Ctx(ctx).WhereIn(mc.AccountId, ids).Delete(); err != nil {
		return err
	}
	_, err := dao.AdAccount.Ctx(ctx).WhereIn(ac.Id, ids).Delete()
	return err
}

func AccountDelete(ctx context.Context, id int) error {
	if id <= 0 {
		return gerror.New("账户 ID 必须为正整数")
	}
	err := dao.AdAccount.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var account *entity.AdAccount
		if err := dao.AdAccount.Ctx(ctx).WherePri(id).LockUpdate().Scan(&account); err != nil {
			return err
		}
		// 删除保持幂等，允许缓存清理失败后重试。
		if account == nil {
			return nil
		}
		return deleteAccounts(ctx, []int{id})
	})
	if err != nil {
		return err
	}
	return clearDeletedCaches(ctx, []int{id}, 0)
}

func PlatformDelete(ctx context.Context, id int) error {
	if id <= 0 {
		return gerror.New("平台 ID 必须为正整数")
	}
	var ids []int
	err := dao.AdPlatform.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var platform *entity.AdPlatform
		if err := dao.AdPlatform.Ctx(ctx).WherePri(id).LockUpdate().Scan(&platform); err != nil {
			return err
		}
		if platform == nil {
			return nil
		}
		var accounts []*entity.AdAccount
		if err := dao.AdAccount.Ctx(ctx).Where(ac.PlatformId, id).OrderAsc(ac.Id).LockUpdate().Scan(&accounts); err != nil {
			return err
		}
		for _, account := range accounts {
			ids = append(ids, account.Id)
		}
		if err := deleteAccounts(ctx, ids); err != nil {
			return err
		}
		if _, err := dao.ReportStats.Ctx(ctx).Where(dao.ReportStats.Columns().PlatformId, id).Delete(); err != nil {
			return err
		}
		if _, err := dao.PlatformCsvStats.Ctx(ctx).Where(dao.PlatformCsvStats.Columns().PlatformId, id).Delete(); err != nil {
			return err
		}
		// 在提交前清理：Redis 失败则回滚，重试仍可取得完整账户列表。
		if err := clearDeletedCaches(ctx, ids, id); err != nil {
			return err
		}
		_, err := dao.AdPlatform.Ctx(ctx).WherePri(id).Delete()
		return err
	})
	if err != nil {
		return err
	}
	return clearDeletedCaches(ctx, ids, id)
}

func clearDeletedCaches(ctx context.Context, accountIDs []int, platformID int) error {
	keys := make([]string, 0, len(accountIDs)*3+1)
	for _, id := range accountIDs {
		keys = append(keys, consts.KeyCacheAccount(id), consts.KeyCacheMappings(id), consts.KeyReqLog(id))
	}
	if platformID > 0 {
		keys = append(keys, consts.KeyCachePlatform(platformID))
	}
	if len(keys) == 0 {
		return nil
	}
	if _, err := g.Redis().Del(ctx, keys...); err != nil {
		return gerror.Wrap(err, "删除缓存清理失败，请重试")
	}
	return nil
}
