CREATE TABLE scheduled_task_logs (
    id           BIGSERIAL PRIMARY KEY,
    task_id      UUID NOT NULL REFERENCES scheduled_tasks(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id    UUID NOT NULL,
    device_name  VARCHAR(100) NOT NULL DEFAULT '',
    task_name    VARCHAR(100) NOT NULL DEFAULT '',
    action_type  VARCHAR(20)  NOT NULL,
    topic        TEXT         NOT NULL,
    payload      JSONB        NOT NULL,
    status       VARCHAR(10)  NOT NULL,
    error        TEXT         NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

ALTER TABLE scheduled_task_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY stl_app ON scheduled_task_logs FOR ALL TO linkflow_app
    USING (user_id::text = current_setting('app.current_user_id', true));

GRANT SELECT, INSERT ON scheduled_task_logs TO linkflow_app;
GRANT ALL ON scheduled_task_logs TO linkflow_admin;
GRANT USAGE, SELECT ON SEQUENCE scheduled_task_logs_id_seq TO linkflow_app;
GRANT USAGE, SELECT ON SEQUENCE scheduled_task_logs_id_seq TO linkflow_admin;

CREATE INDEX idx_stl_task_id    ON scheduled_task_logs(task_id);
CREATE INDEX idx_stl_device_id  ON scheduled_task_logs(device_id);
CREATE INDEX idx_stl_user_id    ON scheduled_task_logs(user_id);
CREATE INDEX idx_stl_created_at ON scheduled_task_logs(created_at DESC);
