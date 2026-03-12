package mqtt

import (
	"context"
	"fmt"
	"sync"
	"time"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"

	"github.com/ldchengyi/linkflow/internal/cache"
	"github.com/ldchengyi/linkflow/internal/config"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/repository"
	"github.com/ldchengyi/linkflow/internal/tts"
	"github.com/ldchengyi/linkflow/internal/ws"
)

// VoiceSettings 语音控制配置快照
type VoiceSettings struct {
	Mode   string
	APIURL string
	APIKey string
}

// TTSSettings TTS 语音播报配置快照
// Deprecated: use tts.TTSSettings, kept for broker internal use
type TTSSettings = tts.TTSSettings

// DeviceInfo 缓存已连接设备的信息
type DeviceInfo struct {
	UserID     string
	ModelID    string // 可能为空（未绑定物模型）
	DeviceName string
}

// Broker 内嵌 Mochi MQTT 服务器
type Broker struct {
	server          *mochi.Server
	config          config.MQTTConfig
	deviceRepo      *repository.DeviceRepository
	thingModelRepo  *repository.ThingModelRepository
	deviceDataRepo  *repository.DeviceDataRepository
	auditLogRepo    *repository.AuditLogRepository
	alertRuleRepo   *repository.AlertRuleRepository
	alertLogRepo    *repository.AlertLogRepository
	otaTaskRepo     *repository.OTATaskRepository
	firmwareRepo    *repository.FirmwareRepository
	settingsRepo    *repository.SettingsRepository
	svcCallLogRepo  *repository.ServiceCallLogRepository
	rdb             *cache.Redis
	hub             *ws.Hub
	baseURL         string
	ttsService      tts.Service

	// 已连接设备缓存: device_id → DeviceInfo
	devices sync.Map
	// 物模型缓存: model_id → []model.Property
	models sync.Map
	// 告警规则缓存: device_id → []model.AlertRule
	alertRules sync.Map

	// 语音设置缓存（30s TTL）
	settingsMu     sync.RWMutex
	voiceMode      string
	difyAPIURL     string
	difyAPIKey     string
	// TTS 设置缓存
	ttsProvider        string
	ttsDoubaoAppID     string
	ttsDoubaoAccessKey string
	ttsDoubaoResourceID string
	ttsDoubaoSpeakerID string
	settingsLoadAt time.Time
}

// NewBroker 创建 MQTT Broker
func NewBroker(
	cfg config.MQTTConfig,
	deviceRepo *repository.DeviceRepository,
	thingModelRepo *repository.ThingModelRepository,
	deviceDataRepo *repository.DeviceDataRepository,
	auditLogRepo *repository.AuditLogRepository,
	alertRuleRepo *repository.AlertRuleRepository,
	alertLogRepo *repository.AlertLogRepository,
	otaTaskRepo *repository.OTATaskRepository,
	firmwareRepo *repository.FirmwareRepository,
	settingsRepo *repository.SettingsRepository,
	svcCallLogRepo *repository.ServiceCallLogRepository,
	rdb *cache.Redis,
	hub *ws.Hub,
	baseURL string,
	ttsService tts.Service,
) *Broker {
	return &Broker{
		config:         cfg,
		deviceRepo:     deviceRepo,
		thingModelRepo: thingModelRepo,
		deviceDataRepo: deviceDataRepo,
		auditLogRepo:   auditLogRepo,
		alertRuleRepo:  alertRuleRepo,
		alertLogRepo:   alertLogRepo,
		otaTaskRepo:    otaTaskRepo,
		firmwareRepo:   firmwareRepo,
		settingsRepo:   settingsRepo,
		svcCallLogRepo: svcCallLogRepo,
		rdb:            rdb,
		hub:            hub,
		baseURL:        baseURL,
		ttsService:     ttsService,
	}
}

// Start 启动 MQTT broker
func (b *Broker) Start() error {
	b.server = mochi.New(&mochi.Options{
		InlineClient: true,
	})

	// 注册认证 + ACL hook
	if err := b.server.AddHook(&AuthHook{broker: b}, nil); err != nil {
		return fmt.Errorf("add auth hook: %w", err)
	}

	// 注册连接管理 + 消息处理 hook
	if err := b.server.AddHook(&EventHook{broker: b}, nil); err != nil {
		return fmt.Errorf("add event hook: %w", err)
	}

	// TCP listener
	addr := fmt.Sprintf("%s:%s", b.config.Host, b.config.Port)
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "tcp",
		Address: addr,
	})
	if err := b.server.AddListener(tcp); err != nil {
		return fmt.Errorf("add tcp listener: %w", err)
	}

	go func() {
		if err := b.server.Serve(); err != nil {
			logger.Log.Errorf("MQTT broker error: %v", err)
		}
	}()

	logger.Log.Infof("MQTT broker started on %s", addr)
	return nil
}

// Stop 优雅关闭
func (b *Broker) Stop(_ context.Context) error {
	if b.server == nil {
		return nil
	}
	logger.Log.Info("Stopping MQTT broker...")
	return b.server.Close()
}

// pushStats 推送统计更新给用户
func (b *Broker) pushStats(userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	onlineCount, err := b.rdb.CountUserOnlineDevices(ctx, userID)
	if err != nil {
		logger.Log.Errorf("pushStats: count online devices failed: %v", err)
		return
	}

	b.hub.SendToUser(userID, &ws.Message{
		Type: "stats",
		Data: map[string]interface{}{"online_devices": onlineCount},
	})
}

// InvalidateAlertRulesCache 清除指定设备的告警规则缓存
func (b *Broker) InvalidateAlertRulesCache(deviceID string) {
	b.alertRules.Delete(deviceID)
}

// IsClientConnected 检查设备是否有真实 MQTT 连接
func (b *Broker) IsClientConnected(clientID string) bool {
	if b.server == nil {
		return false
	}
	cl, ok := b.server.Clients.Get(clientID)
	return ok && !cl.Closed()
}

// Publish 服务端下发消息
func (b *Broker) Publish(topic string, payload []byte, retain bool, qos byte) error {
	if b.server == nil {
		return fmt.Errorf("mqtt server not started")
	}
	return b.server.Publish(topic, payload, retain, qos)
}

// loadSettingsLocked 从 DB 加载所有 settings 到缓存（调用方需持有写锁）
func (b *Broker) loadSettingsLocked() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if b.settingsRepo != nil {
		kv, err := b.settingsRepo.GetAll(ctx)
		if err != nil {
			logger.Log.Errorf("loadSettingsLocked: %v", err)
			return
		}
		b.voiceMode = kv["voice_mode"]
		b.difyAPIURL = kv["dify_api_url"]
		b.difyAPIKey = kv["dify_api_key"]
		b.ttsProvider = kv["tts_provider"]
		b.ttsDoubaoAppID = kv["tts_doubao_app_id"]
		b.ttsDoubaoAccessKey = kv["tts_doubao_access_key"]
		b.ttsDoubaoResourceID = kv["tts_doubao_resource_id"]
		b.ttsDoubaoSpeakerID = kv["tts_doubao_speaker_id"]
		b.settingsLoadAt = time.Now()
	}

	if b.voiceMode == "" {
		b.voiceMode = "local"
	}
	if b.ttsProvider == "" {
		b.ttsProvider = "edge"
	}
}

// ensureSettingsLoaded 确保 settings 缓存有效（30s TTL + double-check）
func (b *Broker) ensureSettingsLoaded() {
	const ttl = 30 * time.Second

	b.settingsMu.RLock()
	if !b.settingsLoadAt.IsZero() && time.Since(b.settingsLoadAt) < ttl {
		b.settingsMu.RUnlock()
		return
	}
	b.settingsMu.RUnlock()

	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()

	// double-check
	if !b.settingsLoadAt.IsZero() && time.Since(b.settingsLoadAt) < ttl {
		return
	}
	b.loadSettingsLocked()
}

// GetVoiceSettings 获取语音设置（30s TTL 内存缓存，过期从 DB 重载）
func (b *Broker) GetVoiceSettings() VoiceSettings {
	b.ensureSettingsLoaded()
	b.settingsMu.RLock()
	defer b.settingsMu.RUnlock()
	return VoiceSettings{Mode: b.voiceMode, APIURL: b.difyAPIURL, APIKey: b.difyAPIKey}
}

// GetTTSSettings 获取 TTS 语音播报设置（30s TTL 内存缓存）
func (b *Broker) GetTTSSettings() TTSSettings {
	b.ensureSettingsLoaded()
	b.settingsMu.RLock()
	defer b.settingsMu.RUnlock()
	return TTSSettings{
		Provider:   b.ttsProvider,
		AppID:      b.ttsDoubaoAppID,
		AccessKey:  b.ttsDoubaoAccessKey,
		ResourceID: b.ttsDoubaoResourceID,
		SpeakerID:  b.ttsDoubaoSpeakerID,
	}
}

// InvalidateSettingsCache 清除语音设置缓存，下次调用时重新从 DB 加载
func (b *Broker) InvalidateSettingsCache() {
	b.settingsMu.Lock()
	b.settingsLoadAt = time.Time{}
	b.settingsMu.Unlock()
}

// SetTTSService 设置 TTS 服务（用于解决循环依赖：broker 创建后再注入）
func (b *Broker) SetTTSService(svc tts.Service) {
	b.ttsService = svc
}
