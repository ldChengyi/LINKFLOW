package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/cache"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
)

type MQTTPublisher interface {
	Publish(topic string, payload []byte, retain bool, qos byte) error
	IsClientConnected(clientID string) bool
}

type DebugHandler struct {
	deviceRepo     *repository.DeviceRepository
	thingModelRepo *repository.ThingModelRepository
	dataRepo       *repository.DeviceDataRepository
	pool           *pgxpool.Pool
	rdb            *cache.Redis
	publisher      MQTTPublisher
}

func NewDebugHandler(deviceRepo *repository.DeviceRepository, thingModelRepo *repository.ThingModelRepository, dataRepo *repository.DeviceDataRepository, pool *pgxpool.Pool, rdb *cache.Redis, publisher MQTTPublisher) *DebugHandler {
	return &DebugHandler{deviceRepo: deviceRepo, thingModelRepo: thingModelRepo, dataRepo: dataRepo, pool: pool, rdb: rdb, publisher: publisher}
}

type DebugRequest struct {
	ActionType string                 `json:"action_type" binding:"required,oneof=property_set service_invoke"`
	PropertyID string                 `json:"property_id"`
	Properties map[string]interface{} `json:"properties"` // 批量属性设置
	ServiceID  string                 `json:"service_id"`
	Value      interface{}            `json:"value"`
}

func (h *DebugHandler) withRLS(c *gin.Context) (context.Context, error) {
	userID := c.GetString("user_id")
	var userRole string
	if role, exists := c.Get("user_role"); exists {
		userRole = string(role.(model.UserRole))
	}
	return database.WithRLS(c.Request.Context(), h.pool, userID, userRole)
}

func (h *DebugHandler) Debug(c *gin.Context) {
	var req DebugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	deviceID := c.Param("id")
	device, err := h.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		Fail(c, http.StatusNotFound, "device not found")
		return
	}

	// 必须在线才能调试
	online, _ := h.rdb.IsDeviceOnline(c.Request.Context(), deviceID)
	if !online {
		Fail(c, http.StatusBadRequest, "device is offline")
		return
	}

	// Validate property/service if thing model bound
	if device.ModelID != nil && *device.ModelID != "" {
		tm, err := h.thingModelRepo.GetByID(ctx, *device.ModelID)
		if err == nil {
			if req.ActionType == "property_set" {
				// 批量模式：校验每个 property_id
				if len(req.Properties) > 0 {
					for pid := range req.Properties {
						if !findProperty(tm.Properties, pid) {
							Fail(c, http.StatusBadRequest, "property not found in thing model: "+pid)
							return
						}
					}
				} else if !findProperty(tm.Properties, req.PropertyID) {
					Fail(c, http.StatusBadRequest, "property not found in thing model")
					return
				}
			} else {
				if !findService(tm.Services, req.ServiceID) {
					Fail(c, http.StatusBadRequest, "service not found in thing model")
					return
				}
			}
		}
	}

	// Build MQTT topic + payload
	var topic string
	var payload []byte

	switch req.ActionType {
	case "property_set":
		topic = fmt.Sprintf("devices/%s/telemetry/down", deviceID)
		var data map[string]interface{}
		if len(req.Properties) > 0 {
			data = req.Properties
		} else {
			data = map[string]interface{}{req.PropertyID: req.Value}
		}
		payload, _ = json.Marshal(data)
	case "service_invoke":
		topic = fmt.Sprintf("devices/%s/service/invoke", deviceID)
		msg := map[string]interface{}{
			"id":      fmt.Sprintf("debug_%s_%d", deviceID[:8], time.Now().Unix()),
			"service": req.ServiceID,
			"params":  req.Value,
		}
		payload, _ = json.Marshal(msg)
	}

	if err := h.publisher.Publish(topic, payload, false, 1); err != nil {
		Fail(c, http.StatusInternalServerError, "failed to publish MQTT message")
		return
	}

	// 仅模拟设备需要自动回传（真实设备由硬件自行回传）
	realConn := h.publisher.IsClientConnected(deviceID)
	if !realConn {
		switch req.ActionType {
		case "property_set":
			upTopic := fmt.Sprintf("devices/%s/telemetry/up", deviceID)
			// 合并历史 payload，避免回传数据覆盖掉其他属性
			echoPayload := payload
			if latest, err := h.dataRepo.GetLatestData(context.Background(), deviceID); err == nil && latest != nil {
				merged := make(map[string]interface{}, len(latest.Payload)+len(req.Properties)+1)
				for k, v := range latest.Payload {
					merged[k] = v
				}
				// 覆盖本次下发的属性
				if len(req.Properties) > 0 {
					for k, v := range req.Properties {
						merged[k] = v
					}
				} else {
					merged[req.PropertyID] = req.Value
				}
				if mergedJSON, err := json.Marshal(merged); err == nil {
					echoPayload = mergedJSON
				}
			}
			h.publisher.Publish(upTopic, echoPayload, false, 1)
		case "service_invoke":
			replyTopic := fmt.Sprintf("devices/%s/service/reply", deviceID)
			reply := map[string]interface{}{
				"id":      fmt.Sprintf("debug_%s_%d", deviceID[:8], time.Now().Unix()),
				"service": req.ServiceID,
				"code":    200,
				"message": "success",
			}
			replyPayload, _ := json.Marshal(reply)
			h.publisher.Publish(replyTopic, replyPayload, false, 1)
		}
	}

	Success(c, gin.H{"topic": topic, "payload": json.RawMessage(payload), "auto_echo": !realConn})
}

func findProperty(props []model.Property, id string) bool {
	for _, p := range props {
		if p.ID == id {
			return true
		}
	}
	return false
}

func findService(svcs []model.Service, id string) bool {
	for _, s := range svcs {
		if s.ID == id {
			return true
		}
	}
	return false
}

// SimulateOnline 模拟设备上线（写 Redis，和真实 MQTT 上线同效果）
func (h *DebugHandler) SimulateOnline(c *gin.Context) {
	deviceID := c.Param("id")

	// 如果设备已有真实 MQTT 连接，不允许模拟上线
	if h.publisher.IsClientConnected(deviceID) {
		Fail(c, http.StatusBadRequest, "device already has a real MQTT connection")
		return
	}

	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	device, err := h.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		Fail(c, http.StatusNotFound, "device not found")
		return
	}

	userID := c.GetString("user_id")
	if err := h.rdb.SetSimulatedOnline(c.Request.Context(), device.ID, userID); err != nil {
		Fail(c, http.StatusInternalServerError, "failed to set online")
		return
	}

	Success(c, gin.H{"device_id": device.ID, "status": "online", "connection_type": "simulated"})
}

// SimulateHeartbeat 模拟设备心跳续期
func (h *DebugHandler) SimulateHeartbeat(c *gin.Context) {
	deviceID := c.Param("id")

	// 真实连接不需要心跳续期
	if h.publisher.IsClientConnected(deviceID) {
		Success(c, gin.H{"device_id": deviceID, "connection_type": "real"})
		return
	}

	online, _ := h.rdb.IsDeviceOnline(c.Request.Context(), deviceID)
	if !online {
		Fail(c, http.StatusBadRequest, "device is not simulated online")
		return
	}

	userID := c.GetString("user_id")
	if err := h.rdb.RefreshSimulatedOnline(c.Request.Context(), deviceID, userID); err != nil {
		Fail(c, http.StatusInternalServerError, "failed to refresh heartbeat")
		return
	}

	Success(c, gin.H{"device_id": deviceID, "status": "online"})
}

// SimulateOffline 模拟设备下线
func (h *DebugHandler) SimulateOffline(c *gin.Context) {
	deviceID := c.Param("id")

	// 真实 MQTT 连接的设备不能模拟下线
	if h.publisher.IsClientConnected(deviceID) {
		Fail(c, http.StatusBadRequest, "cannot simulate offline for a device with real MQTT connection")
		return
	}

	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	if _, err := h.deviceRepo.GetByID(ctx, deviceID); err != nil {
		Fail(c, http.StatusNotFound, "device not found")
		return
	}

	userID := c.GetString("user_id")
	if err := h.rdb.SetDeviceOffline(c.Request.Context(), deviceID, userID); err != nil {
		Fail(c, http.StatusInternalServerError, "failed to set offline")
		return
	}

	Success(c, gin.H{"device_id": deviceID, "status": "offline"})
}

// ConnectionType 查询设备连接类型（real / simulated / offline）
func (h *DebugHandler) ConnectionType(c *gin.Context) {
	deviceID := c.Param("id")

	if h.publisher.IsClientConnected(deviceID) {
		Success(c, gin.H{"device_id": deviceID, "connection_type": "real"})
		return
	}

	online, _ := h.rdb.IsDeviceOnline(c.Request.Context(), deviceID)
	if online {
		Success(c, gin.H{"device_id": deviceID, "connection_type": "simulated"})
		return
	}

	Success(c, gin.H{"device_id": deviceID, "connection_type": "offline"})
}
