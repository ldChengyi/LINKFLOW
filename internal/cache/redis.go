package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	DeviceOnlineKeyPrefix = "device:online:"
	UserOnlineSetPrefix   = "user:online_devices:"
	DeviceOnlineTTL       = 24 * time.Hour
	SimulatedOnlineTTL    = 5 * time.Minute
)

type Redis struct {
	client *redis.Client
}

func NewRedis(addr, password string, db int) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Redis{client: client}, nil
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func (r *Redis) Close() error {
	return r.client.Close()
}

// Set 设置键值对
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del 删除键
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Keys 按模式匹配键
func (r *Redis) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.client.Keys(ctx, pattern).Result()
}

// Exists 检查键是否存在
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

// SetDeviceOnline 标记设备在线（Pipeline: SET key + SADD user set）
func (r *Redis) SetDeviceOnline(ctx context.Context, deviceID, userID string) error {
	return r.setDeviceOnlineWithTTL(ctx, deviceID, userID, DeviceOnlineTTL)
}

// SetSimulatedOnline 模拟上线（短 TTL，需心跳续期）
func (r *Redis) SetSimulatedOnline(ctx context.Context, deviceID, userID string) error {
	return r.setDeviceOnlineWithTTL(ctx, deviceID, userID, SimulatedOnlineTTL)
}

// RefreshSimulatedOnline 续期模拟在线的 TTL
func (r *Redis) RefreshSimulatedOnline(ctx context.Context, deviceID, userID string) error {
	pipe := r.client.Pipeline()
	pipe.Expire(ctx, DeviceOnlineKeyPrefix+deviceID, SimulatedOnlineTTL)
	pipe.Expire(ctx, UserOnlineSetPrefix+userID, SimulatedOnlineTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *Redis) setDeviceOnlineWithTTL(ctx context.Context, deviceID, userID string, ttl time.Duration) error {
	pipe := r.client.Pipeline()
	pipe.Set(ctx, DeviceOnlineKeyPrefix+deviceID, userID, ttl)
	pipe.SAdd(ctx, UserOnlineSetPrefix+userID, deviceID)
	pipe.Expire(ctx, UserOnlineSetPrefix+userID, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// SetDeviceOffline 标记设备离线（Pipeline: DEL key + SREM user set）
func (r *Redis) SetDeviceOffline(ctx context.Context, deviceID, userID string) error {
	pipe := r.client.Pipeline()
	pipe.Del(ctx, DeviceOnlineKeyPrefix+deviceID)
	pipe.SRem(ctx, UserOnlineSetPrefix+userID, deviceID)
	_, err := pipe.Exec(ctx)
	return err
}

// IsDeviceOnline 检查单个设备是否在线
func (r *Redis) IsDeviceOnline(ctx context.Context, deviceID string) (bool, error) {
	return r.Exists(ctx, DeviceOnlineKeyPrefix+deviceID)
}

// BatchCheckOnline 批量检查设备在线状态，返回 map[deviceID]bool
func (r *Redis) BatchCheckOnline(ctx context.Context, deviceIDs []string) (map[string]bool, error) {
	if len(deviceIDs) == 0 {
		return map[string]bool{}, nil
	}
	pipe := r.client.Pipeline()
	cmds := make(map[string]*redis.IntCmd, len(deviceIDs))
	for _, id := range deviceIDs {
		cmds[id] = pipe.Exists(ctx, DeviceOnlineKeyPrefix+id)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(deviceIDs))
	for id, cmd := range cmds {
		result[id] = cmd.Val() > 0
	}
	return result, nil
}

// CountUserOnlineDevices 获取用户在线设备数量
func (r *Redis) CountUserOnlineDevices(ctx context.Context, userID string) (int64, error) {
	return r.client.SCard(ctx, UserOnlineSetPrefix+userID).Result()
}
