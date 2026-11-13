package cache

import (
	"context"
	"encoding/json"

	"ad-monitor/internal/consts"
	"ad-monitor/internal/dao"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/sync/singleflight"
)

// sg 全局 singleflight group, 防止缓存击穿
var sg singleflight.Group

// GetAccount 带 singleflight + Redis 缓存的账户读取
func GetAccount(ctx context.Context, id int) (*entity.AdAccount, error) {
	key := consts.KeyCacheAccount(id)

	// 1. 查 Redis
	v, _ := g.Redis().Get(ctx, key)
	if v != nil && !v.IsNil() && v.String() != "" {
		var acct entity.AdAccount
		if err := json.Unmarshal([]byte(v.String()), &acct); err == nil {
			return &acct, nil
		}
	}

	// 2. singleflight 防击穿
	res, err, _ := sg.Do(key, func() (any, error) {
		var acct entity.AdAccount
		if err := dao.AdAccount.Ctx(ctx).WherePri(id).Scan(&acct); err != nil {
			return nil, err
		}
		data, _ := json.Marshal(&acct)
		_ = g.Redis().SetEX(ctx, key, string(data), 300)
		return &acct, nil
	})
	if err != nil {
		return nil, err
	}
	return res.(*entity.AdAccount), nil
}

// GetPlatform 带缓存的平台读取
func GetPlatform(ctx context.Context, id int) (*entity.AdPlatform, error) {
	key := consts.KeyCachePlatform(id)
	v, _ := g.Redis().Get(ctx, key)
	if v != nil && !v.IsNil() && v.String() != "" {
		var p entity.AdPlatform
		if err := json.Unmarshal([]byte(v.String()), &p); err == nil {
			return &p, nil
		}
	}
	res, err, _ := sg.Do(key, func() (any, error) {
		var p entity.AdPlatform
		if err := dao.AdPlatform.Ctx(ctx).WherePri(id).Scan(&p); err != nil {
			return nil, err
		}
		data, _ := json.Marshal(&p)
		_ = g.Redis().SetEX(ctx, key, string(data), 600)
		return &p, nil
	})
	if err != nil {
		return nil, err
	}
	return res.(*entity.AdPlatform), nil
}

// GetMappings 带缓存的账户行为映射读取
func GetMappings(ctx context.Context, accountID int) ([]*entity.ActionMapping, error) {
	key := consts.KeyCacheMappings(accountID)
	v, _ := g.Redis().Get(ctx, key)
	if v != nil && !v.IsNil() && v.String() != "" {
		var list []*entity.ActionMapping
		if err := json.Unmarshal([]byte(v.String()), &list); err == nil {
			return list, nil
		}
	}
	res, err, _ := sg.Do(key, func() (any, error) {
		var list []*entity.ActionMapping
		if err := dao.ActionMapping.Ctx(ctx).Where(dao.ActionMapping.Columns().AccountId, accountID).Scan(&list); err != nil {
			return nil, err
		}
		data, _ := json.Marshal(&list)
		_ = g.Redis().SetEX(ctx, key, string(data), 300)
		return list, nil
	})
	if err != nil {
		return nil, err
	}
	return res.([]*entity.ActionMapping), nil
}

// InvalidateAccount 失效账户及其映射的缓存
func InvalidateAccount(ctx context.Context, id int) {
	_, _ = g.Redis().Del(ctx, consts.KeyCacheAccount(id))
	_, _ = g.Redis().Del(ctx, consts.KeyCacheMappings(id))
}

// InvalidatePlatform 失效平台缓存
func InvalidatePlatform(ctx context.Context, id int) {
	_, _ = g.Redis().Del(ctx, consts.KeyCachePlatform(id))
}

// WarmupAll 应用启动时预热所有启用的账户和平台
func WarmupAll(ctx context.Context) {
	var accounts []*entity.AdAccount
	if err := dao.AdAccount.Ctx(ctx).Where(dao.AdAccount.Columns().Status, 1).Scan(&accounts); err != nil {
		g.Log().Warningf(ctx, "warmup accounts failed: %v", err)
	} else {
		for _, a := range accounts {
			data, _ := json.Marshal(a)
			_ = g.Redis().SetEX(ctx, consts.KeyCacheAccount(a.Id), string(data), 300)
		}
	}

	var platforms []*entity.AdPlatform
	if err := dao.AdPlatform.Ctx(ctx).Scan(&platforms); err != nil {
		g.Log().Warningf(ctx, "warmup platforms failed: %v", err)
	} else {
		for _, p := range platforms {
			data, _ := json.Marshal(p)
			_ = g.Redis().SetEX(ctx, consts.KeyCachePlatform(p.Id), string(data), 600)
		}
	}

	g.Log().Infof(ctx, "warmup done: %d accounts, %d platforms", len(accounts), len(platforms))
}
