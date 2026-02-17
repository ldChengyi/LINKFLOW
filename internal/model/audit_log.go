package model

import "time"

// 审计日志分类
const (
	AuditCategoryAPI    = "api"    // API 请求
	AuditCategoryDevice = "device" // 设备事件（上线/下线）
)

// AuditLog 操作审计日志
type AuditLog struct {
	ID         int64          `json:"id"`
	UserID     *string        `json:"user_id,omitempty"`
	Category   string         `json:"category"`
	Action     string         `json:"action"`
	Resource   string         `json:"resource"`
	Detail     map[string]any `json:"detail,omitempty"`
	IP         string         `json:"ip"`
	StatusCode int            `json:"status_code"`
	LatencyMs  int64          `json:"latency_ms"`
	UserAgent  string         `json:"user_agent,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// AuditLogQuery 审计日志查询参数
type AuditLogQuery struct {
	UserID    string `form:"user_id"`
	Category  string `form:"category"`
	Action    string `form:"action"`
	StartTime string `form:"start_time"` // RFC3339
	EndTime   string `form:"end_time"`   // RFC3339
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}
