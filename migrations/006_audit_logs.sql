-- 006_audit_logs.sql
-- 审计日志表

CREATE TABLE IF NOT EXISTS audit_logs (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID,                                    -- 操作用户（未认证请求为 NULL）
    category   VARCHAR(20) NOT NULL DEFAULT 'api',      -- 'api' | 'device'
    action     VARCHAR(100) NOT NULL,                   -- "POST /api/devices" 或 "DEVICE_ONLINE"
    resource   VARCHAR(255) NOT NULL,                   -- "/api/devices/xxx" 或 device_id
    detail     JSONB,                                   -- 请求体摘要 或 设备事件详情
    ip         VARCHAR(45) NOT NULL DEFAULT '',
    status_code SMALLINT NOT NULL DEFAULT 0,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    user_agent VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_audit_logs_user_id ON audit_logs (user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs (action);
CREATE INDEX idx_audit_logs_category ON audit_logs (category);

-- 仅 admin 可读写审计日志，app/read 无权限
GRANT SELECT, INSERT ON audit_logs TO linkflow_admin;
GRANT USAGE, SELECT ON SEQUENCE audit_logs_id_seq TO linkflow_admin;
