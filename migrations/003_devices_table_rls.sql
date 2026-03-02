-- 003_devices_table_rls.sql
-- 设备表 + 行级安全策略（RLS）

CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    device_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'offline' CHECK (
        status IN ('online', 'offline')
    ),
    last_seen_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_devices_user_id ON devices (user_id);

CREATE INDEX idx_devices_status ON devices (status);

CREATE INDEX idx_devices_type ON devices (device_type);

-- 更新时间触发器
CREATE TRIGGER devices_updated_at
    BEFORE UPDATE ON devices
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- 授权
GRANT SELECT, INSERT, UPDATE, DELETE ON devices TO linkflow_app;

GRANT SELECT ON devices TO linkflow_read;

-- ============ 行级安全策略 (RLS) ============

-- 启用 RLS
ALTER TABLE devices ENABLE ROW LEVEL SECURITY;

-- 策略：用户只能操作自己的设备
-- 注意：RLS 通过 current_setting 获取当前用户ID（由应用层设置）

-- SELECT 策略
CREATE POLICY devices_select_policy ON devices FOR
SELECT USING (
        user_id::text = current_setting('app.current_user_id', true)
        OR current_setting('app.current_user_role', true) = 'admin'
    );

-- INSERT 策略
CREATE POLICY devices_insert_policy ON devices FOR INSERT
WITH
    CHECK (
        user_id::text = current_setting('app.current_user_id', true)
        OR current_setting('app.current_user_role', true) = 'admin'
    );

-- UPDATE 策略
CREATE POLICY devices_update_policy ON devices
FOR UPDATE
    USING (
        user_id::text = current_setting('app.current_user_id', true)
        OR current_setting('app.current_user_role', true) = 'admin'
    );

-- DELETE 策略
CREATE POLICY devices_delete_policy ON devices FOR DELETE USING (
    user_id::text = current_setting('app.current_user_id', true)
    OR current_setting('app.current_user_role', true) = 'admin'
);

-- 让 linkflow_app 角色受 RLS 约束
ALTER TABLE devices FORCE ROW LEVEL SECURITY;