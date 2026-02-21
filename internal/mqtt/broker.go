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
	"github.com/ldchengyi/linkflow/internal/ws"
)

// VoiceSettings 语音控制配置快照
type VoiceSettings struct {
	Mode   string
	APIURL string
	APIKey string
}

// DeviceInfo 缓存已连接设备的信息
type DeviceInfo struct {
	UserID     string
	ModelID    string // 可能为空（未绑定物模型）
	DeviceName string
}

// Broker 内嵌 Mochi MQTT 服务器
type Broker struct {
	server         *mochi.Server
	config         config.MQTTConfig
	deviceRepo     *repository.DeviceRepository
	thingModelRepo *repository.ThingModelRepository
	deviceDataRepo *repository.DeviceDataRepository
	auditLogRepo   *repository.AuditLogRepository
	alertRuleRepo  *repository.AlertRuleRepository
	alertLogRepo   *repository.AlertLogRepository
	otaTaskRepo    *repository.OTATaskRepository
	firmwareRepo   *repository.FirmwareRepository
	settingsRepo   *repository.SettingsRepository
	rdb            *cache.Redis
	hub            *ws.Hub
	baseURL        string

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
	rdb *cache.Redis,
	hub *ws.Hub,
	baseURL string,
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
		rdb:            rdb,
		hub:            hub,
		baseURL:        baseURL,
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

// GetVoiceSettings 获取语音设置（30s TTL 内存缓存，过期从 DB 重载）
func (b *Broker) GetVoiceSettings() VoiceSettings {
	const ttl = 30 * time.Second

	b.settingsMu.RLock()
	if !b.settingsLoadAt.IsZero() && time.Since(b.settingsLoadAt) < ttl {
		s := VoiceSettings{Mode: b.voiceMode, APIURL: b.difyAPIURL, APIKey: b.difyAPIKey}
		b.settingsMu.RUnlock()
		return s
	}
	b.settingsMu.RUnlock()

	// 重新加载
	b.settingsMu.Lock()
	defer b.settingsMu.Unlock()

	// double-check
	if !b.settingsLoadAt.IsZero() && time.Since(b.settingsLoadAt) < ttl {
		return VoiceSettings{Mode: b.voiceMode, APIURL: b.difyAPIURL, APIKey: b.difyAPIKey}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if b.settingsRepo != nil {
		kv, err := b.settingsRepo.GetAll(ctx)
		if err != nil {
			logger.Log.Errorf("GetVoiceSettings: %v", err)
		} else {
			b.voiceMode = kv["voice_mode"]
			b.difyAPIURL = kv["dify_api_url"]
			b.difyAPIKey = kv["dify_api_key"]
			b.settingsLoadAt = time.Now()
		}
	}

	if b.voiceMode == "" {
		b.voiceMode = "local"
	}
	return VoiceSettings{Mode: b.voiceMode, APIURL: b.difyAPIURL, APIKey: b.difyAPIKey}
}

// InvalidateSettingsCache 清除语音设置缓存，下次调用时重新从 DB 加载
func (b *Broker) InvalidateSettingsCache() {
	b.settingsMu.Lock()
	b.settingsLoadAt = time.Time{}
	b.settingsMu.Unlock()
}
