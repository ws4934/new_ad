package biz

import (
	"context"

	"ad-monitor/internal/dao"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

var pc = dao.AdPlatform.Columns()

// PlatformList 查询所有平台
func PlatformList(ctx context.Context) ([]*entity.AdPlatform, error) {
	var list []*entity.AdPlatform
	err := dao.AdPlatform.Ctx(ctx).OrderAsc(pc.Id).Scan(&list)
	return list, err
}

// PlatformCreate 创建平台
func PlatformCreate(ctx context.Context, name, displayName, config string) (int, error) {
	id, err := dao.AdPlatform.Ctx(ctx).Data(g.Map{
		pc.Name:        name,
		pc.DisplayName: displayName,
		pc.Config:      jsonOrDefault(config, "{}"),
		pc.Status:      1,
	}).InsertAndGetId()
	return int(id), err
}

// PlatformUpdate 更新平台
func PlatformUpdate(ctx context.Context, id int, displayName, config string, status *int) error {
	data := g.Map{}
	if displayName != "" {
		data[pc.DisplayName] = displayName
	}
	if config != "" {
		data[pc.Config] = config
	}
	if status != nil {
		data[pc.Status] = *status
	}
	if len(data) == 0 {
		return nil
	}
	_, err := dao.AdPlatform.Ctx(ctx).WherePri(id).Data(data).Update()
	if err == nil {
		cache.InvalidatePlatform(ctx, id)
	}
	return err
}
