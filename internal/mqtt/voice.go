package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ldchengyi/linkflow/internal/ai"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/ws"
)

// VoiceHandler 语音指令处理器（Pipeline 模式）
type VoiceHandler struct {
	broker   *Broker
	pipeline *Pipeline
}

// NewVoiceHandler 创建语音处理器并组装管道
func NewVoiceHandler(broker *Broker) *VoiceHandler {
	p := NewPipeline(
		&PreprocessNode{},
		&DeviceLoadNode{broker: broker},
		&IntentClassifyNode{},
		&EntityExtractNode{broker: broker},
		&SlotValidateNode{},
		&ExecuteNode{broker: broker},
	)
	return &VoiceHandler{broker: broker, pipeline: p}
}

// HandleVoiceCommand 处理语音指令
func (h *VoiceHandler) HandleVoiceCommand(deviceID string, payload []byte) {
	var cmd model.VoiceCommand
	if err := json.Unmarshal(payload, &cmd); err != nil {
		h.reply(deviceID, &model.VoiceResult{Success: false, Message: "无效的指令格式"})
		return
	}
	if cmd.Text == "" {
		h.reply(deviceID, &model.VoiceResult{Success: false, Message: "指令文本为空"})
		return
	}

	result := h.ProcessCommand(deviceID, cmd.Text)
	h.reply(deviceID, result)
}

// ProcessCommand 同步处理语音指令，返回结果（供 HTTP 调试接口直接调用）
func (h *VoiceHandler) ProcessCommand(deviceID, text string) *model.VoiceResult {
	logger.Log.Infof("Voice command received: device_id=%s, text=%s", deviceID, text)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取发送指令的设备信息（缓存未命中时查 DB，支持模拟设备场景）
	infoVal, ok := h.broker.devices.Load(deviceID)
	if !ok {
		device, err := h.broker.deviceRepo.GetByID(ctx, deviceID)
		if err != nil {
			return &model.VoiceResult{Success: false, Message: "设备信息未找到"}
		}
		info := DeviceInfo{UserID: device.UserID, DeviceName: device.Name}
		if device.ModelID != nil {
			info.ModelID = *device.ModelID
		}
		h.broker.devices.Store(deviceID, info)
		infoVal = info
	}
	info := infoVal.(DeviceInfo)

	var result *model.VoiceResult

	// 根据 voice_mode 选择处理路径
	settings := h.broker.GetVoiceSettings()
	if settings.Mode == "dify" && settings.APIURL != "" && settings.APIKey != "" {
		result = h.handleWithDify(ctx, deviceID, text, info)
	} else {
		// 本地 NLP 路径（默认）
		pc := &PipelineContext{
			RawText:  text,
			DeviceID: deviceID,
			UserID:   info.UserID,
		}
		result = h.pipeline.Run(ctx, pc)
	}

	// TTS 合成
	if h.broker.ttsService != nil && result.Message != "" {
		ttsCtx, ttsCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer ttsCancel()
		filename, err := h.broker.ttsService.Synthesize(ttsCtx, result.Message)
		if err != nil {
			logger.Log.Warnf("Voice TTS failed: device_id=%s, err=%v", deviceID, err)
		} else {
			result.AudioURL = "/api/tts/" + filename
		}
	}

	return result
}

// reply 回复语音指令结果
func (h *VoiceHandler) reply(deviceID string, result *model.VoiceResult) {
	// 生成 TTS 音频
	if h.broker.ttsService != nil && result.Message != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filename, err := h.broker.ttsService.Synthesize(ctx, result.Message)
		if err != nil {
			logger.Log.Warnf("Voice TTS failed: device_id=%s, err=%v", deviceID, err)
		} else {
			result.AudioURL = fmt.Sprintf("%s/tts/%s", h.broker.baseURL, filename)
		}
	}

	topic := fmt.Sprintf("devices/%s/voice/down", deviceID)
	payload, _ := json.Marshal(result)
	if err := h.broker.Publish(topic, payload, false, 1); err != nil {
		logger.Log.Errorf("Voice: failed to reply: device_id=%s, err=%v", deviceID, err)
	}

	// WebSocket 推送语音结果，供前端调试页面实时展示
	if h.broker.hub != nil {
		if infoVal, ok := h.broker.devices.Load(deviceID); ok {
			info := infoVal.(DeviceInfo)
			logger.Log.Infof("Voice WS push: device_id=%s, user_id=%s, success=%v", deviceID, info.UserID, result.Success)
			h.broker.hub.SendToUser(info.UserID, &ws.Message{
				Type: "voice_down",
				Data: map[string]interface{}{
					"device_id": deviceID,
					"success":   result.Success,
					"message":   result.Message,
					"action":    result.Action,
					"audio_url": result.AudioURL,
				},
			})
		} else {
			logger.Log.Warnf("Voice WS push skipped: device_id=%s not found in devices cache", deviceID)
		}
	}
}

// buildDeviceContext 构建 Dify 设备上下文（JSON 字符串）
// 只包含 voice 模块 exposed 白名单内的属性/服务
func (h *VoiceHandler) buildDeviceContext(ctx context.Context, deviceID, userID string) (string, *model.Device, error) {
	device, err := h.broker.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		return "", nil, fmt.Errorf("设备不存在: %w", err)
	}

	if device.ModelID == nil {
		return "", nil, fmt.Errorf("设备未绑定物模型")
	}

	tm, err := h.broker.thingModelRepo.GetByID(ctx, *device.ModelID)
	if err != nil {
		return "", nil, fmt.Errorf("物模型不存在: %w", err)
	}

	vm := findVoiceModule(tm)
	if vm == nil {
		return "", nil, fmt.Errorf("设备未启用语音模块")
	}

	// 构建过滤后的属性列表
	type propCtx struct {
		ID         string         `json:"id"`
		Name       string         `json:"name"`
		DataType   string         `json:"data_type"`
		AccessMode string         `json:"access_mode"`
		Min        *float64       `json:"min,omitempty"`
		Max        *float64       `json:"max,omitempty"`
		EnumValues []model.EnumValue `json:"enum_values,omitempty"`
	}
	type svcCtx struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	type deviceCtx struct {
		DeviceID   string    `json:"device_id"`
		DeviceName string    `json:"device_name"`
		Properties []propCtx `json:"properties"`
		Services   []svcCtx  `json:"services"`
	}

	dc := deviceCtx{
		DeviceID:   device.ID,
		DeviceName: device.Name,
	}

	for _, p := range tm.Properties {
		if !slotExposed(p.ID, vm.Config.ExposedProperties) {
			continue
		}
		pc := propCtx{
			ID:         p.ID,
			Name:       p.Name,
			DataType:   p.DataType,
			AccessMode: p.AccessMode,
			Min:        p.Min,
			Max:        p.Max,
		}
		if p.DataType == "enum" {
			pc.EnumValues = p.EnumValues
		}
		dc.Properties = append(dc.Properties, pc)
	}

	for _, s := range tm.Services {
		if !slotExposed(s.ID, vm.Config.ExposedServices) {
			continue
		}
		dc.Services = append(dc.Services, svcCtx{ID: s.ID, Name: s.Name})
	}

	b, err := json.Marshal(dc)
	if err != nil {
		return "", nil, fmt.Errorf("marshal device context: %w", err)
	}
	return string(b), device, nil
}

// handleWithDify 通过 Dify 工作流处理语音指令
func (h *VoiceHandler) handleWithDify(ctx context.Context, deviceID, text string, info DeviceInfo) *model.VoiceResult {
	settings := h.broker.GetVoiceSettings()

	deviceContext, device, err := h.buildDeviceContext(ctx, deviceID, info.UserID)
	if err != nil {
		logger.Log.Warnf("Voice Dify: buildDeviceContext failed: device_id=%s, err=%v", deviceID, err)
		return &model.VoiceResult{Success: false, Message: "AI 处理失败: " + err.Error()}
	}

	difyCmd, err := ai.CallWorkflow(settings.APIURL, settings.APIKey, deviceContext, text, deviceID)
	if err != nil {
		logger.Log.Errorf("Voice Dify: CallWorkflow failed: device_id=%s, err=%v", deviceID, err)
		return &model.VoiceResult{Success: false, Message: "AI 处理失败: " + err.Error()}
	}

	if difyCmd.Action == "unknown" {
		msg := difyCmd.Message
		if msg == "" {
			msg = "无法识别指令"
		}
		return &model.VoiceResult{Success: false, Message: msg}
	}

	return h.executeDifyCommand(ctx, deviceID, info.UserID, device, difyCmd)
}

// executeDifyCommand 根据 DifyCommand 执行 MQTT 下发
func (h *VoiceHandler) executeDifyCommand(ctx context.Context, deviceID, userID string, device *model.Device, cmd *ai.DifyCommand) *model.VoiceResult {
	switch cmd.Action {
	case "set_property":
		if cmd.PropertyID == "" || cmd.Value == nil {
			return &model.VoiceResult{Success: false, Message: "AI 返回的属性设置参数不完整"}
		}

		topic := fmt.Sprintf("devices/%s/telemetry/down", deviceID)
		payload, _ := json.Marshal(map[string]any{cmd.PropertyID: cmd.Value})

		if err := h.broker.Publish(topic, payload, false, 1); err != nil {
			return &model.VoiceResult{Success: false, Message: "指令下发失败"}
		}

		// 仅模拟设备才直接写入 device_data；真实 MQTT 连接的设备会自行通过 telemetry/up 回传
		if userID != "" && !h.broker.IsClientConnected(deviceID) {
			merged := map[string]any{cmd.PropertyID: cmd.Value}
			if existing, err := h.broker.deviceDataRepo.GetLatestData(ctx, deviceID); err == nil && existing != nil {
				for k, v := range existing.Payload {
					if _, set := merged[k]; !set {
						merged[k] = v
					}
				}
			}
			storeTopic := fmt.Sprintf("devices/%s/telemetry/up", deviceID)
			_ = h.broker.deviceDataRepo.InsertTelemetry(ctx, deviceID, userID, storeTopic, merged, 1, true, nil)
		}

		action := fmt.Sprintf("设置 %s.%s = %v", device.Name, cmd.PropertyID, cmd.Value)
		logger.Log.Infof("Voice Dify executed: %s", action)
		msg := cmd.Message
		if msg == "" {
			msg = "指令已执行"
		}
		return &model.VoiceResult{Success: true, Message: msg, Action: action}

	case "invoke_service":
		if cmd.ServiceID == "" {
			return &model.VoiceResult{Success: false, Message: "AI 返回的服务调用参数不完整"}
		}

		params := cmd.Params
		if params == nil {
			params = map[string]any{}
		}

		topic := fmt.Sprintf("devices/%s/service/invoke", deviceID)
		payload, _ := json.Marshal(map[string]any{
			"id":      fmt.Sprintf("dify_%d", time.Now().UnixMilli()),
			"service": cmd.ServiceID,
			"params":  params,
		})

		if err := h.broker.Publish(topic, payload, false, 1); err != nil {
			return &model.VoiceResult{Success: false, Message: "服务调用失败"}
		}

		action := fmt.Sprintf("调用 %s.%s", device.Name, cmd.ServiceID)
		logger.Log.Infof("Voice Dify executed: %s", action)
		msg := cmd.Message
		if msg == "" {
			msg = "服务已调用"
		}
		return &model.VoiceResult{Success: true, Message: msg, Action: action}

	case "query_status":
		if cmd.PropertyID == "" {
			return &model.VoiceResult{Success: false, Message: "AI 返回的查询参数不完整"}
		}
		data, err := h.broker.deviceDataRepo.GetLatestData(ctx, deviceID)
		if err != nil || data == nil {
			return &model.VoiceResult{Success: false, Message: "无法获取设备数据"}
		}
		val, ok := data.Payload[cmd.PropertyID]
		if !ok {
			return &model.VoiceResult{Success: false, Message: fmt.Sprintf("属性 %s 无数据", cmd.PropertyID)}
		}
		action := fmt.Sprintf("查询 %s.%s = %v", device.Name, cmd.PropertyID, val)
		return &model.VoiceResult{Success: true, Message: fmt.Sprintf("%s 当前值: %v", cmd.PropertyID, val), Action: action}

	default:
		return &model.VoiceResult{Success: false, Message: fmt.Sprintf("未知指令类型: %s", cmd.Action)}
	}
}
