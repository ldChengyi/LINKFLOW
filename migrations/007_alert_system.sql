-- 007: 告警系统
-- alert_rules: 用户定义的告警规则
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    device_id UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    model_id UUID REFERENCES thing_models(id) ON DELETE SET NULL,
    property_id VARCHAR(100) NOT NULL,
    operator VARCHAR(10) NOT NULL CHECK (operator IN ('>', '>=', '<', '<=', '==', '!=')),
    threshold DOUBLE PRECISION NOT NULL DEFAULT 0,
    severity VARCHAR(20) NOT NULL DEFAULT 'info' CHECK (severity IN ('info', 'warning', 'critical')),
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- RLS
ALTER TABLE alert_rules ENABLE ROW LEVEL SECURITY;

CREATE POLICY alert_rules_app_policy ON alert_rules
    FOR ALL
    TO linkflow_app
    USING (user_id::text = current_setting('app.current_user_id', true))
    WITH CHECK (user_id::text = current_setting('app.current_user_id', true));

GRANT SELECT, INSERT, UPDATE, DELETE ON alert_rules TO linkflow_app;
GRANT SELECT ON alert_rules TO linkflow_read;
GRANT ALL ON alert_rules TO linkflow_admin;

-- alert_logs: 触发的告警记录
CREATE TABLE IF NOT EXISTS alert_logs (
    id BIGSERIAL PRIMARY KEY,
    rule_id UUID REFERENCES alert_rules(id) ON DELETE SET NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID NOT NULL,
    device_name VARCHAR(100) NOT NULL DEFAULT '',
    property_id VARCHAR(100) NOT NULL,
    property_name VARCHAR(100) NOT NULL DEFAULT '',
    operator VARCHAR(10) NOT NULL,
    threshold DOUBLE PRECISION NOT NULL,
    actual_value DOUBLE PRECISION NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    rule_name VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- RLS
ALTER TABLE alert_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY alert_logs_app_policy ON alert_logs
    FOR ALL
    TO linkflow_app
    USING (user_id::text = current_setting('app.current_user_id', true))
    WITH CHECK (user_id::text = current_setting('app.current_user_id', true));

GRANT SELECT, INSERT ON alert_logs TO linkflow_app;
GRANT SELECT ON alert_logs TO linkflow_read;
GRANT ALL ON alert_logs TO linkflow_admin;

-- 索引
CREATE INDEX idx_alert_rules_user_id ON alert_rules(user_id);
CREATE INDEX idx_alert_rules_device_id ON alert_rules(device_id);
CREATE INDEX idx_alert_logs_user_id ON alert_logs(user_id);
CREATE INDEX idx_alert_logs_created_at ON alert_logs(created_at DESC);
CREATE INDEX idx_alert_logs_rule_id ON alert_logs(rule_id);
