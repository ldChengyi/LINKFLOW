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

	// 先取 userID，再清缓存
	var userID string
	if infoVal, ok := h.broker.devices.Load(deviceID); ok {
		userID = infoVal.(DeviceInfo).UserID
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

// shouldProcess 判断是否需要处理该 topic 的消息
func shouldProcess(topic string) bool {
	return strings.Contains(topic, "/telemetry/up") || strings.Contains(topic, "/event") || strings.Contains(topic, "/voice/up")
}
