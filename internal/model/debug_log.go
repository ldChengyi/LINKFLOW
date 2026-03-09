package model

import "time"

type DebugLog struct {
	ID             int64                  `json:"id"`
	UserID         string                 `json:"user_id"`
	DeviceID       string                 `json:"device_id"`
	DeviceName     string                 `json:"device_name"`
	ConnectionType string                 `json:"connection_type"`
	ActionType     string                 `json:"action_type"`
	Request        map[string]interface{} `json:"request"`
	Response       map[string]interface{} `json:"response,omitempty"`
	Success        bool                   `json:"success"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}
