package issue

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// 表名
const (
	TableMonitor = "issue_monitor"
	TableLog     = "issue_monitor_log"
	TableSetting = "app_setting"
	TableRptDay  = "issue_send_report_day"
	TableRptPid  = "issue_send_report_pid"
)

// schemaSQL 启动时执行, 兼容已有部署 (init.sql 只在 PG 首次初始化时运行)。
const schemaSQL = `
CREATE TABLE IF NOT EXISTS issue_monitor (
    id                SERIAL PRIMARY KEY,
    name              VARCHAR(100) NOT NULL,
    app_key           VARCHAR(64)  NOT NULL,
    app_secret        VARCHAR(128) NOT NULL,
    activity_category SMALLINT     NOT NULL DEFAULT 1,
    interval_sec      INT          NOT NULL DEFAULT 30,
    notify_emails     TEXT         NOT NULL DEFAULT '',
    enabled           BOOLEAN      NOT NULL DEFAULT true,
    can_issue         BOOLEAN,
    auto_payout_cnt   BIGINT,
    manual_claim_2_go BIGINT,
    extra_msg         TEXT         NOT NULL DEFAULT '',
    last_error        TEXT         NOT NULL DEFAULT '',
    checked_at        TIMESTAMPTZ,
    changed_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS issue_monitor_log (
    id                BIGSERIAL PRIMARY KEY,
    monitor_id        INT NOT NULL REFERENCES issue_monitor(id) ON DELETE CASCADE,
    can_issue         BOOLEAN NOT NULL,
    auto_payout_cnt   BIGINT,
    manual_claim_2_go BIGINT,
    status_changed    BOOLEAN NOT NULL DEFAULT false,
    mail_result       TEXT    NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_issue_monitor_log_mid ON issue_monitor_log(monitor_id, id DESC);
CREATE TABLE IF NOT EXISTS issue_send_report_day (
    app_key    VARCHAR(64) NOT NULL,
    biz_date   CHAR(8)     NOT NULL,
    category   SMALLINT    NOT NULL,
    total_fund BIGINT      NOT NULL DEFAULT 0,
    total_use  BIGINT      NOT NULL DEFAULT 0,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (app_key, category, biz_date)
);
CREATE TABLE IF NOT EXISTS issue_send_report_pid (
    app_key  VARCHAR(64) NOT NULL,
    biz_date CHAR(8)     NOT NULL,
    category SMALLINT    NOT NULL,
    pid      VARCHAR(128) NOT NULL,
    fund_num BIGINT      NOT NULL DEFAULT 0,
    use_num  BIGINT      NOT NULL DEFAULT 0,
    PRIMARY KEY (app_key, category, biz_date, pid)
);
CREATE TABLE IF NOT EXISTS app_setting (
    key        VARCHAR(64) PRIMARY KEY,
    value      JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

// EnsureSchema 建表 (幂等)。
func EnsureSchema(ctx context.Context) {
	if _, err := g.DB().Exec(ctx, schemaSQL); err != nil {
		g.Log().Fatalf(ctx, "issue: ensure schema: %v", err)
	}
}
