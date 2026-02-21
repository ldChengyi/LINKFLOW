-- 平台设置表（全局配置，无 RLS，使用 admin pool）
CREATE TABLE IF NOT EXISTS platform_settings (
  key        VARCHAR(100) PRIMARY KEY,
  value      TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO platform_settings (key, value) VALUES
  ('voice_mode',    'local'),
  ('dify_api_url',  ''),
  ('dify_api_key',  '')
ON CONFLICT (key) DO NOTHING;
