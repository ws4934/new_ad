package account

import (
	"context"

	v1 "ad-monitor/api/account/v1"
	"ad-monitor/internal/logic/biz"
	"ad-monitor/internal/logic/stats"

	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct{}

func New() *Controller {
	return &Controller{}
}

// List 账户列表 (带今日统计)
func (c *Controller) List(ctx context.Context, req *v1.AccountListReq) (res *v1.AccountListRes, err error) {
	list, err := biz.AccountList(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.AccountListRes{List: list}, nil
}

// Create 创建账户
func (c *Controller) Create(ctx context.Context, req *v1.AccountCreateReq) (res *v1.AccountCreateRes, err error) {
	id, err := biz.AccountCreate(ctx, req)
	if err != nil {
		return nil, err
	}
	return &v1.AccountCreateRes{Id: id}, nil
}

// Update 更新账户
func (c *Controller) Update(ctx context.Context, req *v1.AccountUpdateReq) (res *v1.AccountUpdateRes, err error) {
	if err := biz.AccountUpdate(ctx, req); err != nil {
		return nil, err
	}
	return &v1.AccountUpdateRes{}, nil
}

// MonitorURL 生成监测链接
func (c *Controller) MonitorURL(ctx context.Context, req *v1.AccountMonitorURLReq) (res *v1.AccountMonitorURLRes, err error) {
	r := g.RequestFromCtx(ctx)
	baseURL := r.GetSchema() + "://" + r.Host
	url, err := biz.AccountMonitorURL(ctx, req.Id, baseURL)
	if err != nil {
		return nil, err
	}
	return &v1.AccountMonitorURLRes{URL: url}, nil
}

// Requests 最近 50 条监测请求
func (c *Controller) Requests(ctx context.Context, req *v1.AccountRequestsReq) (*v1.AccountRequestsRes, error) {
	list, err := stats.RequestLog(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.AccountRequestsRes{List: list}, nil
}

func (c *Controller) Delete(ctx context.Context, req *v1.AccountDeleteReq) (*v1.AccountDeleteRes, error) {
	if err := biz.AccountDelete(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.AccountDeleteRes{}, nil
}
