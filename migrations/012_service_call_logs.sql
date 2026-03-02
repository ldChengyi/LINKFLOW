-- 012: 服务调用历史记录表
CREATE TABLE IF NOT EXISTS service_call_logs (
    id            BIGSERIAL PRIMARY KEY,
    device_id     UUID NOT NULL,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_name   VARCHAR(100) NOT NULL DEFAULT '',
    service_id    VARCHAR(100) NOT NULL,
    service_name  VARCHAR(100) NOT NULL DEFAULT '',
    request_id    VARCHAR(64) NOT NULL,
    input_params  JSONB NOT NULL DEFAULT '{}',
    output_params JSONB,
    status        VARCHAR(10) NOT NULL DEFAULT 'pending',  -- pending | success | failed | timeout
    error         TEXT NOT NULL DEFAULT '',
    response_code INT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    replied_at    TIMESTAMPTZ
);

-- RLS
ALTER TABLE service_call_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY scl_app ON service_call_logs FOR ALL TO linkflow_app
    USING (user_id::text = current_setting('app.current_user_id', true));

GRANT SELECT, INSERT, UPDATE ON service_call_logs TO linkflow_app;
GRANT ALL ON service_call_logs TO linkflow_admin;
GRANT USAGE, SELECT ON SEQUENCE service_call_logs_id_seq TO linkflow_app;
GRANT USAGE, SELECT ON SEQUENCE service_call_logs_id_seq TO linkflow_admin;

-- 索引
CREATE INDEX idx_scl_device_id  ON service_call_logs(device_id);
CREATE INDEX idx_scl_user_id    ON service_call_logs(user_id);
CREATE INDEX idx_scl_request_id ON service_call_logs(request_id);
CREATE INDEX idx_scl_created_at ON service_call_logs(created_at DESC);
