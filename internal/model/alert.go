package model

import "time"

// AlertRule 告警规则
type AlertRule struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	Name            string     `json:"name"`
	DeviceID        string     `json:"device_id"`
	DeviceName      string     `json:"device_name"`
	ModelID         string     `json:"model_id"`
	PropertyID      string     `json:"property_id"`
	Operator        string     `json:"operator"` // > >= < <= == !=
	Threshold       float64    `json:"threshold"`
	Severity        string     `json:"severity"` // info warning critical
	Enabled         bool       `json:"enabled"`
	CooldownMinutes int        `json:"cooldown_minutes"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateAlertRuleRequest 创建告警规则请求
type CreateAlertRuleRequest struct {
	Name            string  `json:"name" binding:"required,min=1,max=100"`
	DeviceID        string  `json:"device_id" binding:"required"`
	PropertyID      string  `json:"property_id" binding:"required"`
	Operator        string  `json:"operator" binding:"required,oneof=> >= < <= == !="`
	Threshold       float64 `json:"threshold"`
	Severity        string  `json:"severity" binding:"required,oneof=info warning critical"`
	Enabled         bool    `json:"enabled"`
	CooldownMinutes *int    `json:"cooldown_minutes"`
}

// UpdateAlertRuleRequest 更新告警规则请求
type UpdateAlertRuleRequest struct {
	Name            string  `json:"name" binding:"required,min=1,max=100"`
	DeviceID        string  `json:"device_id" binding:"required"`
	PropertyID      string  `json:"property_id" binding:"required"`
	Operator        string  `json:"operator" binding:"required,oneof=> >= < <= == !="`
	Threshold       float64 `json:"threshold"`
	Severity        string  `json:"severity" binding:"required,oneof=info warning critical"`
	Enabled         bool    `json:"enabled"`
	CooldownMinutes *int    `json:"cooldown_minutes"`
}

// AlertLog 告警日志
type AlertLog struct {
	ID             int64      `json:"id"`
	RuleID         string     `json:"rule_id"`
	UserID         string     `json:"user_id"`
	DeviceID       string     `json:"device_id"`
	DeviceName     string     `json:"device_name"`
	PropertyID     string     `json:"property_id"`
	PropertyName   string     `json:"property_name"`
	Operator       string     `json:"operator"`
	Threshold      float64    `json:"threshold"`
	ActualValue    float64    `json:"actual_value"`
	Severity       string     `json:"severity"`
	RuleName       string     `json:"rule_name"`
	Acknowledged   bool       `json:"acknowledged"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}
