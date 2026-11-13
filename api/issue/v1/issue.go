package v1

import (
	"ad-monitor/internal/logic/issue"

	"github.com/gogf/gf/v2/frame/g"
)

type MonitorForm struct {
	Name             string `json:"name"             v:"required"`
	AppKey           string `json:"appKey"           v:"required"`
	AppSecret        string `json:"appSecret"`
	ActivityCategory int    `json:"activityCategory" v:"required|in:1,2,4"`
	IntervalSec      int    `json:"intervalSec"`
	NotifyEmails     string `json:"notifyEmails"`
	Enabled          bool   `json:"enabled"`
}

func (f MonitorForm) Input() issue.Input {
	return issue.Input{
		Name: f.Name, AppKey: f.AppKey, AppSecret: f.AppSecret, ActivityCategory: f.ActivityCategory,
		IntervalSec: f.IntervalSec, NotifyEmails: f.NotifyEmails, Enabled: f.Enabled,
	}
}

type ListReq struct {
	g.Meta `path:"/api/v2/issue-monitors" tags:"Issue" method:"get" summary:"直发查询配置列表(含最新结果)"`
}
type ListRes struct {
	List []*issue.Monitor `json:"list"`
}

type CreateReq struct {
	g.Meta `path:"/api/v2/issue-monitors" tags:"Issue" method:"post" summary:"新增直发查询配置"`
	MonitorForm
}
type CreateRes struct {
	Id int `json:"id"`
}

type UpdateReq struct {
	g.Meta `path:"/api/v2/issue-monitors/{id}" tags:"Issue" method:"put" summary:"更新直发查询配置"`
	Id     int `json:"id" v:"required|min:1"`
	MonitorForm
}
type UpdateRes struct{}

type DeleteReq struct {
	g.Meta `path:"/api/v2/issue-monitors/{id}" tags:"Issue" method:"delete" summary:"删除直发查询配置"`
	Id     int `json:"id" v:"required|min:1"`
}
type DeleteRes struct{}

type CheckReq struct {
	g.Meta `path:"/api/v2/issue-monitors/{id}/check" tags:"Issue" method:"post" summary:"立即查询"`
	Id     int `json:"id" v:"required|min:1"`
}
type CheckRes struct {
	Monitor *issue.Monitor `json:"monitor"`
}

type LogsReq struct {
	g.Meta `path:"/api/v2/issue-monitors/{id}/logs" tags:"Issue" method:"get" summary:"变化记录"`
	Id     int `json:"id"    v:"required|min:1"`
	Limit  int `json:"limit" d:"50" v:"between:1,500"`
}
type LogsRes struct {
	List []*issue.LogItem `json:"list"`
}

type SmtpGetReq struct {
	g.Meta `path:"/api/v2/settings/smtp" tags:"Setting" method:"get" summary:"SMTP 配置"`
}
type SmtpGetRes struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	From        string `json:"from"`
	PasswordSet bool   `json:"passwordSet"`
}

type SmtpSaveReq struct {
	g.Meta   `path:"/api/v2/settings/smtp" tags:"Setting" method:"put" summary:"保存 SMTP 配置"`
	Host     string `json:"host"     v:"required"`
	Port     int    `json:"port"     v:"required|between:1,65535"`
	Username string `json:"username" v:"required"`
	Password string `json:"password"`
	From     string `json:"from"`
}
type SmtpSaveRes struct{}

type SmtpTestReq struct {
	g.Meta `path:"/api/v2/settings/smtp/test" tags:"Setting" method:"post" summary:"发送测试邮件"`
	To     string `json:"to" v:"required|email"`
}
type SmtpTestRes struct{}

type StatsReq struct {
	g.Meta `path:"/api/v2/issue-monitors/{id}/stats" tags:"Issue" method:"get" summary:"趋势与状态周期"`
	Id     int `json:"id"    v:"required|min:1"`
	Hours  int `json:"hours" d:"24" v:"between:1,168"`
}
type StatsRes struct {
	Series []issue.SeriesPoint `json:"series"`
	Rounds []issue.Round       `json:"rounds"`
}

type ReportReq struct {
	g.Meta   `path:"/api/v2/issue-report" tags:"Issue" method:"get" summary:"红包发放个数 (pid 维度)"`
	AppKey   string `json:"appKey"   v:"required"`
	Date     string `json:"date"     v:"date-format:Ymd#日期格式应为 YYYYMMDD"`
	Category int    `json:"category" d:"1" v:"between:1,6"`
}
type ReportRes struct{ *issue.Report }

type ReportTrendReq struct {
	g.Meta   `path:"/api/v2/issue-report/trend" tags:"Issue" method:"get" summary:"红包发放个数近 N 天趋势"`
	AppKey   string `json:"appKey"   v:"required"`
	Category int    `json:"category" d:"1" v:"between:1,6"`
	Days     int    `json:"days"     d:"7" v:"between:1,31"`
}
type ReportTrendRes struct {
	List []issue.TrendPoint `json:"list"`
}
