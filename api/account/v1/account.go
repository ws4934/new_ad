package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// --- 列表 ---

type AccountListReq struct {
	g.Meta `path:"/api/v2/accounts" tags:"Account" method:"get" summary:"账户列表"`
}

// AccountItem 账户列表项 (带今日统计)
type AccountItem struct {
	Id            int         `json:"id"`
	PlatformId    int         `json:"platformId"`
	PlatformName  string      `json:"platformName"`
	Name          string      `json:"name"`
	Credentials   string      `json:"credentials"`
	TaolianConfig string      `json:"taolianConfig"`
	FakeEnabled   bool        `json:"fakeEnabled"`
	FakeRate      int         `json:"fakeRate"`
	FakeTypes     string      `json:"fakeTypes"`
	FakeDelayMin  int         `json:"fakeDelayMin"`
	FakeDelayMax  int         `json:"fakeDelayMax"`
	DebugMode     bool        `json:"debugMode"`
	DebugEvents   string      `json:"debugEvents"`
	Status        int         `json:"status"`
	CreatedAt     *gtime.Time `json:"createdAt"`
	UpdatedAt     *gtime.Time `json:"updatedAt"`
	// 今日统计
	TodayClick    int `json:"todayClick"`
	TodayTaolian  int `json:"todayTaolian"`
	TodayCallback int `json:"todayCallback"`
	TodayFake     int `json:"todayFake"`
}

type AccountListRes struct {
	List []*AccountItem `json:"list"`
}

// --- 创建 ---

type AccountCreateReq struct {
	g.Meta        `path:"/api/v2/accounts" tags:"Account" method:"post" summary:"创建账户"`
	PlatformId    int    `json:"platformId"    v:"required"`
	Name          string `json:"name"          v:"required"`
	Credentials   string `json:"credentials"`
	TaolianConfig string `json:"taolianConfig"`
	FakeEnabled   bool   `json:"fakeEnabled"`
	FakeRate      int    `json:"fakeRate"      v:"between:0,100"`
	FakeTypes     string `json:"fakeTypes"`
	FakeDelayMin  int    `json:"fakeDelayMin"  v:"min:0"`
	FakeDelayMax  int    `json:"fakeDelayMax"  v:"min:0"`
	DebugMode     bool   `json:"debugMode"`
	DebugEvents   string `json:"debugEvents"`
}

type AccountCreateRes struct {
	Id int `json:"id"`
}

// --- 更新 ---

type AccountUpdateReq struct {
	g.Meta        `path:"/api/v2/accounts/{id}" tags:"Account" method:"put" summary:"更新账户"`
	Id            int    `json:"id"            v:"required"`
	Name          string `json:"name"`
	Credentials   string `json:"credentials"`
	TaolianConfig string `json:"taolianConfig"`
	FakeEnabled   *bool  `json:"fakeEnabled"`
	FakeRate      *int   `json:"fakeRate"      v:"between:0,100"`
	FakeTypes     string `json:"fakeTypes"`
	FakeDelayMin  *int   `json:"fakeDelayMin"  v:"min:0"`
	FakeDelayMax  *int   `json:"fakeDelayMax"  v:"min:0"`
	DebugMode     *bool  `json:"debugMode"`
	DebugEvents   string `json:"debugEvents"`
	Status        *int   `json:"status"`
}

type AccountUpdateRes struct{}

// --- 监测链接生成 ---

type AccountMonitorURLReq struct {
	g.Meta `path:"/api/v2/accounts/{id}/monitor-url" tags:"Account" method:"get" summary:"生成监测链接"`
	Id     int `json:"id" v:"required"`
}

type AccountMonitorURLRes struct {
	URL string `json:"url"`
}

// AccountDeleteReq 删除账户。
type AccountDeleteReq struct {
	g.Meta `path:"/api/v2/accounts/{id}" tags:"Account" method:"delete" summary:"删除账户"`
	Id     int `json:"id" v:"required|min:1"`
}

type AccountDeleteRes struct{}
