package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
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

	logger.Log.Infof("Voice command received: device_id=%s, text=%s", deviceID, cmd.Text)

	// 获取发送指令的设备信息
	infoVal, ok := h.broker.devices.Load(deviceID)
	if !ok {
		h.reply(deviceID, &model.VoiceResult{Success: false, Message: "设备信息未找到"})
		return
	}
	info := infoVal.(DeviceInfo)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 构建管道上下文并执行
	pc := &PipelineContext{
		RawText:  cmd.Text,
		DeviceID: deviceID,
		UserID:   info.UserID,
	}

	result := h.pipeline.Run(ctx, pc)
	h.reply(deviceID, result)
}

// reply 回复语音指令结果
func (h *VoiceHandler) reply(deviceID string, result *model.VoiceResult) {
	topic := fmt.Sprintf("devices/%s/voice/down", deviceID)
	payload, _ := json.Marshal(result)
	if err := h.broker.Publish(topic, payload, false, 1); err != nil {
		logger.Log.Errorf("Voice: failed to reply: device_id=%s, err=%v", deviceID, err)
	}
}
