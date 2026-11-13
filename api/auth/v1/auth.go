package v1

import "github.com/gogf/gf/v2/frame/g"

type LoginReq struct {
	g.Meta   `path:"/api/v2/login" method:"post" tags:"Auth" summary:"后台登录"`
	Password string `json:"password" v:"required"`
}

type LoginRes struct {
	Token string `json:"token"`
}
