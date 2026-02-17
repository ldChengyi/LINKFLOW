package mqtt

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"

	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/ws"
)

// EventHook 连接管理 + 消息处理
type EventHook struct {
	mochi.HookBase
	broker *Broker
}

func (h *EventHook) ID() string {
	return "linkflow-event"
}

func (h *EventHook) Provides(b byte) bool {
	return b == mochi.OnConnect || b == mochi.OnDisconnect || b == mochi.OnPublished
}

// OnConnect 设备上线（在线状态已在 AuthHook 中处理）
func (h *EventHook) OnConnect(cl *mochi.Client, pk packets.Packet) error {
	logger.Log.Infof("Device connected: device_id=%s", cl.ID)
	return nil
}

// OnDisconnect 设备离线
func (h *EventHook) OnDisconnect(cl *mochi.Client, err error, expire bool) {
	deviceID := cl.ID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 先取设备信息，再清缓存
	var userID, deviceName string
	if infoVal, ok := h.broker.devices.Load(deviceID); ok {
		info := infoVal.(DeviceInfo)
		userID = info.UserID
		deviceName = info.DeviceName
	}

	// Redis: 标记离线
	if userID != "" {
		if err2 := h.broker.rdb.SetDeviceOffline(ctx, deviceID, userID); err2 != nil {
			logger.Log.Errorf("Failed to set device offline in Redis: device_id=%s, err=%v", deviceID, err2)
		}
	}

	// PG: 持久化记录
	if err2 := h.broker.deviceRepo.UpdateDeviceStatus(ctx, deviceID, "offline"); err2 != nil {
		logger.Log.Errorf("Failed to set device offline: device_id=%s, err=%v", deviceID, err2)
	}

	// 清除设备缓存
	h.broker.devices.Delete(deviceID)

	// WebSocket 推送设备离线 + 统计更新
	if h.broker.hub != nil && userID != "" {
		h.broker.hub.SendToUser(userID, &ws.Message{
			Type: "device_status",
			Data: map[string]interface{}{"device_id": deviceID, "device_name": deviceName, "status": "offline"},
		})
		go h.broker.pushStats(userID)
	}

	// 异步写入设备离线审计日志
	if userID != "" {
		go func(devID, uID, dName string) {
			aCtx, aCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer aCancel()
			h.broker.auditLogRepo.Create(aCtx, &model.AuditLog{
				UserID:    &uID,
				Category:  model.AuditCategoryDevice,
				Action:    "DEVICE_OFFLINE",
				Resource:  devID,
				Detail:    map[string]any{"device_name": dName},
				CreatedAt: time.Now(),
			})
		}(deviceID, userID, deviceName)
	}

	logger.Log.Infof("Device offline: device_id=%s", deviceID)
}

// OnPublished 处理设备上报的消息
func (h *EventHook) OnPublished(cl *mochi.Client, pk packets.Packet) {
	topic := pk.TopicName
	if !shouldProcess(topic) {
		return
	}

	deviceID := cl.ID

	// 语音指令走单独的处理链
	if strings.Contains(topic, "/voice/up") {
		voiceHandler := NewVoiceHandler(h.broker)
		go voiceHandler.HandleVoiceCommand(deviceID, pk.Payload)
		return
	}

	// 从缓存获取设备信息
	infoVal, ok := h.broker.devices.Load(deviceID)
	if !ok {
		logger.Log.Warnf("Device info not found in cache: device_id=%s", deviceID)
		return
	}
	info := infoVal.(DeviceInfo)

	// 解析 payload
	var payload map[string]interface{}
	if err := json.Unmarshal(pk.Payload, &payload); err != nil {
		logger.Log.Warnf("Invalid JSON payload: device_id=%s, topic=%s, err=%v", deviceID, topic, err)
		// 非 JSON 数据，存储原始内容
		payload = map[string]interface{}{"raw": string(pk.Payload)}
		h.store(deviceID, info.UserID, topic, payload, pk.FixedHeader.Qos, false, map[string]string{"_payload": "JSON 解析失败"})
		return
	}

	// 校验 payload
	valid := true
	var errors map[string]string

	if info.ModelID != "" {
		properties := h.getModelProperties(info.ModelID)
		if properties != nil {
			result := ValidatePayload(payload, properties)
			valid = result.Valid
			if !result.Valid {
				errors = result.Errors
				logger.Log.Warnf("Payload validation failed: device_id=%s, errors=%v", deviceID, errors)
			}
		}
	}

	h.store(deviceID, info.UserID, topic, payload, pk.FixedHeader.Qos, valid, errors)

	// WebSocket 推送遥测数据
	if h.broker.hub != nil {
		h.broker.hub.SendToUser(info.UserID, &ws.Message{
			Type: "telemetry",
			Data: map[string]interface{}{
				"device_id":   deviceID,
				"device_name": info.DeviceName,
				"payload":     payload,
				"valid":       valid,
				"errors":      errors,
				"time":        time.Now(),
			},
		})
	}

	// 告警规则评估
	if valid && info.ModelID != "" {
		go h.checkAlertRules(deviceID, info.UserID, info.DeviceName, info.ModelID, payload)
	}
}

// getModelProperties 获取物模型属性（带缓存）
func (h *EventHook) getModelProperties(modelID string) []model.Property {
	// 先查缓存
	if cached, ok := h.broker.models.Load(modelID); ok {
		return cached.([]model.Property)
	}

	// 查数据库
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tm, err := h.broker.thingModelRepo.GetByID(ctx, modelID)
	if err != nil {
		logger.Log.Errorf("Failed to get thing model: model_id=%s, err=%v", modelID, err)
		return nil
	}

	// 缓存
	h.broker.models.Store(modelID, tm.Properties)
	return tm.Properties
}

// store 存储遥测数据
func (h *EventHook) store(deviceID, userID, topic string, payload map[string]interface{}, qos byte, valid bool, errors map[string]string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.broker.deviceDataRepo.InsertTelemetry(ctx, deviceID, userID, topic, payload, qos, valid, errors); err != nil {
		logger.Log.Errorf("Failed to store telemetry: device_id=%s, err=%v", deviceID, err)
	}
}

// checkAlertRules 评估告警规则
func (h *EventHook) checkAlertRules(deviceID, userID, deviceName, modelID string, payload map[string]interface{}) {
	// 从缓存获取规则，未命中则查数据库
	var rules []*model.AlertRule
	if cached, ok := h.broker.alertRules.Load(deviceID); ok {
		rules = cached.([]*model.AlertRule)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var err error
		rules, err = h.broker.alertRuleRepo.ListEnabledByDeviceID(ctx, deviceID)
		if err != nil {
			logger.Log.Errorf("checkAlertRules: load rules failed: device_id=%s, err=%v", deviceID, err)
			return
		}
		h.broker.alertRules.Store(deviceID, rules)
	}

	if len(rules) == 0 {
		return
	}

	// 获取物模型属性名映射
	properties := h.getModelProperties(modelID)
	propNameMap := make(map[string]string)
	for _, p := range properties {
		propNameMap[p.ID] = p.Name
	}

	for _, rule := range rules {
		val, ok := payload[rule.PropertyID]
		if !ok {
			continue
		}

		var numVal float64
		switch v := val.(type) {
		case float64:
			numVal = v
		case json.Number:
			numVal, _ = v.Float64()
		case bool:
			if v {
				numVal = 1
			}
		default:
			continue
		}

		triggered := false
		switch rule.Operator {
		case ">":
			triggered = numVal > rule.Threshold
		case ">=":
			triggered = numVal >= rule.Threshold
		case "<":
			triggered = numVal < rule.Threshold
		case "<=":
			triggered = numVal <= rule.Threshold
		case "==":
			triggered = numVal == rule.Threshold
		case "!=":
			triggered = numVal != rule.Threshold
		}

		if !triggered {
			continue
		}

		propName := propNameMap[rule.PropertyID]
		if propName == "" {
			propName = rule.PropertyID
		}

		alertLog := &model.AlertLog{
			RuleID:       rule.ID,
			UserID:       userID,
			DeviceID:     deviceID,
			DeviceName:   deviceName,
			PropertyID:   rule.PropertyID,
			PropertyName: propName,
			Operator:     rule.Operator,
			Threshold:    rule.Threshold,
			ActualValue:  numVal,
			Severity:     rule.Severity,
			RuleName:     rule.Name,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := h.broker.alertLogRepo.Create(ctx, alertLog); err != nil {
			logger.Log.Errorf("checkAlertRules: write alert log failed: %v", err)
		}
		cancel()

		// WebSocket 推送告警
		if h.broker.hub != nil {
			h.broker.hub.SendToUser(userID, &ws.Message{
				Type: "alert",
				Data: alertLog,
			})
		}

		logger.Log.Warnf("Alert triggered: rule=%s, device=%s, property=%s, value=%.2f %s %.2f",
			rule.Name, deviceName, rule.PropertyID, numVal, rule.Operator, rule.Threshold)
	}
}

// shouldProcess 判断是否需要处理该 topic 的消息
func shouldProcess(topic string) bool {
	return strings.Contains(topic, "/telemetry/up") || strings.Contains(topic, "/event") || strings.Contains(topic, "/voice/up")
}
