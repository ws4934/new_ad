package issue

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const smtpSettingKey = "smtp"

// SMTPConfig 发件邮箱配置, 存 app_setting(key='smtp')。
// 端口 465 走 SSL 直连, 其余端口在服务器支持时自动 STARTTLS。
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"` // 为空则用 Username
}

// GetSMTP 读取发件配置 (未配置返回零值)。
func GetSMTP(ctx context.Context) (*SMTPConfig, error) {
	cfg := &SMTPConfig{}
	v, err := g.Model(TableSetting).Ctx(ctx).Where("key", smtpSettingKey).Value("value")
	if err != nil || v.IsEmpty() {
		return cfg, err
	}
	err = v.Scan(cfg)
	return cfg, err
}

// SaveSMTP 保存发件配置。Password 为空表示保持原密码。
func SaveSMTP(ctx context.Context, cfg SMTPConfig) error {
	cfg.Host = strings.TrimSpace(cfg.Host)
	cfg.Username = strings.TrimSpace(cfg.Username)
	cfg.From = strings.TrimSpace(cfg.From)
	if cfg.Password == "" {
		old, err := GetSMTP(ctx)
		if err != nil {
			return err
		}
		cfg.Password = old.Password
	}
	_, err := g.Model(TableSetting).Ctx(ctx).Data(g.Map{
		"key": smtpSettingKey, "value": cfg, "updated_at": time.Now(),
	}).OnConflict("key").Save()
	return err
}

// SendMail 用已保存的 SMTP 配置发送纯文本邮件。
func SendMail(ctx context.Context, to []string, subject, body string) error {
	cfg, err := GetSMTP(ctx)
	if err != nil {
		return err
	}
	if cfg.Host == "" || cfg.Port == 0 || cfg.Username == "" {
		return gerror.New("未配置发件邮箱 (SMTP)")
	}
	from := cfg.From
	if from == "" {
		from = cfg.Username
	}
	msg := buildMessage(from, to, subject, body)
	return send(cfg, from, to, msg)
}

func buildMessage(from string, to []string, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + strings.Join(to, ", ") + "\r\n")
	b.WriteString("Subject: " + mime.BEncoding.Encode("UTF-8", subject) + "\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	enc := base64.StdEncoding.EncodeToString([]byte(body))
	for len(enc) > 76 {
		b.WriteString(enc[:76] + "\r\n")
		enc = enc[76:]
	}
	b.WriteString(enc + "\r\n")
	return []byte(b.String())
}

func send(cfg *SMTPConfig, from string, to []string, msg []byte) error {
	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	tlsCfg := &tls.Config{ServerName: cfg.Host}

	var conn net.Conn
	var err error
	if cfg.Port == 465 {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("连接 SMTP 失败: %w", err)
	}
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

	c, err := smtp.NewClient(conn, cfg.Host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()

	if cfg.Port != 465 {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err = c.StartTLS(tlsCfg); err != nil {
				return fmt.Errorf("STARTTLS: %w", err)
			}
		}
	}
	if cfg.Password != "" {
		if err = c.Auth(smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)); err != nil {
			return fmt.Errorf("SMTP 认证失败: %w", err)
		}
	}
	if err = c.Mail(addrOnly(from)); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err = c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("收件人 %s: %w", rcpt, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// addrOnly 从 "名字 <a@b.com>" 中取出邮箱地址。
func addrOnly(s string) string {
	if i := strings.LastIndex(s, "<"); i >= 0 {
		return strings.TrimSuffix(s[i+1:], ">")
	}
	return s
}
