package auth

import (
	"context"
	"crypto/subtle"
	"sync"
	"time"

	"ad-monitor/internal/consts"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret     []byte
	jwtSecretOnce sync.Once
)

// getJWTSecret 从配置读取 JWT 密钥, 仅初始化一次
// 配置路径: adMonitor.jwtSecret, 未配置时回退到 adminPassword 派生值
func getJWTSecret() []byte {
	jwtSecretOnce.Do(func() {
		ctx := context.Background()
		secret := g.Cfg().MustGet(ctx, "adMonitor.jwtSecret").String()
		if secret == "" {
			// 回退: 用 adminPassword 拼接固定前缀, 避免裸密码直接当密钥
			pwd := g.Cfg().MustGet(ctx, "adMonitor.adminPassword").String()
			if pwd == "" {
				// 两者皆空时密钥是公开常量, 任何人可离线伪造合法 token
				// (CheckPassword 拒绝登录, 但伪造的 token 能直接过 ValidateToken)
				panic("adMonitor.jwtSecret 与 adminPassword 均未配置, 拒绝以可伪造的常量密钥启动")
			}
			secret = "ad-monitor-jwt:" + pwd
		}
		jwtSecret = []byte(secret)
	})
	return jwtSecret
}

// EnsureSecretConfigured 启动时调用, 提前触发密钥初始化
// 把"配置缺失"暴露在启动阶段而不是第一个登录请求
func EnsureSecretConfigured() {
	_ = getJWTSecret()
}

// CheckPassword 校验后台管理密码 (常数时间比较, 防时序侧信道)
func CheckPassword(ctx context.Context, password string) bool {
	configured := g.Cfg().MustGet(ctx, "adMonitor.adminPassword").String()
	if configured == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(password), []byte(configured)) == 1
}

// --- 登录防爆破: IP 级失败计数, 10 分钟内失败 5 次则封禁 ---

const (
	loginFailLimit = 5
	loginFailTTL   = 600 // 秒
)

// IsLoginBlocked 该 IP 是否已被临时封禁
func IsLoginBlocked(ctx context.Context, ip string) bool {
	v, err := g.Redis().Get(ctx, consts.KeyLoginFail(ip))
	if err != nil || v == nil {
		return false
	}
	return v.Int() >= loginFailLimit
}

// RecordLoginFail 记录一次登录失败, 每次失败都刷新 TTL
func RecordLoginFail(ctx context.Context, ip string) {
	key := consts.KeyLoginFail(ip)
	if _, err := g.Redis().Incr(ctx, key); err != nil {
		return
	}
	if _, err := g.Redis().Expire(ctx, key, loginFailTTL); err != nil {
		// Expire 失败会让无 TTL 的计数器永久封禁该 IP, fail-open: 直接删掉
		_, _ = g.Redis().Del(ctx, key)
	}
}

// ClearLoginFail 登录成功后清除失败计数
func ClearLoginFail(ctx context.Context, ip string) {
	_, _ = g.Redis().Del(ctx, consts.KeyLoginFail(ip))
}

// GenerateToken 生成 JWT (24 小时有效)
func GenerateToken() (string, error) {
	claims := jwt.MapClaims{
		"role": "admin",
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ValidateToken 校验 JWT, 返回是否有效
// WithValidMethods 显式限定 HS256, 防算法混淆类攻击
func ValidateToken(tokenStr string) bool {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return getJWTSecret(), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return false
	}
	return token.Valid
}
