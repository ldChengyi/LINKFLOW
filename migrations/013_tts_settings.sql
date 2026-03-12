-- TTS 语音播报设置（豆包声音复刻）
INSERT INTO platform_settings (key, value) VALUES
  ('tts_provider',           'edge'),
  ('tts_doubao_app_id',      ''),
  ('tts_doubao_access_key',  ''),
  ('tts_doubao_resource_id', 'seed-icl-1.0'),
  ('tts_doubao_speaker_id',  '')
ON CONFLICT (key) DO NOTHING;
