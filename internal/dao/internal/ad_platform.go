// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AdPlatformDao is the data access object for the table ad_platform.
type AdPlatformDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  AdPlatformColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// AdPlatformColumns defines and stores column names for the table ad_platform.
type AdPlatformColumns struct {
	Id          string //
	Name        string // 平台标识: weibo / xiaohongshu
	DisplayName string //
	Config      string // 平台级配置(reportUrl/caidVersions等)
	Status      string //
	CreatedAt   string //
	UpdatedAt   string //
}

// adPlatformColumns holds the columns for the table ad_platform.
var adPlatformColumns = AdPlatformColumns{
	Id:          "id",
	Name:        "name",
	DisplayName: "display_name",
	Config:      "config",
	Status:      "status",
	CreatedAt:   "created_at",
	UpdatedAt:   "updated_at",
}

// NewAdPlatformDao creates and returns a new DAO object for table data access.
func NewAdPlatformDao(handlers ...gdb.ModelHandler) *AdPlatformDao {
	return &AdPlatformDao{
		group:    "default",
		table:    "ad_platform",
		columns:  adPlatformColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AdPlatformDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AdPlatformDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AdPlatformDao) Columns() AdPlatformColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AdPlatformDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AdPlatformDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *AdPlatformDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
