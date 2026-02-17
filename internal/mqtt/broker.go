package mqtt

import (
	"context"
	"fmt"
	"sync"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"

	"github.com/ldchengyi/linkflow/internal/cache"
	"github.com/ldchengyi/linkflow/internal/config"
	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/repository"
)

// DeviceInfo 缓存已连接设备的信息
type DeviceInfo struct {
	UserID  string
	ModelID string // 可能为空（未绑定物模型）
}

// Broker 内嵌 Mochi MQTT 服务器
type Broker struct {
	server         *mochi.Server
	config         config.MQTTConfig
	deviceRepo     *repository.DeviceRepository
	thingModelRepo *repository.ThingModelRepository
	deviceDataRepo *repository.DeviceDataRepository
	rdb            *cache.Redis

	// 已连接设备缓存: device_id → DeviceInfo
	devices sync.Map
	// 物模型缓存: model_id → []model.Property
	models sync.Map
}

// NewBroker 创建 MQTT Broker
func NewBroker(
	cfg config.MQTTConfig,
	deviceRepo *repository.DeviceRepository,
	thingModelRepo *repository.ThingModelRepository,
	deviceDataRepo *repository.DeviceDataRepository,
	rdb *cache.Redis,
) *Broker {
	return &Broker{
		config:         cfg,
		deviceRepo:     deviceRepo,
		thingModelRepo: thingModelRepo,
		deviceDataRepo: deviceDataRepo,
		rdb:            rdb,
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

// Publish 服务端下发消息
func (b *Broker) Publish(topic string, payload []byte, retain bool, qos byte) error {
	if b.server == nil {
		return fmt.Errorf("mqtt server not started")
	}
	return b.server.Publish(topic, payload, retain, qos)
}
