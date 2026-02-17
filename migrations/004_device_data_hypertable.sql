-- 004_device_data_hypertable.sql
-- 设备遥测数据时序表（TimescaleDB hypertable）

CREATE TABLE IF NOT EXISTS device_data (
    time        TIMESTAMPTZ NOT NULL,
    device_id   UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    topic       VARCHAR(255) NOT NULL,
    payload     JSONB NOT NULL,
    qos         SMALLINT NOT NULL DEFAULT 0,
    valid       BOOLEAN NOT NULL DEFAULT true,
    errors      JSONB
);

-- 转换为 TimescaleDB hypertable
SELECT create_hypertable('device_data', 'time', if_not_exists => TRUE);

-- 索引
CREATE INDEX IF NOT EXISTS idx_device_data_device_time ON device_data (device_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_data_user_time ON device_data (user_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_device_data_valid ON device_data (valid) WHERE valid = false;

-- 授权
GRANT SELECT, INSERT ON device_data TO linkflow_app;
GRANT SELECT ON device_data TO linkflow_read;

-- RLS
ALTER TABLE device_data ENABLE ROW LEVEL SECURITY;

CREATE POLICY device_data_select_policy ON device_data
    FOR SELECT
    USING (
        user_id::text = current_setting('app.current_user_id', true)
        OR current_setting('app.current_user_role', true) = 'admin'
    );

CREATE POLICY device_data_insert_policy ON device_data
    FOR INSERT
    WITH CHECK (
        user_id::text = current_setting('app.current_user_id', true)
        OR current_setting('app.current_user_role', true) = 'admin'
    );

ALTER TABLE device_data FORCE ROW LEVEL SECURITY;
