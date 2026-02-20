-- 010_ota_system.sql
-- OTA 固件升级系统

-- 设备表新增固件版本字段
ALTER TABLE devices ADD COLUMN IF NOT EXISTS firmware_version VARCHAR(50) DEFAULT '';

-- ============ 固件表 ============

CREATE TABLE firmwares (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    version VARCHAR(50) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    description TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_firmwares_user_id ON firmwares(user_id);

GRANT SELECT, INSERT, DELETE ON firmwares TO linkflow_app;
GRANT SELECT ON firmwares TO linkflow_read;

ALTER TABLE firmwares ENABLE ROW LEVEL SECURITY;

CREATE POLICY firmwares_select_policy ON firmwares
    FOR SELECT USING (
        user_id::text = current_setting('app.current_user_id', true)
        OR current_setting('app.current_user_role', true) = 'admin'
    );

CREATE POLICY firmwares_insert_policy ON firmwares
    FOR INSERT WITH CHECK (
        user_id::text = current_setting('app.current_user_id', true)
        OR current_setting('app.current_user_role', true) = 'admin'
    );

CREATE POLICY firmwares_delete_policy ON firmwares
    FOR DELETE USING (
        user_id::text = current_setting('app.current_user_id', true)
        OR current_setting('app.current_user_role', true) = 'admin'
    );

ALTER TABLE firmwares FORCE ROW LEVEL SECURITY;

-- ============ OTA 任务表 ============

CREATE TABLE ota_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    device_name VARCHAR(100) DEFAULT '',
    firmware_id UUID NOT NULL REFERENCES firmwares(id) ON DELETE CASCADE,
    firmware_version VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'pushing', 'downloading', 'installing', 'completed', 'failed', 'cancelled')),
    progress INT NOT NULL DEFAULT 0,
    error_msg TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ota_tasks_user_id ON ota_tasks(user_id);
CREATE INDEX idx_ota_tasks_device_id ON ota_tasks(device_id);
CREATE INDEX idx_ota_tasks_status ON ota_tasks(status);

CREATE TRIGGER ota_tasks_updated_at
    BEFORE UPDATE ON ota_tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

GRANT SELECT, INSERT, UPDATE ON ota_tasks TO linkflow_app;
GRANT SELECT ON ota_tasks TO linkflow_read;

ALTER TABLE ota_tasks ENABLE ROW LEVEL SECURITY;

CREATE POLICY ota_tasks_select_policy ON ota_tasks
    FOR SELECT USING (
        user_id::text = current_setting('app.current_user_id', true)
        OR current_setting('app.current_user_role', true) = 'admin'
    );

CREATE POLICY ota_tasks_insert_policy ON ota_tasks
    FOR INSERT WITH CHECK (
        user_id::text = current_setting('app.current_user_id', true)
        OR current_setting('app.current_user_role', true) = 'admin'
    );

CREATE POLICY ota_tasks_update_policy ON ota_tasks
    FOR UPDATE USING (
        user_id::text = current_setting('app.current_user_id', true)
        OR current_setting('app.current_user_role', true) = 'admin'
    );

ALTER TABLE ota_tasks FORCE ROW LEVEL SECURITY;
