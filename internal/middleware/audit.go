package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
	"go.uber.org/zap"
)

// 敏感字段列表
var sensitiveFields = map[string]bool{
	"password":      true,
	"device_secret": true,
	"token":         true,
	"credential":    true,
}

// 路由 -> 操作描述映射
var actionMap = map[string]struct {
	Action       string // 人类可读操作
	ResourceType string // 资源类型
}{
	"POST /api/auth/register":    {"用户注册", "auth"},
	"POST /api/auth/login":       {"用户登录", "auth"},
	"POST /api/auth/logout":      {"用户登出", "auth"},
	"POST /api/thing-models":     {"创建物模型", "thing_model"},
	"PUT /api/thing-models/:id":  {"更新物模型", "thing_model"},
	"DELETE /api/thing-models/:id": {"删除物模型", "thing_model"},
	"POST /api/devices":               {"创建设备", "device"},
	"PUT /api/devices/:id":            {"更新设备", "device"},
	"DELETE /api/devices/:id":         {"删除设备", "device"},
	"POST /api/scheduled-tasks":       {"创建定时任务", "scheduled_task"},
	"PUT /api/scheduled-tasks/:id":    {"更新定时任务", "scheduled_task"},
	"DELETE /api/scheduled-tasks/:id": {"删除定时任务", "scheduled_task"},
	"POST /api/devices/:id/debug":            {"在线调试", "device"},
	"POST /api/devices/:id/simulate/online":  {"模拟上线", "device"},
	"POST /api/devices/:id/simulate/offline": {"模拟下线", "device"},
	"POST /api/alert-rules":          {"创建告警规则", "alert_rule"},
	"PUT /api/alert-rules/:id":       {"更新告警规则", "alert_rule"},
	"DELETE /api/alert-rules/:id":    {"删除告警规则", "alert_rule"},
	"PUT /api/auth/password":         {"修改密码", "auth"},
	"PUT /api/alert-logs/:id/acknowledge": {"确认告警", "alert_log"},
}

// AuditLog 操作审计日志中间件
// 记录所有写操作（POST/PUT/DELETE）+ 登录登出，异步写入数据库
func AuditLog(repo *repository.AuditLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		// 跳过 GET、OPTIONS、HEAD
		if method == "GET" || method == "OPTIONS" || method == "HEAD" {
			c.Next()
			return
		}

		start := time.Now()
		body := readRequestBody(c)
		detail := sanitizeBody(body)

		c.Next()

		routeKey := method + " " + c.FullPath()
		action, resourceType := resolveAction(routeKey, method, body)
		resourcePath := c.Request.URL.Path
		enrichDetail(detail, body, resourceType, resourcePath)

		var userID *string
		if id := GetUserID(c); id != "" {
			userID = &id
		}

		auditLog := &model.AuditLog{
			UserID: userID, Category: model.AuditCategoryAPI, Action: action,
			Resource: resourcePath, Detail: detail, IP: c.ClientIP(),
			StatusCode: c.Writer.Status(), LatencyMs: time.Since(start).Milliseconds(),
			UserAgent: c.Request.UserAgent(), CreatedAt: time.Now(),
		}

		go func(log *model.AuditLog) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := repo.Create(ctx, log); err != nil {
				logger.Log.Error("审计日志写入失败",
					zap.Error(err),
					zap.String("action", log.Action),
					zap.String("resource", log.Resource),
				)
			}
		}(auditLog)
	}
}

// readRequestBody 读取并缓存请求体
func readRequestBody(c *gin.Context) map[string]any {
	if c.Request.Body == nil {
		return nil
	}
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Log.Warn("审计日志: 读取请求体失败", zap.Error(err))
		return nil
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	if len(bodyBytes) > 10240 {
		return nil
	}
	var body map[string]any
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		return nil
	}
	return body
}

// sanitizeBody 脱敏请求体中的敏感字段
func sanitizeBody(body map[string]any) map[string]any {
	detail := make(map[string]any, len(body))
	for k, v := range body {
		if sensitiveFields[k] {
			detail[k] = "***"
		} else {
			detail[k] = v
		}
	}
	return detail
}

// resolveAction 解析路由生成可读的 action 和 resource_type
func resolveAction(routeKey, method string, body map[string]any) (string, string) {
	action := routeKey
	resourceType := ""
	if mapped, ok := actionMap[routeKey]; ok {
		action = mapped.Action
		resourceType = mapped.ResourceType
	}
	if resourceType == "scheduled_task" && method == "PUT" {
		if enabled, ok := body["enabled"].(bool); ok {
			if enabled {
				action = "启用定时任务"
			} else {
				action = "禁用定时任务"
			}
		}
	}
	return action, resourceType
}

// enrichDetail 向 detail 中补充资源信息
func enrichDetail(detail map[string]any, body map[string]any, resourceType, resourcePath string) {
	if name, ok := body["name"].(string); ok && name != "" {
		detail["resource_name"] = name
	}
	if email, ok := body["email"].(string); ok && email != "" && resourceType == "auth" {
		detail["resource_name"] = email
	}
	if resourceType != "" {
		detail["resource_type"] = resourceType
	}
	if id := extractResourceID(resourcePath, resourceType); id != "" {
		detail["resource_id"] = id
	}
}

// extractResourceID 从 URL 路径中提取资源 ID
func extractResourceID(path, resourceType string) string {
	var prefix string
	switch resourceType {
	case "thing_model":
		prefix = "/api/thing-models/"
	case "device":
		prefix = "/api/devices/"
	case "scheduled_task":
		prefix = "/api/scheduled-tasks/"
	case "alert_rule":
		prefix = "/api/alert-rules/"
	case "alert_log":
		prefix = "/api/alert-logs/"
	default:
		return ""
	}
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	id := strings.TrimPrefix(path, prefix)
	// 去掉后续路径段（如 /data/latest）
	if idx := strings.Index(id, "/"); idx != -1 {
		id = id[:idx]
	}
	return id
}
