package platform

import (
	"context"

	v1 "ad-monitor/api/platform/v1"
	"ad-monitor/internal/logic/biz"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

// List 平台列表
func (c *Controller) List(ctx context.Context, req *v1.PlatformListReq) (res *v1.PlatformListRes, err error) {
	list, err := biz.PlatformList(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.PlatformListRes{List: list}, nil
}

// Create 创建平台
func (c *Controller) Create(ctx context.Context, req *v1.PlatformCreateReq) (res *v1.PlatformCreateRes, err error) {
	id, err := biz.PlatformCreate(ctx, req.Name, req.DisplayName, req.Config)
	if err != nil {
		return nil, err
	}
	return &v1.PlatformCreateRes{Id: id}, nil
}

// Update 更新平台
func (c *Controller) Update(ctx context.Context, req *v1.PlatformUpdateReq) (res *v1.PlatformUpdateRes, err error) {
	if err := biz.PlatformUpdate(ctx, req.Id, req.DisplayName, req.Config, req.Status); err != nil {
		return nil, err
	}
	return &v1.PlatformUpdateRes{}, nil
}

func (c *Controller) Delete(ctx context.Context, req *v1.PlatformDeleteReq) (*v1.PlatformDeleteRes, error) {
	if err := biz.PlatformDelete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.PlatformDeleteRes{}, nil
}
