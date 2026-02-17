package model

import (
	"encoding/json"
	"time"
)

// Device 设备模型
type Device struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	ModelID      *string                `json:"model_id"`
	ModelName    string                 `json:"model_name"`
	Name         string                 `json:"name"`
	DeviceSecret string                 `json:"device_secret"`
	Status       string                 `json:"status"`
	LastOnlineAt *time.Time             `json:"last_online_at"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// DeviceDB 数据库存储结构
type DeviceDB struct {
	ID           string     `db:"id"`
	UserID       string     `db:"user_id"`
	ModelID      *string    `db:"model_id"`
	ModelName    *string    `db:"model_name"`
	Name         string     `db:"name"`
	DeviceSecret string     `db:"device_secret"`
	Status       string     `db:"status"`
	LastOnlineAt *time.Time `db:"last_online_at"`
	Metadata     []byte     `db:"metadata"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

// ToModel 转换为业务模型
func (db *DeviceDB) ToModel() (*Device, error) {
	d := &Device{
		ID:           db.ID,
		UserID:       db.UserID,
		ModelID:      db.ModelID,
		Name:         db.Name,
		DeviceSecret: db.DeviceSecret,
		Status:       db.Status,
		LastOnlineAt: db.LastOnlineAt,
		CreatedAt:    db.CreatedAt,
		UpdatedAt:    db.UpdatedAt,
	}
	if db.ModelName != nil {
		d.ModelName = *db.ModelName
	}
	if db.Metadata != nil {
		if err := json.Unmarshal(db.Metadata, &d.Metadata); err != nil {
			return nil, err
		}
	}
	if d.Metadata == nil {
		d.Metadata = make(map[string]interface{})
	}
	return d, nil
}

// CreateDeviceRequest 创建设备请求
type CreateDeviceRequest struct {
	Name     string                 `json:"name" binding:"required,min=1,max=100"`
	ModelID  string                 `json:"model_id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// UpdateDeviceRequest 更新设备请求
type UpdateDeviceRequest struct {
	Name     string                 `json:"name" binding:"required,min=1,max=100"`
	ModelID  string                 `json:"model_id"`
	Metadata map[string]interface{} `json:"metadata"`
}
