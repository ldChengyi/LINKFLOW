package tts

import "context"

// Service TTS 服务接口
type Service interface {
	// Synthesize 文本转语音，返回音频文件相对路径（如 "abc123.mp3"）
	Synthesize(ctx context.Context, text string) (string, error)
}

// TTSSettingsProvider 动态获取 TTS 配置（由 Broker 实现）
type TTSSettingsProvider interface {
	GetTTSSettings() TTSSettings
}

// TTSSettings TTS 配置快照
type TTSSettings struct {
	Provider   string
	AppID      string
	AccessKey  string
	ResourceID string
	SpeakerID  string
}
