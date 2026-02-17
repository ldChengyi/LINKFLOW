-- 005_modules.sql
-- 功能模块表 + 物模型模块绑定

-- ============ Modules Table (功能模块) ============
CREATE TABLE IF NOT EXISTS modules (
    id VARCHAR(32) PRIMARY KEY,                    -- 'voice', 'ota' 等
    name VARCHAR(100) NOT NULL,
    description TEXT,
    -- 模块配置 schema（定义用户可配置的字段）
    config_schema JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TRIGGER IF EXISTS modules_updated_at ON modules;
CREATE TRIGGER modules_updated_at
    BEFORE UPDATE ON modules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- 模块是平台级资源，所有角色可读，仅 admin 可写
GRANT SELECT ON modules TO linkflow_app;
GRANT SELECT ON modules TO linkflow_read;

-- ============ 物模型新增 modules 字段 ============
-- 存储格式: [{"id": "voice", "config": {"exposed_properties": [...], "exposed_services": [...]}}]
ALTER TABLE thing_models ADD COLUMN IF NOT EXISTS modules JSONB NOT NULL DEFAULT '[]';

-- ============ 插入内置语音模块 ============
INSERT INTO modules (id, name, description, config_schema) VALUES (
    'voice',
    '语音控制模块',
    '支持设备端语音识别后上报文本指令，平台解析并路由到目标设备执行操作。启用后设备可通过 voice/up topic 上报语音指令。',
    '{
        "exposed_properties": {
            "type": "array",
            "description": "对语音开放的属性ID列表（仅rw属性有效）",
            "items": { "type": "string" }
        },
        "exposed_services": {
            "type": "array",
            "description": "对语音开放的服务ID列表",
            "items": { "type": "string" }
        }
    }'
) ON CONFLICT (id) DO NOTHING;
