package model

// PlatformSettings 平台全局设置
type PlatformSettings struct {
	VoiceMode  string `json:"voice_mode"`    // "local" | "dify"
	DifyAPIURL string `json:"dify_api_url"`
	DifyAPIKey string `json:"dify_api_key"`  // GET 时返回掩码 "****xxxx"
}

// UpdateSettingsRequest 更新设置请求
type UpdateSettingsRequest struct {
	VoiceMode  string `json:"voice_mode"`
	DifyAPIURL string `json:"dify_api_url"`
	DifyAPIKey string `json:"dify_api_key"`  // 空或掩码（****）时不更新
}
