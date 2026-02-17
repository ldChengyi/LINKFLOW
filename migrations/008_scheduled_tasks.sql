-- 008: 定时任务系统
CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    cron_expr VARCHAR(100) NOT NULL,
    action_type VARCHAR(20) NOT NULL CHECK (action_type IN ('property_set', 'service_invoke')),
    property_id VARCHAR(100),
    service_id VARCHAR(100),
    value JSONB,
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- RLS
ALTER TABLE scheduled_tasks ENABLE ROW LEVEL SECURITY;

CREATE POLICY scheduled_tasks_app_policy ON scheduled_tasks
    FOR ALL
    TO linkflow_app
    USING (user_id::text = current_setting('app.current_user_id', true))
    WITH CHECK (user_id::text = current_setting('app.current_user_id', true));

GRANT SELECT, INSERT, UPDATE, DELETE ON scheduled_tasks TO linkflow_app;
GRANT SELECT ON scheduled_tasks TO linkflow_read;
GRANT ALL ON scheduled_tasks TO linkflow_admin;

-- 索引
CREATE INDEX idx_scheduled_tasks_user_id ON scheduled_tasks(user_id);
CREATE INDEX idx_scheduled_tasks_device_id ON scheduled_tasks(device_id);
CREATE INDEX idx_scheduled_tasks_enabled ON scheduled_tasks(enabled) WHERE enabled = true;

-- updated_at trigger
DROP TRIGGER IF EXISTS scheduled_tasks_updated_at ON scheduled_tasks;
CREATE TRIGGER scheduled_tasks_updated_at
    BEFORE UPDATE ON scheduled_tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
