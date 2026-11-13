package system

import (
	"context"

	v1 "ad-monitor/api/system/v1"
	"ad-monitor/internal/logic/updater"
)

type Controller struct{}

func New() *Controller { return &Controller{} }

func (c *Controller) UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (*v1.UpdateStatusRes, error) {
	return &v1.UpdateStatusRes{Status: updater.GetStatus()}, nil
}

func (c *Controller) UpdateCheck(ctx context.Context, req *v1.UpdateCheckReq) (*v1.UpdateCheckRes, error) {
	s, err := updater.Check(ctx, true)
	return &v1.UpdateCheckRes{Status: s}, err
}

func (c *Controller) UpdateStart(ctx context.Context, req *v1.UpdateStartReq) (*v1.UpdateStartRes, error) {
	s, err := updater.Start(ctx, updater.SelfTerminate)
	return &v1.UpdateStartRes{Status: s}, err
}
