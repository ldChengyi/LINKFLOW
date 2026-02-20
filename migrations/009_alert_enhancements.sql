-- 告警规则增强：冷却时间 + 最后触发时间
ALTER TABLE alert_rules
  ADD COLUMN IF NOT EXISTS cooldown_minutes INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_triggered_at TIMESTAMPTZ;

-- 告警日志增强：确认状态
ALTER TABLE alert_logs
  ADD COLUMN IF NOT EXISTS acknowledged BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS acknowledged_at TIMESTAMPTZ;

-- 补充 UPDATE 权限（acknowledge 功能需要）
GRANT UPDATE ON alert_logs TO linkflow_app;

-- 未确认告警索引
CREATE INDEX IF NOT EXISTS idx_alert_logs_unacknowledged
  ON alert_logs(user_id) WHERE acknowledged = false;
