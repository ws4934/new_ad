package auth

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

// MiddlewareJWT 从 Authorization: Bearer xxx 提取 token 并校验
// 校验失败返回 401, 不进入下一个 handler
func MiddlewareJWT(r *ghttp.Request) {
	header := r.GetHeader("Authorization")
	if header == "" {
		r.Response.WriteStatus(http.StatusUnauthorized)
		r.Response.WriteJsonExit(map[string]any{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	token := strings.TrimPrefix(header, "Bearer ")
	if token == header || token == "" {
		r.Response.WriteStatus(http.StatusUnauthorized)
		r.Response.WriteJsonExit(map[string]any{
			"code":    401,
			"message": "无效的 Authorization 格式",
		})
		return
	}

	if !ValidateToken(token) {
		r.Response.WriteStatus(http.StatusUnauthorized)
		r.Response.WriteJsonExit(map[string]any{
			"code":    401,
			"message": "token 无效或已过期",
		})
		return
	}

	r.Middleware.Next()
}
