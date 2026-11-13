package mapping

import (
	"context"

	v1 "ad-monitor/api/mapping/v1"
	"ad-monitor/internal/logic/biz"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

// List 行为映射列表
func (c *Controller) List(ctx context.Context, req *v1.MappingListReq) (res *v1.MappingListRes, err error) {
	list, err := biz.MappingList(ctx, req.AccountId)
	if err != nil {
		return nil, err
	}
	return &v1.MappingListRes{List: list}, nil
}

// Save 批量保存行为映射
func (c *Controller) Save(ctx context.Context, req *v1.MappingSaveReq) (res *v1.MappingSaveRes, err error) {
	if err := biz.MappingSave(ctx, req.AccountId, req.Items); err != nil {
		return nil, err
	}
	return &v1.MappingSaveRes{}, nil
}

// Delete 删除行为映射
func (c *Controller) Delete(ctx context.Context, req *v1.MappingDeleteReq) (res *v1.MappingDeleteRes, err error) {
	if err := biz.MappingDelete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.MappingDeleteRes{}, nil
}

// Options 枚举选项
func (c *Controller) Options(ctx context.Context, req *v1.MappingOptionsReq) (res *v1.MappingOptionsRes, err error) {
	return biz.MappingOptions(), nil
}
