package issue

import (
	"context"

	v1 "ad-monitor/api/issue/v1"
	"ad-monitor/internal/logic/issue"
)

type Controller struct{}

func New() *Controller { return &Controller{} }

func (c *Controller) List(ctx context.Context, req *v1.ListReq) (*v1.ListRes, error) {
	list, err := issue.List(ctx)
	return &v1.ListRes{List: list}, err
}

func (c *Controller) Create(ctx context.Context, req *v1.CreateReq) (*v1.CreateRes, error) {
	id, err := issue.Create(ctx, req.Input())
	return &v1.CreateRes{Id: id}, err
}

func (c *Controller) Update(ctx context.Context, req *v1.UpdateReq) (*v1.UpdateRes, error) {
	return &v1.UpdateRes{}, issue.Update(ctx, req.Id, req.Input())
}

func (c *Controller) Delete(ctx context.Context, req *v1.DeleteReq) (*v1.DeleteRes, error) {
	return &v1.DeleteRes{}, issue.Delete(ctx, req.Id)
}

func (c *Controller) Check(ctx context.Context, req *v1.CheckReq) (*v1.CheckRes, error) {
	m, err := issue.CheckNow(ctx, req.Id)
	return &v1.CheckRes{Monitor: m}, err
}

func (c *Controller) Logs(ctx context.Context, req *v1.LogsReq) (*v1.LogsRes, error) {
	list, err := issue.Logs(ctx, req.Id, req.Limit)
	return &v1.LogsRes{List: list}, err
}

func (c *Controller) SmtpGet(ctx context.Context, req *v1.SmtpGetReq) (*v1.SmtpGetRes, error) {
	cfg, err := issue.GetSMTP(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.SmtpGetRes{
		Host: cfg.Host, Port: cfg.Port, Username: cfg.Username, From: cfg.From, PasswordSet: cfg.Password != "",
	}, nil
}

func (c *Controller) SmtpSave(ctx context.Context, req *v1.SmtpSaveReq) (*v1.SmtpSaveRes, error) {
	return &v1.SmtpSaveRes{}, issue.SaveSMTP(ctx, issue.SMTPConfig{
		Host: req.Host, Port: req.Port, Username: req.Username, Password: req.Password, From: req.From,
	})
}

func (c *Controller) SmtpTest(ctx context.Context, req *v1.SmtpTestReq) (*v1.SmtpTestRes, error) {
	return &v1.SmtpTestRes{}, issue.SendMail(ctx, []string{req.To}, "[直发提醒] 测试邮件", "SMTP 配置可用, 直发状态变化时会发送到此邮箱。")
}
