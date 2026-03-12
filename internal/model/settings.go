package model

// PlatformSettings 平台全局设置
type PlatformSettings struct {
	VoiceMode  string `json:"voice_mode"`    // "local" | "dify"
	DifyAPIURL string `json:"dify_api_url"`
	DifyAPIKey string `json:"dify_api_key"`  // GET 时返回掩码 "****xxxx"

	// TTS 语音播报
	TTSProvider        string `json:"tts_provider"`          // "edge" | "doubao"
	TTSDoubaoAppID     string `json:"tts_doubao_app_id"`
	TTSDoubaoAccessKey string `json:"tts_doubao_access_key"` // GET 时返回掩码
	TTSDoubaoResourceID string `json:"tts_doubao_resource_id"`
	TTSDoubaoSpeakerID string `json:"tts_doubao_speaker_id"`
}

// UpdateSettingsRequest 更新设置请求
type UpdateSettingsRequest struct {
	VoiceMode  string `json:"voice_mode"`
	DifyAPIURL string `json:"dify_api_url"`
	DifyAPIKey string `json:"dify_api_key"`  // 空或掩码（****）时不更新

	// TTS 语音播报
	TTSProvider        string `json:"tts_provider"`
	TTSDoubaoAppID     string `json:"tts_doubao_app_id"`
	TTSDoubaoAccessKey string `json:"tts_doubao_access_key"` // 空或掩码时不更新
	TTSDoubaoResourceID string `json:"tts_doubao_resource_id"`
	TTSDoubaoSpeakerID string `json:"tts_doubao_speaker_id"`
}
