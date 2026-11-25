package biz

import (
	"context"

	v1 "ad-monitor/api/account/v1"
	"ad-monitor/internal/consts"
	"ad-monitor/internal/dao"
	"ad-monitor/internal/logic/adapter"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/logic/stats"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

var ac = dao.AdAccount.Columns()

// jsonOrDefault JSONB 字段空值保护: 空字符串转为指定默认值
func jsonOrDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

// AccountList 查询所有账户, 并批量获取今日统计
func AccountList(ctx context.Context) ([]*v1.AccountItem, error) {
	var accounts []*entity.AdAccount
	if err := dao.AdAccount.Ctx(ctx).OrderAsc(ac.Id).Scan(&accounts); err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return []*v1.AccountItem{}, nil
	}

	// 批量加载平台名称
	platformNames := make(map[int]string)
	var platforms []*entity.AdPlatform
	if err := dao.AdPlatform.Ctx(ctx).Scan(&platforms); err == nil {
		for _, p := range platforms {
			platformNames[p.Id] = p.Name
		}
	}

	// 构造 Redis keys, 批量 MGet 今日计数
	day := stats.Today()
	keys := make([]string, 0, len(accounts)*6)
	for _, a := range accounts {
		keys = append(keys,
			consts.KeyStatsClick(a.Id, day),
			consts.KeyStatsTaolianOk(a.Id, day),
			consts.KeyStatsCallbackOk(a.Id, day),
			consts.KeyStatsFakeOk(a.Id, day),
			consts.KeyStatsDebugOk(a.Id, day),
			consts.KeyStatsDup(a.Id, day),
		)
	}

	counters, err := g.Redis().MGet(ctx, keys...)
	if err != nil {
		g.Log().Warningf(ctx, "mget counters failed: %v", err)
	}

	items := make([]*v1.AccountItem, 0, len(accounts))
	for i, a := range accounts {
		item := &v1.AccountItem{
			Id:            a.Id,
			PlatformId:    a.PlatformId,
			PlatformName:  platformNames[a.PlatformId],
			Name:          a.Name,
			Credentials:   a.Credentials,
			TaolianConfig: a.TaolianConfig,
			FakeEnabled:   a.FakeEnabled,
			FakeRate:      a.FakeRate,
			FakeTypes:     a.FakeTypes,
			FakeDelayMin:  a.FakeDelayMin,
			FakeDelayMax:  a.FakeDelayMax,
			DebugMode:     a.DebugMode,
			DebugEvents:   a.DebugEvents,
			Status:        a.Status,
			CreatedAt:     a.CreatedAt,
			UpdatedAt:     a.UpdatedAt,
		}
		if counters != nil {
			base := i * 6
			if v, ok := counters[keys[base]]; ok && v != nil {
				item.TodayClick = v.Int()
			}
			if v, ok := counters[keys[base+1]]; ok && v != nil {
				item.TodayTaolian = v.Int()
			}
			if v, ok := counters[keys[base+2]]; ok && v != nil {
				item.TodayCallback = v.Int()
			}
			if v, ok := counters[keys[base+3]]; ok && v != nil {
				item.TodayFake = v.Int()
			}
			if v, ok := counters[keys[base+4]]; ok && v != nil {
				item.TodayDebug = v.Int()
			}
			if v, ok := counters[keys[base+5]]; ok && v != nil {
				item.TodayDup = v.Int()
			}
		}
		items = append(items, item)
	}
	return items, nil
}

// AccountCreate 创建账户
func AccountCreate(ctx context.Context, req *v1.AccountCreateReq) (int, error) {
	id, err := dao.AdAccount.Ctx(ctx).Data(g.Map{
		ac.PlatformId:    req.PlatformId,
		ac.Name:          req.Name,
		ac.Credentials:   jsonOrDefault(req.Credentials, "{}"),
		ac.TaolianConfig: jsonOrDefault(req.TaolianConfig, "{}"),
		ac.FakeEnabled:   req.FakeEnabled,
		ac.FakeRate:      req.FakeRate,
		ac.FakeTypes:     jsonOrDefault(req.FakeTypes, "[]"),
		ac.FakeDelayMin:  req.FakeDelayMin,
		ac.FakeDelayMax:  req.FakeDelayMax,
		ac.DebugMode:     req.DebugMode,
		ac.DebugEvents:   jsonOrDefault(req.DebugEvents, "[]"),
		ac.Status:        1,
	}).InsertAndGetId()
	return int(id), err
}

// AccountUpdate 更新账户
func AccountUpdate(ctx context.Context, req *v1.AccountUpdateReq) error {
	data := g.Map{}
	if req.Name != "" {
		data[ac.Name] = req.Name
	}
	if req.Credentials != "" {
		data[ac.Credentials] = req.Credentials
	}
	if req.TaolianConfig != "" {
		data[ac.TaolianConfig] = req.TaolianConfig
	}
	if req.FakeEnabled != nil {
		data[ac.FakeEnabled] = *req.FakeEnabled
	}
	if req.FakeRate != nil {
		data[ac.FakeRate] = *req.FakeRate
	}
	if req.FakeTypes != "" {
		data[ac.FakeTypes] = req.FakeTypes
	}
	if req.FakeDelayMin != nil {
		data[ac.FakeDelayMin] = *req.FakeDelayMin
	}
	if req.FakeDelayMax != nil {
		data[ac.FakeDelayMax] = *req.FakeDelayMax
	}
	if req.DebugMode != nil {
		data[ac.DebugMode] = *req.DebugMode
	}
	if req.DebugEvents != "" {
		data[ac.DebugEvents] = req.DebugEvents
	}
	if req.Status != nil {
		data[ac.Status] = *req.Status
	}
	if len(data) == 0 {
		return nil
	}
	_, err := dao.AdAccount.Ctx(ctx).WherePri(req.Id).Data(data).Update()
	if err == nil {
		cache.InvalidateAccount(ctx, req.Id)
	}
	return err
}

// AccountMonitorURL 生成监测链接
func AccountMonitorURL(ctx context.Context, accountID int, baseURL string) (string, error) {
	acct, err := cache.GetAccount(ctx, accountID)
	if err != nil {
		return "", err
	}
	plat, err := cache.GetPlatform(ctx, acct.PlatformId)
	if err != nil {
		return "", err
	}
	ad := adapter.Get(plat.Name)
	if ad == nil {
		return "", gerror.New("adapter not found: " + plat.Name)
	}
	url := adapter.BuildMonitorURL(baseURL, accountID, ad.Spec())
	return url, nil
}
