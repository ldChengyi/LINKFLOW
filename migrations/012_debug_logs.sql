-- 012_debug_logs.sql: 设备调试日志表
CREATE TABLE IF NOT EXISTS debug_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    device_name TEXT NOT NULL,
    connection_type TEXT NOT NULL CHECK (connection_type IN ('real', 'simulated')),
    action_type TEXT NOT NULL CHECK (action_type IN ('property_set', 'service_invoke')),
    request JSONB NOT NULL,
    response JSONB,
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_debug_logs_user_device ON debug_logs(user_id, device_id, created_at DESC);
CREATE INDEX idx_debug_logs_created_at ON debug_logs(created_at DESC);

-- RLS 策略
ALTER TABLE debug_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY debug_logs_user_isolation ON debug_logs
    FOR ALL
    USING (user_id = current_setting('app.current_user_id', true)::uuid);
