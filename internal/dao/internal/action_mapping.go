// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// ActionMappingDao is the data access object for the table action_mapping.
type ActionMappingDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  ActionMappingColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// ActionMappingColumns defines and stores column names for the table action_mapping.
type ActionMappingColumns struct {
	Id           string //
	AccountId    string //
	TaolianType  string //
	TaolianDesc  string //
	PlatformType string //
	PlatformDesc string //
	Enabled      string //
	CreatedAt    string //
	UpdatedAt    string //
}

// actionMappingColumns holds the columns for the table action_mapping.
var actionMappingColumns = ActionMappingColumns{
	Id:           "id",
	AccountId:    "account_id",
	TaolianType:  "taolian_type",
	TaolianDesc:  "taolian_desc",
	PlatformType: "platform_type",
	PlatformDesc: "platform_desc",
	Enabled:      "enabled",
	CreatedAt:    "created_at",
	UpdatedAt:    "updated_at",
}

// NewActionMappingDao creates and returns a new DAO object for table data access.
func NewActionMappingDao(handlers ...gdb.ModelHandler) *ActionMappingDao {
	return &ActionMappingDao{
		group:    "default",
		table:    "action_mapping",
		columns:  actionMappingColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *ActionMappingDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *ActionMappingDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *ActionMappingDao) Columns() ActionMappingColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *ActionMappingDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *ActionMappingDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *ActionMappingDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
