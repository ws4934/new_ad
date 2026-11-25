package biz

import (
	"context"

	v1 "ad-monitor/api/mapping/v1"
	"ad-monitor/internal/dao"
	"ad-monitor/internal/logic/cache"
	"ad-monitor/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

var mc = dao.ActionMapping.Columns()

// MappingList 查询指定账户的行为映射
func MappingList(ctx context.Context, accountId int) ([]*entity.ActionMapping, error) {
	var list []*entity.ActionMapping
	err := dao.ActionMapping.Ctx(ctx).Where(mc.AccountId, accountId).OrderAsc(mc.Id).Scan(&list)
	return list, err
}

// MappingSave 批量保存行为映射 (事务内先删后插, 批量插入)
func MappingSave(ctx context.Context, accountId int, items []*v1.MappingSaveItem) error {
	err := dao.ActionMapping.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.ActionMapping.Ctx(ctx).Where(mc.AccountId, accountId).Delete(); err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		rows := make([]g.Map, 0, len(items))
		for _, item := range items {
			rows = append(rows, g.Map{
				mc.AccountId:    accountId,
				mc.TaolianType:  item.TaolianType,
				mc.TaolianDesc:  item.TaolianDesc,
				mc.PlatformType: item.PlatformType,
				mc.PlatformDesc: item.PlatformDesc,
				mc.Enabled:      item.Enabled,
			})
		}
		_, err := dao.ActionMapping.Ctx(ctx).Data(rows).Insert()
		return err
	})
	if err != nil {
		return err
	}
	cache.InvalidateAccount(ctx, accountId)
	return nil
}

// MappingDelete 删除单条映射
func MappingDelete(ctx context.Context, id int) error {
	var m entity.ActionMapping
	if err := dao.ActionMapping.Ctx(ctx).WherePri(id).Scan(&m); err != nil {
		return err
	}
	_, err := dao.ActionMapping.Ctx(ctx).WherePri(id).Delete()
	if err == nil {
		cache.InvalidateAccount(ctx, m.AccountId)
	}
	return err
}

// MappingOptions 返回硬编码的枚举选项
func MappingOptions() *v1.MappingOptionsRes {
	return &v1.MappingOptionsRes{
		TaolianTypes: []*v1.TypeOption{
			{Type: 5, Desc: "落地页访问"},
			{Type: 11, Desc: "淘客订单支付"},
			{Type: 28, Desc: "淘客商品点击"},
			{Type: 36, Desc: "会场首登"},
			{Type: 37, Desc: "红包领取"},
			{Type: 48, Desc: "取消淘客订单支付"},
			{Type: 49, Desc: "高佣订单支付"},
			{Type: 76, Desc: "飞猪-酒店日历房-订单回传"},
			{Type: 77, Desc: "飞猪-酒店日历房-新客订单回传"},
		},
		PlatformTypes: map[string][]*v1.TypeOption{
			"weibo": {
				{Type: "1", Desc: "激活"},
				{Type: "2", Desc: "下单购买"},
				{Type: "3", Desc: "注册"},
				{Type: "4", Desc: "付费"},
				{Type: "7", Desc: "次日留存"},
				{Type: "8", Desc: "app内访问"},
				{Type: "9", Desc: "3日留存"},
				{Type: "10", Desc: "7日留存"},
			},
			// 直达回传值以微博后台「添加事件」页展示为准, 文档示例 behavior=2001
			"weibo_direct": {
				{Type: "2001", Desc: "付款订单数"},
			},
			"xiaohongshu": {
				{Type: "401", Desc: "APP打开"},
				{Type: "402", Desc: "APP进店"},
				{Type: "403", Desc: "APP互动"},
				{Type: "404", Desc: "APP支付"},
			},
		},
	}
}
