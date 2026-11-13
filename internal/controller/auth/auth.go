package auth

import (
	"context"

	v1 "ad-monitor/api/auth/v1"
	authlogic "ad-monitor/internal/logic/auth"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct{}

func New() *Controller { return &Controller{} }

// Login 后台登录 (固定密码验证, 返回 JWT)
// IP 级防爆破: 10 分钟内失败 5 次则临时封禁
func (c *Controller) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	ip := g.RequestFromCtx(ctx).GetClientIp()
	if authlogic.IsLoginBlocked(ctx, ip) {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "失败次数过多, 请 10 分钟后再试")
	}
	if !authlogic.CheckPassword(ctx, req.Password) {
		authlogic.RecordLoginFail(ctx, ip)
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "密码错误")
	}
	authlogic.ClearLoginFail(ctx, ip)
	token, err := authlogic.GenerateToken()
	if err != nil {
		return nil, err
	}
	return &v1.LoginRes{Token: token}, nil
}
