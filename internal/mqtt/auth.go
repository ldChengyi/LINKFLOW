package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"

	"github.com/ldchengyi/linkflow/internal/logger"
	"github.com/ldchengyi/linkflow/internal/model"
	"github.com/ldchengyi/linkflow/internal/ws"
)

// AuthHook 设备认证 + ACL 控制
type AuthHook struct {
	mochi.HookBase
	broker *Broker
}

func (h *AuthHook) ID() string {
	return "linkflow-auth"
}

func (h *AuthHook) Provides(b byte) bool {
	return b == mochi.OnConnectAuthenticate || b == mochi.OnACLCheck
}

// OnConnectAuthenticate 验证设备 clientID + password(device_secret)
func (h *AuthHook) OnConnectAuthenticate(cl *mochi.Client, pk packets.Packet) bool {
	deviceID := cl.ID
	password := string(pk.Connect.Password)

	if deviceID == "" || password == "" {
		logger.Log.Warnf("MQTT auth rejected: empty credentials")
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	device, err := h.broker.deviceRepo.AuthenticateDevice(ctx, deviceID, password)
	if err != nil {
		logger.Log.Warnf("MQTT auth failed: device_id=%s, err=%v", deviceID, err)
		return false
	}

	// 缓存设备信息
	modelID := ""
	if device.ModelID != nil {
		modelID = *device.ModelID
	}
	h.broker.devices.Store(deviceID, DeviceInfo{
		UserID:     device.UserID,
		ModelID:    modelID,
		DeviceName: device.Name,
	})

	// 认证成功后立即标记设备在线（OnConnect 可能先于此执行，所以在这里处理）
	if err := h.broker.rdb.SetDeviceOnline(ctx, deviceID, device.UserID); err != nil {
		logger.Log.Errorf("Failed to set device online in Redis: device_id=%s, err=%v", deviceID, err)
	}
	if err := h.broker.deviceRepo.UpdateDeviceStatus(ctx, deviceID, "online"); err != nil {
		logger.Log.Errorf("Failed to set device online in PG: device_id=%s, err=%v", deviceID, err)
	}

	logger.Log.Infof("MQTT auth success & online: device_id=%s, user_id=%s", deviceID, device.UserID)

	// WebSocket 推送设备上线 + 统计更新
	if h.broker.hub != nil {
		h.broker.hub.SendToUser(device.UserID, &ws.Message{
			Type: "device_status",
			Data: map[string]interface{}{"device_id": deviceID, "device_name": device.Name, "status": "online"},
		})
		go h.broker.pushStats(device.UserID)
	}

	// 异步写入设备上线审计日志
	go func(devID, userID, devName string) {
		aCtx, aCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer aCancel()
		h.broker.auditLogRepo.Create(aCtx, &model.AuditLog{
			UserID:   &userID,
			Category: model.AuditCategoryDevice,
			Action:   "DEVICE_ONLINE",
			Resource: devID,
			Detail:   map[string]any{"device_name": devName},
			CreatedAt: time.Now(),
		})
	}(deviceID, device.UserID, device.Name)

	// 异步检查 pending OTA 任务
	go h.checkPendingOTA(deviceID)

	return true
}

// checkPendingOTA 设备上线时检查是否有待执行的 OTA 任务
func (h *AuthHook) checkPendingOTA(deviceID string) {
	if h.broker.otaTaskRepo == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tasks, err := h.broker.otaTaskRepo.ListPendingByDeviceID(ctx, deviceID)
	if err != nil || len(tasks) == 0 {
		return
	}

	task := tasks[0]
	fw, err := h.broker.firmwareRepo.GetByID(ctx, task.FirmwareID)
	if err != nil {
		logger.Log.Errorf("checkPendingOTA: firmware not found: %v", err)
		return
	}

	downloadURL := fmt.Sprintf("%s/api/firmwares/%s/download", h.broker.baseURL, fw.ID)
	cmd := map[string]interface{}{
		"task_id":  task.ID,
		"version":  fw.Version,
		"url":      downloadURL,
		"checksum": fw.Checksum,
		"size":     fw.FileSize,
	}
	payload, _ := json.Marshal(cmd)
	topic := fmt.Sprintf("devices/%s/ota/down", deviceID)

	if err := h.broker.server.Publish(topic, payload, false, 1); err != nil {
		logger.Log.Errorf("checkPendingOTA: publish failed: %v", err)
		return
	}

	h.broker.otaTaskRepo.UpdateProgress(ctx, task.ID, "pushing", 0, "")
	logger.Log.Infof("OTA task pushed on connect: device=%s, task=%s", deviceID, task.ID)
}

// OnACLCheck 主题访问控制：设备只能访问 devices/{自己的ID}/* 下的 topic
func (h *AuthHook) OnACLCheck(cl *mochi.Client, topic string, write bool) bool {
	deviceID := cl.ID
	prefix := "devices/" + deviceID + "/"

	if !strings.HasPrefix(topic, prefix) {
		logger.Log.Warnf("MQTT ACL denied: device_id=%s, topic=%s", deviceID, topic)
		return false
	}

	// 设备不能发布到 down topic（仅服务端可写）
	if write && (strings.HasSuffix(topic, "/down") || strings.HasSuffix(topic, "/invoke")) {
		logger.Log.Warnf("MQTT ACL denied (write to down topic): device_id=%s, topic=%s", deviceID, topic)
		return false
	}

	return true
}
