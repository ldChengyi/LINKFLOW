package model

import (
	"encoding/json"
	"time"
)

type ServiceCallLog struct {
	ID           int64            `json:"id"`
	DeviceID     string           `json:"device_id"`
	UserID       string           `json:"user_id"`
	DeviceName   string           `json:"device_name"`
	ServiceID    string           `json:"service_id"`
	ServiceName  string           `json:"service_name"`
	RequestID    string           `json:"request_id"`
	InputParams  json.RawMessage  `json:"input_params"`
	OutputParams *json.RawMessage `json:"output_params,omitempty"`
	Status       string           `json:"status"`
	Error        string           `json:"error,omitempty"`
	ResponseCode *int             `json:"response_code,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	RepliedAt    *time.Time       `json:"replied_at,omitempty"`
}
