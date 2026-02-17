package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
)

type DeviceRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceRepository(pool *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{pool: pool}
}

func (r *DeviceRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.QueryRow(ctx, sql, args...)
	}
	return r.pool.QueryRow(ctx, sql, args...)
}

func (r *DeviceRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.Query(ctx, sql, args...)
	}
	return r.pool.Query(ctx, sql, args...)
}

func (r *DeviceRepository) exec(ctx context.Context, sql string, args ...any) error {
	if conn := database.RLSConn(ctx); conn != nil {
		_, err := conn.Exec(ctx, sql, args...)
		return err
	}
	_, err := r.pool.Exec(ctx, sql, args...)
	return err
}

func generateSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate device secret: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func nullUUID(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Create 创建设备
func (r *DeviceRepository) Create(ctx context.Context, userID string, req *model.CreateDeviceRequest) (*model.Device, error) {
	secret, err := generateSecret()
	if err != nil {
		return nil, err
	}

	metadata, _ := json.Marshal(req.Metadata)
	if req.Metadata == nil {
		metadata = []byte("{}")
	}

	var db model.DeviceDB
	err = r.queryRow(ctx, `
		INSERT INTO devices (user_id, model_id, name, device_secret, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, model_id, name, device_secret, status, last_online_at, metadata, created_at, updated_at
	`, userID, nullUUID(req.ModelID), req.Name, secret, metadata).Scan(
		&db.ID, &db.UserID, &db.ModelID, &db.Name, &db.DeviceSecret,
		&db.Status, &db.LastOnlineAt, &db.Metadata, &db.CreatedAt, &db.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return db.ToModel()
}

// GetByID 根据ID获取设备
func (r *DeviceRepository) GetByID(ctx context.Context, id string) (*model.Device, error) {
	var db model.DeviceDB
	err := r.queryRow(ctx, `
		SELECT d.id, d.user_id, d.model_id, tm.name AS model_name,
		       d.name, d.device_secret, d.status, d.last_online_at, d.metadata,
		       d.created_at, d.updated_at
		FROM devices d
		LEFT JOIN thing_models tm ON d.model_id = tm.id
		WHERE d.id = $1
	`, id).Scan(
		&db.ID, &db.UserID, &db.ModelID, &db.ModelName,
		&db.Name, &db.DeviceSecret, &db.Status, &db.LastOnlineAt, &db.Metadata,
		&db.CreatedAt, &db.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return db.ToModel()
}

// List 获取设备列表
func (r *DeviceRepository) List(ctx context.Context, offset, limit int) ([]*model.Device, int, error) {
	var total int
	err := r.queryRow(ctx, `SELECT COUNT(*) FROM devices`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.query(ctx, `
		SELECT d.id, d.user_id, d.model_id, tm.name AS model_name,
		       d.name, d.device_secret, d.status, d.last_online_at, d.metadata,
		       d.created_at, d.updated_at
		FROM devices d
		LEFT JOIN thing_models tm ON d.model_id = tm.id
		ORDER BY d.created_at DESC
		OFFSET $1 LIMIT $2
	`, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var devices []*model.Device
	for rows.Next() {
		var db model.DeviceDB
		err := rows.Scan(
			&db.ID, &db.UserID, &db.ModelID, &db.ModelName,
			&db.Name, &db.DeviceSecret, &db.Status, &db.LastOnlineAt, &db.Metadata,
			&db.CreatedAt, &db.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		d, err := db.ToModel()
		if err != nil {
			return nil, 0, err
		}
		devices = append(devices, d)
	}
	return devices, total, nil
}

// Update 更新设备
func (r *DeviceRepository) Update(ctx context.Context, id string, req *model.UpdateDeviceRequest) (*model.Device, error) {
	metadata, _ := json.Marshal(req.Metadata)
	if req.Metadata == nil {
		metadata = []byte("{}")
	}

	var db model.DeviceDB
	err := r.queryRow(ctx, `
		UPDATE devices
		SET name = $2, model_id = $3, metadata = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, model_id, name, device_secret, status, last_online_at, metadata, created_at, updated_at
	`, id, req.Name, nullUUID(req.ModelID), metadata).Scan(
		&db.ID, &db.UserID, &db.ModelID, &db.Name, &db.DeviceSecret,
		&db.Status, &db.LastOnlineAt, &db.Metadata, &db.CreatedAt, &db.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return db.ToModel()
}

// Delete 删除设备
func (r *DeviceRepository) Delete(ctx context.Context, id string) error {
	return r.exec(ctx, `DELETE FROM devices WHERE id = $1`, id)
}

// AuthenticateDevice 验证设备凭证（MQTT 认证用，直接用 pool 绕过 RLS）
func (r *DeviceRepository) AuthenticateDevice(ctx context.Context, deviceID, deviceSecret string) (*model.Device, error) {
	var db model.DeviceDB
	err := r.pool.QueryRow(ctx, `
		SELECT d.id, d.user_id, d.model_id, tm.name AS model_name,
		       d.name, d.device_secret, d.status, d.last_online_at, d.metadata,
		       d.created_at, d.updated_at
		FROM devices d
		LEFT JOIN thing_models tm ON d.model_id = tm.id
		WHERE d.id = $1 AND d.device_secret = $2
	`, deviceID, deviceSecret).Scan(
		&db.ID, &db.UserID, &db.ModelID, &db.ModelName,
		&db.Name, &db.DeviceSecret, &db.Status, &db.LastOnlineAt, &db.Metadata,
		&db.CreatedAt, &db.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return db.ToModel()
}

// UpdateDeviceStatus 更新设备在线状态（MQTT 连接/断开时调用）
func (r *DeviceRepository) UpdateDeviceStatus(ctx context.Context, deviceID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE devices SET status = $2, last_online_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, deviceID, status)
	return err
}

// Count 获取设备总数（受 RLS 约束）
func (r *DeviceRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.queryRow(ctx, `SELECT COUNT(*) FROM devices`).Scan(&count)
	return count, err
}

// ListByUserID 获取用户的所有设备（语音模块用，绕过 RLS）
func (r *DeviceRepository) ListByUserID(ctx context.Context, userID string) ([]*model.Device, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT d.id, d.user_id, d.model_id, tm.name AS model_name,
		       d.name, d.device_secret, d.status, d.last_online_at, d.metadata,
		       d.created_at, d.updated_at
		FROM devices d
		LEFT JOIN thing_models tm ON d.model_id = tm.id
		WHERE d.user_id = $1
		ORDER BY d.created_at
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*model.Device
	for rows.Next() {
		var db model.DeviceDB
		err := rows.Scan(
			&db.ID, &db.UserID, &db.ModelID, &db.ModelName,
			&db.Name, &db.DeviceSecret, &db.Status, &db.LastOnlineAt, &db.Metadata,
			&db.CreatedAt, &db.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		d, err := db.ToModel()
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, nil
}
