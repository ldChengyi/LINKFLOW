package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
)

// VoiceHandler 语音指令处理器
type VoiceHandler struct {
	broker *Broker
}

// NewVoiceHandler 创建语音处理器
func NewVoiceHandler(broker *Broker) *VoiceHandler {
	return &VoiceHandler{broker: broker}
}

// HandleVoiceCommand 处理语音指令
func (h *VoiceHandler) HandleVoiceCommand(deviceID string, payload []byte) {
	var cmd model.VoiceCommand
	if err := json.Unmarshal(payload, &cmd); err != nil {
		h.reply(deviceID, model.VoiceResult{Success: false, Message: "无效的指令格式"})
		return
	}

	if cmd.Text == "" {
		h.reply(deviceID, model.VoiceResult{Success: false, Message: "指令文本为空"})
		return
	}

	logger.Log.Infof("Voice command received: device_id=%s, text=%s", deviceID, cmd.Text)

	// 获取发送指令的设备信息
	infoVal, ok := h.broker.devices.Load(deviceID)
	if !ok {
		h.reply(deviceID, model.VoiceResult{Success: false, Message: "设备信息未找到"})
		return
	}
	info := infoVal.(DeviceInfo)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取同一用户的所有设备作为潜在目标
	targets, err := h.broker.deviceRepo.ListByUserID(ctx, info.UserID)
	if err != nil {
		logger.Log.Errorf("Voice: failed to list user devices: %v", err)
		h.reply(deviceID, model.VoiceResult{Success: false, Message: "获取设备列表失败"})
		return
	}

	// 解析意图
	intent := h.parseIntent(ctx, cmd.Text, targets)
	if intent == nil {
		h.reply(deviceID, model.VoiceResult{Success: false, Message: "无法理解指令: " + cmd.Text})
		return
	}

	// 执行指令
	result := h.execute(ctx, intent)
	h.reply(deviceID, result)
}

// voiceIntent 解析后的语音意图
type voiceIntent struct {
	TargetDevice *model.Device
	ThingModel   *model.ThingModel
	ActionType   string // "property_set", "service_invoke"
	PropertyID   string
	ServiceID    string
	Value        interface{}
}

// parseIntent 从文本中解析意图
func (h *VoiceHandler) parseIntent(ctx context.Context, text string, devices []*model.Device) *voiceIntent {
	text = strings.TrimSpace(text)

	// 1. 匹配目标设备
	var targetDevice *model.Device
	var bestMatchLen int
	for _, d := range devices {
		if strings.Contains(text, d.Name) && len(d.Name) > bestMatchLen {
			targetDevice = d
			bestMatchLen = len(d.Name)
		}
	}

	if targetDevice == nil {
		// 如果只有一个设备（除了语音设备本身），默认选它
		if len(devices) == 2 {
			for _, d := range devices {
				if d.ModelID != nil {
					targetDevice = d
					break
				}
			}
		}
		if targetDevice == nil {
			return nil
		}
	}

	// 2. 获取目标设备的物模型
	if targetDevice.ModelID == nil {
		return nil
	}

	tm, err := h.broker.thingModelRepo.GetByID(ctx, *targetDevice.ModelID)
	if err != nil {
		return nil
	}

	// 3. 检查物模型是否启用了语音模块
	var voiceModule *model.ThingModelModule
	for i, m := range tm.Modules {
		if m.ID == "voice" {
			voiceModule = &tm.Modules[i]
			break
		}
	}
	if voiceModule == nil {
		return nil
	}

	// 4. 匹配属性操作
	for _, prop := range tm.Properties {
		if prop.AccessMode != "rw" {
			continue
		}
		if !isExposed(prop.ID, voiceModule.Config.ExposedProperties) {
			continue
		}

		// 尝试匹配属性名
		if !strings.Contains(text, prop.Name) && !strings.Contains(text, prop.ID) {
			continue
		}

		// 根据数据类型解析值
		value := h.parseValue(text, prop)
		if value != nil {
			return &voiceIntent{
				TargetDevice: targetDevice,
				ThingModel:   tm,
				ActionType:   "property_set",
				PropertyID:   prop.ID,
				Value:        value,
			}
		}
	}

	// 5. 尝试 bool 类型的模糊匹配（"打开/关闭" + 设备名）
	for _, prop := range tm.Properties {
		if prop.AccessMode != "rw" || prop.DataType != "bool" {
			continue
		}
		if !isExposed(prop.ID, voiceModule.Config.ExposedProperties) {
			continue
		}

		if matchBoolOn(text) {
			return &voiceIntent{
				TargetDevice: targetDevice,
				ThingModel:   tm,
				ActionType:   "property_set",
				PropertyID:   prop.ID,
				Value:        true,
			}
		}
		if matchBoolOff(text) {
			return &voiceIntent{
				TargetDevice: targetDevice,
				ThingModel:   tm,
				ActionType:   "property_set",
				PropertyID:   prop.ID,
				Value:        false,
			}
		}
	}

	// 6. 匹配服务调用
	for _, svc := range tm.Services {
		if !isExposed(svc.ID, voiceModule.Config.ExposedServices) {
			continue
		}
		if strings.Contains(text, svc.Name) || strings.Contains(text, svc.ID) {
			return &voiceIntent{
				TargetDevice: targetDevice,
				ThingModel:   tm,
				ActionType:   "service_invoke",
				ServiceID:    svc.ID,
			}
		}
	}

	return nil
}

// parseValue 根据属性类型从文本中提取值
func (h *VoiceHandler) parseValue(text string, prop model.Property) interface{} {
	switch prop.DataType {
	case "bool":
		if matchBoolOn(text) {
			return true
		}
		if matchBoolOff(text) {
			return false
		}
	case "int":
		// 匹配 "调到X" / "设为X" / "设置到X" 中的数字
		if v, ok := extractNumber(text); ok {
			intVal := int(v)
			if prop.Min != nil && float64(intVal) < *prop.Min {
				intVal = int(*prop.Min)
			}
			if prop.Max != nil && float64(intVal) > *prop.Max {
				intVal = int(*prop.Max)
			}
			return intVal
		}
		// 匹配 "调高/增大" → 当前值 + step
		if matchIncrease(text) {
			return map[string]string{"_action": "increase"}
		}
		if matchDecrease(text) {
			return map[string]string{"_action": "decrease"}
		}
	case "float":
		if v, ok := extractNumber(text); ok {
			if prop.Min != nil && v < *prop.Min {
				v = *prop.Min
			}
			if prop.Max != nil && v > *prop.Max {
				v = *prop.Max
			}
			return v
		}
		if matchIncrease(text) {
			return map[string]string{"_action": "increase"}
		}
		if matchDecrease(text) {
			return map[string]string{"_action": "decrease"}
		}
	case "enum":
		for _, ev := range prop.EnumValues {
			if strings.Contains(text, ev.Label) {
				return ev.Value
			}
		}
	}
	return nil
}

// execute 执行解析后的意图
func (h *VoiceHandler) execute(ctx context.Context, intent *voiceIntent) model.VoiceResult {
	deviceID := intent.TargetDevice.ID

	// 获取设备所属用户
	infoVal, _ := h.broker.devices.Load(intent.TargetDevice.ID)
	var userID string
	if infoVal != nil {
		userID = infoVal.(DeviceInfo).UserID
	}

	switch intent.ActionType {
	case "property_set":
		value := intent.Value

		// 处理相对值操作（increase/decrease）
		if actionMap, ok := value.(map[string]string); ok {
			action := actionMap["_action"]
			resolved, err := h.resolveRelativeValue(ctx, deviceID, intent.PropertyID, intent.ThingModel, action)
			if err != nil {
				return model.VoiceResult{Success: false, Message: err.Error()}
			}
			value = resolved
		}

		topic := fmt.Sprintf("devices/%s/telemetry/down", deviceID)
		payloadMap := map[string]interface{}{intent.PropertyID: value}
		payloadBytes, _ := json.Marshal(payloadMap)

		if err := h.broker.Publish(topic, payloadBytes, false, 1); err != nil {
			logger.Log.Errorf("Voice: failed to publish: %v", err)
			return model.VoiceResult{Success: false, Message: "指令下发失败"}
		}

		// 同步写入 device_data，合并已有属性后存储完整状态
		if userID != "" {
			merged := map[string]interface{}{intent.PropertyID: value}
			if existing, err := h.broker.deviceDataRepo.GetLatestData(ctx, deviceID); err == nil && existing != nil {
				for k, v := range existing.Payload {
					if _, set := merged[k]; !set {
						merged[k] = v
					}
				}
			}
			storeTopic := fmt.Sprintf("devices/%s/telemetry/up", deviceID)
			if err := h.broker.deviceDataRepo.InsertTelemetry(ctx, deviceID, userID, storeTopic, merged, 1, true, nil); err != nil {
				logger.Log.Errorf("Voice: failed to store property: %v", err)
			}
		}

		action := fmt.Sprintf("设置 %s.%s = %v", intent.TargetDevice.Name, intent.PropertyID, value)
		logger.Log.Infof("Voice executed: %s", action)
		return model.VoiceResult{Success: true, Message: "指令已执行", Action: action}

	case "service_invoke":
		topic := fmt.Sprintf("devices/%s/service/invoke", deviceID)
		payload, _ := json.Marshal(map[string]interface{}{
			"id":      fmt.Sprintf("voice_%d", time.Now().UnixMilli()),
			"service": intent.ServiceID,
			"params":  map[string]interface{}{},
		})

		if err := h.broker.Publish(topic, payload, false, 1); err != nil {
			logger.Log.Errorf("Voice: failed to publish service invoke: %v", err)
			return model.VoiceResult{Success: false, Message: "服务调用失败"}
		}

		action := fmt.Sprintf("调用 %s.%s", intent.TargetDevice.Name, intent.ServiceID)
		logger.Log.Infof("Voice executed: %s", action)
		return model.VoiceResult{Success: true, Message: "服务已调用", Action: action}
	}

	return model.VoiceResult{Success: false, Message: "未知操作类型"}
}

// resolveRelativeValue 解析相对值（调高/调低）
func (h *VoiceHandler) resolveRelativeValue(ctx context.Context, deviceID, propertyID string, tm *model.ThingModel, action string) (interface{}, error) {
	// 获取当前值
	data, err := h.broker.deviceDataRepo.GetLatestData(ctx, deviceID)
	if err != nil || data == nil {
		return nil, fmt.Errorf("无法获取设备当前数据")
	}

	currentVal, ok := data.Payload[propertyID]
	if !ok {
		return nil, fmt.Errorf("属性 %s 无当前值", propertyID)
	}

	// 找到属性定义获取 step
	var prop *model.Property
	for i, p := range tm.Properties {
		if p.ID == propertyID {
			prop = &tm.Properties[i]
			break
		}
	}
	if prop == nil {
		return nil, fmt.Errorf("属性 %s 未定义", propertyID)
	}

	step := 10.0 // 默认步长
	if prop.Step != nil {
		step = *prop.Step
	}

	current, _ := toFloat64(currentVal)
	var newVal float64
	if action == "increase" {
		newVal = current + step
	} else {
		newVal = current - step
	}

	if prop.Min != nil && newVal < *prop.Min {
		newVal = *prop.Min
	}
	if prop.Max != nil && newVal > *prop.Max {
		newVal = *prop.Max
	}

	if prop.DataType == "int" {
		return int(newVal), nil
	}
	return newVal, nil
}

// reply 回复语音指令结果
func (h *VoiceHandler) reply(deviceID string, result model.VoiceResult) {
	topic := fmt.Sprintf("devices/%s/voice/down", deviceID)
	payload, _ := json.Marshal(result)
	if err := h.broker.Publish(topic, payload, false, 1); err != nil {
		logger.Log.Errorf("Voice: failed to reply: device_id=%s, err=%v", deviceID, err)
	}
}

// ---- 关键词匹配工具函数 ----

var (
	onWords       = []string{"打开", "开启", "开灯", "启动", "开"}
	offWords      = []string{"关闭", "关掉", "关灯", "停止", "关"}
	increaseWords = []string{"调高", "增大", "调亮", "升高", "加大", "高一点", "亮一点", "大一点"}
	decreaseWords = []string{"调低", "减小", "调暗", "降低", "减小", "低一点", "暗一点", "小一点"}
	numberRegex   = regexp.MustCompile(`(\d+\.?\d*)`)
)

func matchBoolOn(text string) bool {
	for _, w := range onWords {
		if strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func matchBoolOff(text string) bool {
	for _, w := range offWords {
		if strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func matchIncrease(text string) bool {
	for _, w := range increaseWords {
		if strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func matchDecrease(text string) bool {
	for _, w := range decreaseWords {
		if strings.Contains(text, w) {
			return true
		}
	}
	return false
}

func extractNumber(text string) (float64, bool) {
	matches := numberRegex.FindStringSubmatch(text)
	if len(matches) < 2 {
		return 0, false
	}
	v, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func isExposed(id string, exposed []string) bool {
	if len(exposed) == 0 {
		return false
	}
	for _, e := range exposed {
		if e == id {
			return true
		}
	}
	return false
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	}
	return 0, false
}
