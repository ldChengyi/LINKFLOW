package model

import "time"

// Firmware 固件
type Firmware struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	FilePath    string    `json:"-"`
	FileSize    int64     `json:"file_size"`
	Checksum    string    `json:"checksum"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// OTATask OTA升级任务
type OTATask struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	DeviceID        string    `json:"device_id"`
	DeviceName      string    `json:"device_name"`
	FirmwareID      string    `json:"firmware_id"`
	FirmwareVersion string    `json:"firmware_version"`
	Status          string    `json:"status"`
	Progress        int       `json:"progress"`
	ErrorMsg        string    `json:"error_msg"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateOTATaskRequest 创建OTA任务请求
type CreateOTATaskRequest struct {
	DeviceID   string `json:"device_id" binding:"required"`
	FirmwareID string `json:"firmware_id" binding:"required"`
}
