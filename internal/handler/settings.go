package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/repository"
)

// SettingsCacheInvalidator Broker 实现此接口以使 settings 缓存失效
type SettingsCacheInvalidator interface {
	InvalidateSettingsCache()
}

// SettingsHandler 平台设置处理器
type SettingsHandler struct {
	repo        *repository.SettingsRepository
	invalidator SettingsCacheInvalidator
}

// NewSettingsHandler 创建设置处理器
func NewSettingsHandler(repo *repository.SettingsRepository, invalidator SettingsCacheInvalidator) *SettingsHandler {
	return &SettingsHandler{repo: repo, invalidator: invalidator}
}

// Get GET /api/settings — 返回平台设置（API Key 掩码）
func (h *SettingsHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	kv, err := h.repo.GetAll(ctx)
	if err != nil {
		logger.Log.Errorf("SettingsHandler.Get: %v", err)
		Fail(c, http.StatusInternalServerError, "failed to load settings")
		return
	}

	settings := &model.PlatformSettings{
		VoiceMode:           kv["voice_mode"],
		DifyAPIURL:          kv["dify_api_url"],
		DifyAPIKey:          maskAPIKey(kv["dify_api_key"]),
		TTSProvider:         kv["tts_provider"],
		TTSDoubaoAppID:      kv["tts_doubao_app_id"],
		TTSDoubaoAccessKey:  maskAPIKey(kv["tts_doubao_access_key"]),
		TTSDoubaoResourceID: kv["tts_doubao_resource_id"],
		TTSDoubaoSpeakerID:  kv["tts_doubao_speaker_id"],
	}
	Success(c, settings)
}

// Update PUT /api/settings — 更新平台设置
func (h *SettingsHandler) Update(c *gin.Context) {
	var req model.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	ctx := c.Request.Context()

	// 只更新非空且非掩码的字段
	toUpdate := map[string]string{}

	if req.VoiceMode == "local" || req.VoiceMode == "dify" {
		toUpdate["voice_mode"] = req.VoiceMode
	}
	if req.DifyAPIURL != "" {
		toUpdate["dify_api_url"] = req.DifyAPIURL
	}
	if req.DifyAPIKey != "" && !isMasked(req.DifyAPIKey) {
		toUpdate["dify_api_key"] = req.DifyAPIKey
	}

	// TTS 设置
	if req.TTSProvider == "edge" || req.TTSProvider == "doubao" {
		toUpdate["tts_provider"] = req.TTSProvider
	}
	if req.TTSDoubaoAppID != "" {
		toUpdate["tts_doubao_app_id"] = req.TTSDoubaoAppID
	}
	if req.TTSDoubaoAccessKey != "" && !isMasked(req.TTSDoubaoAccessKey) {
		toUpdate["tts_doubao_access_key"] = req.TTSDoubaoAccessKey
	}
	if req.TTSDoubaoResourceID != "" {
		toUpdate["tts_doubao_resource_id"] = req.TTSDoubaoResourceID
	}
	if req.TTSDoubaoSpeakerID != "" {
		toUpdate["tts_doubao_speaker_id"] = req.TTSDoubaoSpeakerID
	}

	if len(toUpdate) > 0 {
		if err := h.repo.SetMany(ctx, toUpdate); err != nil {
			logger.Log.Errorf("SettingsHandler.Update: %v", err)
			Fail(c, http.StatusInternalServerError, "failed to save settings")
			return
		}
		// 通知 Broker 清除缓存
		h.invalidator.InvalidateSettingsCache()
	}

	// 返回更新后的设置
	kv, err := h.repo.GetAll(ctx)
	if err != nil {
		logger.Log.Errorf("SettingsHandler.Update reload: %v", err)
		Fail(c, http.StatusInternalServerError, "failed to reload settings")
		return
	}
	settings := &model.PlatformSettings{
		VoiceMode:           kv["voice_mode"],
		DifyAPIURL:          kv["dify_api_url"],
		DifyAPIKey:          maskAPIKey(kv["dify_api_key"]),
		TTSProvider:         kv["tts_provider"],
		TTSDoubaoAppID:      kv["tts_doubao_app_id"],
		TTSDoubaoAccessKey:  maskAPIKey(kv["tts_doubao_access_key"]),
		TTSDoubaoResourceID: kv["tts_doubao_resource_id"],
		TTSDoubaoSpeakerID:  kv["tts_doubao_speaker_id"],
	}
	Success(c, settings)
}

// maskAPIKey 对 API Key 做掩码处理（保留后4位）
func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 4 {
		return "****"
	}
	return "****" + key[len(key)-4:]
}

// isMasked 检查是否为掩码值（以 **** 开头）
func isMasked(key string) bool {
	return strings.HasPrefix(key, "****")
}
