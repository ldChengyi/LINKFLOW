package model

import "time"

// Module 功能模块（平台级）
type Module struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	ConfigSchema map[string]interface{} `json:"config_schema"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ModuleDB 数据库存储结构
type ModuleDB struct {
	ID           string     `db:"id"`
	Name         string     `db:"name"`
	Description  *string    `db:"description"`
	ConfigSchema []byte     `db:"config_schema"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

// ThingModelModule 物模型绑定的模块配置
type ThingModelModule struct {
	ID     string            `json:"id"`     // 模块ID，如 "voice"
	Config ModuleConfig      `json:"config"` // 用户配置
}

// ModuleConfig 模块配置（语音模块）
type ModuleConfig struct {
	ExposedProperties []string `json:"exposed_properties,omitempty"`
	ExposedServices   []string `json:"exposed_services,omitempty"`
}

// VoiceCommand 语音指令上报结构
type VoiceCommand struct {
	Text string `json:"text"`
}

// VoiceResult 语音指令执行结果
type VoiceResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Action  string `json:"action,omitempty"`  // 执行的动作描述
}
