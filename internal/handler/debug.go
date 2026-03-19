package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/cache"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
	"go.uber.org/zap"
)

type MQTTPublisher interface {
	Publish(topic string, payload []byte, retain bool, qos byte) error
	IsClientConnected(clientID string) bool
}

type VoiceProcessor interface {
	ProcessCommand(deviceID, text string) *model.VoiceResult
}

type DebugHandler struct {
	deviceRepo     *repository.DeviceRepository
	thingModelRepo *repository.ThingModelRepository
	dataRepo       *repository.DeviceDataRepository
	svcCallLogRepo *repository.ServiceCallLogRepository
	debugLogRepo   *repository.DebugLogRepository
	pool           *pgxpool.Pool
	rdb            *cache.Redis
	publisher      MQTTPublisher
	voiceProcessor VoiceProcessor
}

func NewDebugHandler(deviceRepo *repository.DeviceRepository, thingModelRepo *repository.ThingModelRepository, dataRepo *repository.DeviceDataRepository, svcCallLogRepo *repository.ServiceCallLogRepository, debugLogRepo *repository.DebugLogRepository, pool *pgxpool.Pool, rdb *cache.Redis, publisher MQTTPublisher) *DebugHandler {
	return &DebugHandler{deviceRepo: deviceRepo, thingModelRepo: thingModelRepo, dataRepo: dataRepo, svcCallLogRepo: svcCallLogRepo, debugLogRepo: debugLogRepo, pool: pool, rdb: rdb, publisher: publisher}
}

// SetVoiceProcessor 设置语音处理器（在 broker 初始化后调用）
func (h *DebugHandler) SetVoiceProcessor(vp VoiceProcessor) {
	h.voiceProcessor = vp
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

// VoiceDebug 语音指令调试（同步调用语音处理链路，直接返回结果）
func (h *DebugHandler) VoiceDebug(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, "text is required")
		return
	}

	if h.voiceProcessor == nil {
		Fail(c, http.StatusInternalServerError, "voice processor not initialized")
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

	// 必须在线
	online, _ := h.rdb.IsDeviceOnline(c.Request.Context(), deviceID)
	if !online {
		Fail(c, http.StatusBadRequest, "device is offline")
		return
	}

	// 必须绑定物模型且启用 voice 模块
	if device.ModelID == nil || *device.ModelID == "" {
		Fail(c, http.StatusBadRequest, "device has no thing model")
		return
	}
	tm, err := h.thingModelRepo.GetByID(ctx, *device.ModelID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to load thing model")
		return
	}
	hasVoice := false
	for _, m := range tm.Modules {
		if m.ID == "voice" {
			hasVoice = true
			break
		}
	}
	if !hasVoice {
		Fail(c, http.StatusBadRequest, "thing model does not have voice module enabled")
		return
	}

	// 同步调用语音处理链路
	result := h.voiceProcessor.ProcessCommand(deviceID, req.Text)

	// 判断连接类型
	connType := h.getConnectionType(deviceID)

	// 记录调试日志
	userID := c.GetString("user_id")
	debugLog := &model.DebugLog{
		UserID:         userID,
		DeviceID:       deviceID,
		DeviceName:     device.Name,
		ConnectionType: connType,
		ActionType:     "voice_command",
		Request:        map[string]interface{}{"text": req.Text},
		Success:        result.Success,
		Response:       map[string]interface{}{"message": result.Message, "action": result.Action},
	}
	if !result.Success {
		debugLog.ErrorMessage = result.Message
	}
	h.saveDebugLog(debugLog)

	Success(c, gin.H{
		"success":   result.Success,
		"message":   result.Message,
		"action":    result.Action,
		"audio_url": result.AudioURL,
	})
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

	online, _ := h.rdb.IsDeviceOnline(c.Request.Context(), deviceID)
	if !online {
		Fail(c, http.StatusBadRequest, "device is offline")
		return
	}

	if err := h.validateAgainstThingModel(ctx, device, &req); err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := c.GetString("user_id")
	topic, payload, err := h.buildMQTTMessage(ctx, deviceID, device, &req, userID)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to build MQTT message")
		return
	}
	connType := h.getConnectionType(deviceID)
	debugLog := h.buildDebugLogEntry(userID, deviceID, device, connType, &req)

	if err := h.publisher.Publish(topic, payload, false, 1); err != nil {
		debugLog.Success = false
		debugLog.ErrorMessage = "failed to publish MQTT message"
		h.saveDebugLog(debugLog)
		Fail(c, http.StatusInternalServerError, "failed to publish MQTT message")
		return
	}

	debugLog.Response = map[string]interface{}{"topic": topic, "published": true}
	h.saveDebugLog(debugLog)

	autoEcho := connType != "real"
	if autoEcho {
		h.echoForSimulated(deviceID, &req, payload)
	}

	Success(c, gin.H{"topic": topic, "payload": json.RawMessage(payload), "auto_echo": autoEcho})
}

// validateAgainstThingModel 校验属性/服务是否存在于物模型
func (h *DebugHandler) validateAgainstThingModel(ctx context.Context, device *model.Device, req *DebugRequest) error {
	if device.ModelID == nil || *device.ModelID == "" {
		return nil
	}
	tm, err := h.thingModelRepo.GetByID(ctx, *device.ModelID)
	if err != nil {
		return nil
	}
	if req.ActionType == "property_set" {
		if len(req.Properties) > 0 {
			for pid := range req.Properties {
				if !findProperty(tm.Properties, pid) {
					return fmt.Errorf("property not found in thing model: %s", pid)
				}
			}
		} else if !findProperty(tm.Properties, req.PropertyID) {
			return fmt.Errorf("property not found in thing model")
		}
	} else if !findService(tm.Services, req.ServiceID) {
		return fmt.Errorf("service not found in thing model")
	}
	return nil
}

// buildMQTTMessage 构建 MQTT topic 和 payload，service_invoke 时同步写服务调用日志
func (h *DebugHandler) buildMQTTMessage(ctx context.Context, deviceID string, device *model.Device, req *DebugRequest, userID string) (string, []byte, error) {
	switch req.ActionType {
	case "property_set":
		topic := fmt.Sprintf("devices/%s/telemetry/down", deviceID)
		data := req.Properties
		if len(data) == 0 {
			data = map[string]interface{}{req.PropertyID: req.Value}
		}
		payload, err := json.Marshal(data)
		if err != nil {
			return "", nil, fmt.Errorf("marshal property payload: %w", err)
		}
		return topic, payload, nil
	case "service_invoke":
		topic := fmt.Sprintf("devices/%s/service/invoke", deviceID)
		requestID := fmt.Sprintf("debug_%s_%d", deviceID[:8], time.Now().Unix())
		msg := map[string]interface{}{"id": requestID, "service": req.ServiceID, "params": req.Value}
		payload, err := json.Marshal(msg)
		if err != nil {
			return "", nil, fmt.Errorf("marshal service payload: %w", err)
		}
		h.logServiceCall(ctx, deviceID, device, req, requestID, userID)
		return topic, payload, nil
	default:
		return "", nil, fmt.Errorf("unsupported action type: %s", req.ActionType)
	}
}

// logServiceCall 写入服务调用日志
func (h *DebugHandler) logServiceCall(ctx context.Context, deviceID string, device *model.Device, req *DebugRequest, requestID, userID string) {
	if h.svcCallLogRepo == nil {
		return
	}
	serviceName := req.ServiceID
	if device.ModelID != nil && *device.ModelID != "" {
		if tm, err := h.thingModelRepo.GetByID(ctx, *device.ModelID); err == nil {
			for _, s := range tm.Services {
				if s.ID == req.ServiceID {
					serviceName = s.Name
					break
				}
			}
		}
	}
	inputJSON, _ := json.Marshal(req.Value)
	if _, err := h.svcCallLogRepo.Create(context.Background(), &model.ServiceCallLog{
		DeviceID: deviceID, UserID: userID, DeviceName: device.Name,
		ServiceID: req.ServiceID, ServiceName: serviceName,
		RequestID: requestID, InputParams: inputJSON, Status: "pending",
	}); err != nil {
		logger.Log.Error("failed to save service call log", zap.String("device_id", deviceID), zap.Error(err))
	}
}

func (h *DebugHandler) getConnectionType(deviceID string) string {
	if h.publisher.IsClientConnected(deviceID) {
		return "real"
	}
	return "simulated"
}

// buildDebugLogEntry 构建调试日志条目
func (h *DebugHandler) buildDebugLogEntry(userID, deviceID string, device *model.Device, connType string, req *DebugRequest) *model.DebugLog {
	dl := &model.DebugLog{
		UserID: userID, DeviceID: deviceID, DeviceName: device.Name,
		ConnectionType: connType, ActionType: req.ActionType,
		Request: make(map[string]interface{}), Success: true,
	}
	if req.ActionType == "property_set" {
		if len(req.Properties) > 0 {
			dl.Request["properties"] = req.Properties
		} else {
			dl.Request["property_id"] = req.PropertyID
			dl.Request["value"] = req.Value
		}
	} else {
		dl.Request["service_id"] = req.ServiceID
		dl.Request["params"] = req.Value
	}
	return dl
}

func (h *DebugHandler) saveDebugLog(dl *model.DebugLog) {
	if h.debugLogRepo != nil {
		if err := h.debugLogRepo.Create(context.Background(), dl); err != nil {
			logger.Log.Error("failed to save debug log", zap.String("device_id", dl.DeviceID), zap.Error(err))
		}
	}
}

// echoForSimulated 模拟设备自动回传（闭合数据链路）
func (h *DebugHandler) echoForSimulated(deviceID string, req *DebugRequest, payload []byte) {
	switch req.ActionType {
	case "property_set":
		upTopic := fmt.Sprintf("devices/%s/telemetry/up", deviceID)
		if err := h.publisher.Publish(upTopic, h.mergePropertyPayload(deviceID, req, payload), false, 1); err != nil {
			logger.Log.Error("failed to echo property payload", zap.String("device_id", deviceID), zap.Error(err))
		}
	case "service_invoke":
		h.echoServiceReply(deviceID, req, payload)
	}
}

// mergePropertyPayload 合并历史属性，避免回传覆盖其他属性
func (h *DebugHandler) mergePropertyPayload(deviceID string, req *DebugRequest, payload []byte) []byte {
	latest, err := h.dataRepo.GetLatestData(context.Background(), deviceID)
	if err != nil || latest == nil {
		return payload
	}
	merged := make(map[string]interface{}, len(latest.Payload)+len(req.Properties)+1)
	for k, v := range latest.Payload {
		merged[k] = v
	}
	if len(req.Properties) > 0 {
		for k, v := range req.Properties {
			merged[k] = v
		}
	} else {
		merged[req.PropertyID] = req.Value
	}
	if mergedJSON, err := json.Marshal(merged); err == nil {
		return mergedJSON
	}
	return payload
}

// echoServiceReply 模拟服务调用回复
func (h *DebugHandler) echoServiceReply(deviceID string, req *DebugRequest, payload []byte) {
	var invokeMsg map[string]interface{}
	if err := json.Unmarshal(payload, &invokeMsg); err != nil {
		logger.Log.Error("failed to unmarshal invoke payload for echo", zap.String("device_id", deviceID), zap.Error(err))
		return
	}
	replyID, _ := invokeMsg["id"].(string)
	if replyID == "" {
		replyID = fmt.Sprintf("debug_%s_%d", deviceID[:8], time.Now().Unix())
	}
	replyTopic := fmt.Sprintf("devices/%s/service/reply", deviceID)
	reply := map[string]interface{}{"id": replyID, "service": req.ServiceID, "code": 200, "message": "success"}
	replyPayload, _ := json.Marshal(reply)
	if err := h.publisher.Publish(replyTopic, replyPayload, false, 1); err != nil {
		logger.Log.Error("failed to echo service reply", zap.String("device_id", deviceID), zap.Error(err))
	}
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

// ListDebugLogs 查询设备调试历史
func (h *DebugHandler) ListDebugLogs(c *gin.Context) {
	deviceID := c.Param("id")
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	ctx, err := h.withRLS(c)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "database error")
		return
	}
	defer database.ReleaseRLSConn(ctx)

	logs, total, err := h.debugLogRepo.List(ctx, deviceID, page, pageSize)
	if err != nil {
		Fail(c, http.StatusInternalServerError, "failed to query debug logs")
		return
	}

	Success(c, gin.H{"list": logs, "total": total, "page": page, "page_size": pageSize})
}
