package model

import (
	"encoding/json"
	"time"
)

// ThingModel 物模型
type ThingModel struct {
	ID          string             `json:"id"`
	UserID      string             `json:"user_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Properties  []Property         `json:"properties"`
	Events      []Event            `json:"events"`
	Services    []Service          `json:"services"`
	Modules     []ThingModelModule `json:"modules"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// Property 属性定义
type Property struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	DataType   string      `json:"dataType"`   // int, float, bool, string, enum
	Unit       string      `json:"unit,omitempty"`
	Min        *float64    `json:"min,omitempty"`
	Max        *float64    `json:"max,omitempty"`
	Step       *float64    `json:"step,omitempty"`
	EnumValues []EnumValue `json:"enumValues,omitempty"`
	AccessMode string      `json:"accessMode"` // r, rw
}

// EnumValue 枚举值
type EnumValue struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

// Event 事件定义
type Event struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Params []Param `json:"params,omitempty"`
}

// Service 服务定义
type Service struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	InputParams  []Param `json:"inputParams,omitempty"`
	OutputParams []Param `json:"outputParams,omitempty"`
}

// Param 参数定义
type Param struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	DataType string `json:"dataType"`
}

// ThingModelDB 数据库存储结构
type ThingModelDB struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Properties  []byte    `db:"properties"`
	Events      []byte    `db:"events"`
	Services    []byte    `db:"services"`
	Modules     []byte    `db:"modules"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// ToModel 转换为业务模型
func (db *ThingModelDB) ToModel() (*ThingModel, error) {
	m := &ThingModel{
		ID:        db.ID,
		UserID:    db.UserID,
		Name:      db.Name,
		CreatedAt: db.CreatedAt,
		UpdatedAt: db.UpdatedAt,
	}
	if db.Description != nil {
		m.Description = *db.Description
	}

	if err := json.Unmarshal(db.Properties, &m.Properties); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(db.Events, &m.Events); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(db.Services, &m.Services); err != nil {
		return nil, err
	}
	if db.Modules != nil {
		if err := json.Unmarshal(db.Modules, &m.Modules); err != nil {
			return nil, err
		}
	}
	if m.Modules == nil {
		m.Modules = []ThingModelModule{}
	}
	return m, nil
}

// CreateThingModelRequest 创建物模型请求
type CreateThingModelRequest struct {
	Name        string             `json:"name" binding:"required,min=1,max=100"`
	Description string             `json:"description"`
	Properties  []Property         `json:"properties"`
	Events      []Event            `json:"events"`
	Services    []Service          `json:"services"`
	Modules     []ThingModelModule `json:"modules"`
}

// UpdateThingModelRequest 更新物模型请求
type UpdateThingModelRequest struct {
	Name        string             `json:"name" binding:"required,min=1,max=100"`
	Description string             `json:"description"`
	Properties  []Property         `json:"properties"`
	Events      []Event            `json:"events"`
	Services    []Service          `json:"services"`
	Modules     []ThingModelModule `json:"modules"`
}
