// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PlatformCsvStatsDao is the data access object for the table platform_csv_stats.
type PlatformCsvStatsDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  PlatformCsvStatsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// PlatformCsvStatsColumns defines and stores column names for the table platform_csv_stats.
type PlatformCsvStatsColumns struct {
	Id            string //
	PlatformId    string //
	StatDate      string //
	CsvNewImeiMd5 string //
	CsvNewIdfaMd5 string //
	CsvNewOaidMd5 string //
	CsvNewCaidMd5 string //
	CreatedAt     string //
}

// platformCsvStatsColumns holds the columns for the table platform_csv_stats.
var platformCsvStatsColumns = PlatformCsvStatsColumns{
	Id:            "id",
	PlatformId:    "platform_id",
	StatDate:      "stat_date",
	CsvNewImeiMd5: "csv_new_imei_md5",
	CsvNewIdfaMd5: "csv_new_idfa_md5",
	CsvNewOaidMd5: "csv_new_oaid_md5",
	CsvNewCaidMd5: "csv_new_caid_md5",
	CreatedAt:     "created_at",
}

// NewPlatformCsvStatsDao creates and returns a new DAO object for table data access.
func NewPlatformCsvStatsDao(handlers ...gdb.ModelHandler) *PlatformCsvStatsDao {
	return &PlatformCsvStatsDao{
		group:    "default",
		table:    "platform_csv_stats",
		columns:  platformCsvStatsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PlatformCsvStatsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PlatformCsvStatsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PlatformCsvStatsDao) Columns() PlatformCsvStatsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PlatformCsvStatsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PlatformCsvStatsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PlatformCsvStatsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
