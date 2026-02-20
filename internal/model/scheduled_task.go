package model

import (
	"encoding/json"
	"time"
)

type ScheduledTaskLog struct {
	ID         int64           `json:"id"`
	TaskID     string          `json:"task_id"`
	UserID     string          `json:"user_id"`
	DeviceID   string          `json:"device_id"`
	DeviceName string          `json:"device_name"`
	TaskName   string          `json:"task_name"`
	ActionType string          `json:"action_type"`
	Topic      string          `json:"topic"`
	Payload    json.RawMessage `json:"payload"`
	Status     string          `json:"status"` // "success" | "failed"
	Error      string          `json:"error"`
	CreatedAt  time.Time       `json:"created_at"`
}

type ScheduledTask struct {
	ID         string          `json:"id"`
	UserID     string          `json:"user_id"`
	DeviceID   string          `json:"device_id"`
	DeviceName string          `json:"device_name"`
	Name       string          `json:"name"`
	CronExpr   string          `json:"cron_expr"`
	ActionType string          `json:"action_type"` // property_set / service_invoke
	PropertyID string          `json:"property_id,omitempty"`
	ServiceID  string          `json:"service_id,omitempty"`
	Value      json.RawMessage `json:"value,omitempty"`
	Enabled    bool            `json:"enabled"`
	LastRunAt  *time.Time      `json:"last_run_at"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type CreateScheduledTaskRequest struct {
	DeviceID   string          `json:"device_id" binding:"required"`
	Name       string          `json:"name" binding:"required,min=1,max=100"`
	CronExpr   string          `json:"cron_expr" binding:"required"`
	ActionType string          `json:"action_type" binding:"required,oneof=property_set service_invoke"`
	PropertyID string          `json:"property_id"`
	ServiceID  string          `json:"service_id"`
	Value      json.RawMessage `json:"value"`
	Enabled    bool            `json:"enabled"`
}

type UpdateScheduledTaskRequest struct {
	DeviceID   string          `json:"device_id" binding:"required"`
	Name       string          `json:"name" binding:"required,min=1,max=100"`
	CronExpr   string          `json:"cron_expr" binding:"required"`
	ActionType string          `json:"action_type" binding:"required,oneof=property_set service_invoke"`
	PropertyID string          `json:"property_id"`
	ServiceID  string          `json:"service_id"`
	Value      json.RawMessage `json:"value"`
	Enabled    bool            `json:"enabled"`
}
