-- ============================================
-- 广告监测上报平台 - 数据库初始化脚本 v2
-- 数据库: PostgreSQL
-- ============================================

-- 广告平台表
CREATE TABLE IF NOT EXISTS ad_platform (
    id           SERIAL PRIMARY KEY,
    name         VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    config       JSONB NOT NULL DEFAULT '{}',
    status       SMALLINT NOT NULL DEFAULT 1,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE ad_platform IS '广告平台表';
COMMENT ON COLUMN ad_platform.name IS '平台标识: weibo / xiaohongshu';
COMMENT ON COLUMN ad_platform.config IS '平台级配置(reportUrl/caidVersions等)';

-- 平台账户表
CREATE TABLE IF NOT EXISTS ad_account (
    id               SERIAL PRIMARY KEY,
    platform_id      INT NOT NULL REFERENCES ad_platform(id),
    name             VARCHAR(100) NOT NULL,
    credentials      JSONB NOT NULL DEFAULT '{}',
    taolian_config   JSONB NOT NULL DEFAULT '{}',
    fake_enabled     BOOLEAN NOT NULL DEFAULT false,
    fake_rate        SMALLINT NOT NULL DEFAULT 0,
    fake_types       JSONB NOT NULL DEFAULT '[]',
    fake_delay_min   INT NOT NULL DEFAULT 30,
    fake_delay_max   INT NOT NULL DEFAULT 1800,
    debug_mode       BOOLEAN NOT NULL DEFAULT false,
    debug_events     JSONB NOT NULL DEFAULT '[]',
    status           SMALLINT NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMENT ON TABLE ad_account IS '平台账户表';
COMMENT ON COLUMN ad_account.credentials IS '账户凭证(company/advertiser_id/platform等)';
COMMENT ON COLUMN ad_account.taolian_config IS '淘联关联配置(channel/taskId/tbkId/sourcePack)';
COMMENT ON COLUMN ad_account.fake_enabled IS '虚假上报开关';
COMMENT ON COLUMN ad_account.fake_rate IS '虚假上报概率: 1-100';
COMMENT ON COLUMN ad_account.fake_types IS '虚假回传指定 taolian_type 列表,空则从已启用 mapping 随机';
COMMENT ON COLUMN ad_account.debug_mode IS '联调模式开关';
COMMENT ON COLUMN ad_account.debug_events IS '联调关注的广告平台事件类型,如["2","4"](微博action_type),空则关注全部';
COMMENT ON COLUMN ad_account.fake_delay_min IS '虚假回传延迟下限(秒)';
COMMENT ON COLUMN ad_account.fake_delay_max IS '虚假回传延迟上限(秒)';

-- 行为映射表(账户级)
CREATE TABLE IF NOT EXISTS action_mapping (
    id            SERIAL PRIMARY KEY,
    account_id    INT NOT NULL REFERENCES ad_account(id) ON DELETE CASCADE,
    taolian_type  INT NOT NULL,
    taolian_desc  VARCHAR(100),
    platform_type VARCHAR(50) NOT NULL,
    platform_desc VARCHAR(100),
    enabled       BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(account_id, taolian_type)
);
COMMENT ON TABLE action_mapping IS '行为映射: 淘联transformType↔广告平台action_type';

-- 每日聚合统计表(账户维度)
CREATE TABLE IF NOT EXISTS report_stats (
    id                  BIGSERIAL PRIMARY KEY,
    account_id          INT NOT NULL REFERENCES ad_account(id),
    platform_id         INT NOT NULL REFERENCES ad_platform(id),
    stat_date           DATE NOT NULL,
    click_count         INT NOT NULL DEFAULT 0,
    taolian_ok          INT NOT NULL DEFAULT 0,
    taolian_err         INT NOT NULL DEFAULT 0,
    callback_ok         INT NOT NULL DEFAULT 0,
    callback_err        INT NOT NULL DEFAULT 0,
    fake_ok             INT NOT NULL DEFAULT 0,
    fake_err            INT NOT NULL DEFAULT 0,
    drop_count          INT NOT NULL DEFAULT 0,
    debug_ok            INT NOT NULL DEFAULT 0,
    debug_err           INT NOT NULL DEFAULT 0,
    cb_recv             INT NOT NULL DEFAULT 0,
    cb_drop             INT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(account_id, platform_id, stat_date)
);
COMMENT ON TABLE report_stats IS '账户维度每日聚合统计';
COMMENT ON COLUMN report_stats.drop_count IS '监测点击丢弃';
COMMENT ON COLUMN report_stats.debug_ok IS '联调模拟回传成功';
COMMENT ON COLUMN report_stats.debug_err IS '联调模拟回传失败';
COMMENT ON COLUMN report_stats.cb_recv IS '收到淘联转化回调';
COMMENT ON COLUMN report_stats.cb_drop IS '淘联回调丢弃';
-- 报表按日期范围查询 (账户可选), UNIQUE 索引首列是 account_id 无法覆盖纯日期范围扫描
CREATE INDEX IF NOT EXISTS idx_report_stats_stat_date ON report_stats(stat_date);

-- CSV 设备号新增统计表(平台维度)
CREATE TABLE IF NOT EXISTS platform_csv_stats (
    id                  BIGSERIAL PRIMARY KEY,
    platform_id         INT NOT NULL REFERENCES ad_platform(id),
    stat_date           DATE NOT NULL,
    csv_new_imei_md5    INT NOT NULL DEFAULT 0,
    csv_new_idfa_md5    INT NOT NULL DEFAULT 0,
    csv_new_oaid_md5    INT NOT NULL DEFAULT 0,
    csv_new_caid_md5    INT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(platform_id, stat_date)
);
COMMENT ON TABLE platform_csv_stats IS '平台维度每日 CSV 新增设备号统计';

-- 初始化平台数据
-- 注意: caid 版本号不在这里配置, SplitCaid 从广告平台给的 caid 值前缀 "20250325_xxx" 自动解析
INSERT INTO ad_platform (name, display_name, config) VALUES
('weibo', '微博', '{"reportUrl": "https://appmonitor.biz.weibo.com/sdkserver/active"}'),
('weibo_direct', '微博直达', '{"reportUrl": "https://api.biz.weibo.com/v4/track/activate"}'),
('xiaohongshu', '小红书', '{"reportUrl": "https://adapi.xiaohongshu.com/api/open/conversion"}')
ON CONFLICT (name) DO NOTHING;
