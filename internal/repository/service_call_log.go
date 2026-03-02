package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
)

type ServiceCallLogRepository struct {
	pool *pgxpool.Pool
}

func NewServiceCallLogRepository(pool *pgxpool.Pool) *ServiceCallLogRepository {
	return &ServiceCallLogRepository{pool: pool}
}

func (r *ServiceCallLogRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.QueryRow(ctx, sql, args...)
	}
	return r.pool.QueryRow(ctx, sql, args...)
}

func (r *ServiceCallLogRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.Query(ctx, sql, args...)
	}
	return r.pool.Query(ctx, sql, args...)
}

// Create 写入服务调用日志（使用 Admin pool 绕过 RLS）
func (r *ServiceCallLogRepository) Create(ctx context.Context, log *model.ServiceCallLog) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO service_call_logs
		    (device_id, user_id, device_name, service_id, service_name, request_id, input_params, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, log.DeviceID, log.UserID, log.DeviceName, log.ServiceID, log.ServiceName,
		log.RequestID, log.InputParams, log.Status).Scan(&id)
	return id, err
}

// UpdateReply 更新服务调用响应（Broker 收到 reply 时调用，Admin pool）
func (r *ServiceCallLogRepository) UpdateReply(ctx context.Context, requestID string, outputParams []byte, code int, status, errMsg string) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE service_call_logs
		SET output_params = $1, response_code = $2, status = $3, error = $4, replied_at = $5
		WHERE request_id = $6 AND status = 'pending'
	`, outputParams, code, status, errMsg, now, requestID)
	return err
}

// List 分页查询服务调用日志（支持 device_id 筛选，RLS 生效）
func (r *ServiceCallLogRepository) List(ctx context.Context, offset, limit int, deviceID string) ([]*model.ServiceCallLog, int, error) {
	var total int
	if deviceID != "" {
		err := r.queryRow(ctx, `SELECT COUNT(*) FROM service_call_logs WHERE device_id = $1`, deviceID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.queryRow(ctx, `SELECT COUNT(*) FROM service_call_logs`).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	var (
		q    string
		args []any
	)
	if deviceID != "" {
		q = `SELECT id, device_id, user_id, device_name, service_id, service_name, request_id,
		            input_params, output_params, status, error, response_code, created_at, replied_at
		     FROM service_call_logs WHERE device_id = $3 ORDER BY created_at DESC OFFSET $1 LIMIT $2`
		args = []any{offset, limit, deviceID}
	} else {
		q = `SELECT id, device_id, user_id, device_name, service_id, service_name, request_id,
		            input_params, output_params, status, error, response_code, created_at, replied_at
		     FROM service_call_logs ORDER BY created_at DESC OFFSET $1 LIMIT $2`
		args = []any{offset, limit}
	}

	rows, err := r.query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*model.ServiceCallLog
	for rows.Next() {
		var l model.ServiceCallLog
		if err := rows.Scan(
			&l.ID, &l.DeviceID, &l.UserID, &l.DeviceName, &l.ServiceID, &l.ServiceName,
			&l.RequestID, &l.InputParams, &l.OutputParams, &l.Status, &l.Error,
			&l.ResponseCode, &l.CreatedAt, &l.RepliedAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, &l)
	}
	return logs, total, nil
}

// GetByRequestID 根据 requestID 查询（Broker 内部使用，Admin pool）
func (r *ServiceCallLogRepository) GetByRequestID(ctx context.Context, requestID string) (*model.ServiceCallLog, error) {
	var l model.ServiceCallLog
	err := r.pool.QueryRow(ctx, `
		SELECT id, device_id, user_id, device_name, service_id, service_name, request_id,
		       input_params, output_params, status, error, response_code, created_at, replied_at
		FROM service_call_logs WHERE request_id = $1
	`, requestID).Scan(
		&l.ID, &l.DeviceID, &l.UserID, &l.DeviceName, &l.ServiceID, &l.ServiceName,
		&l.RequestID, &l.InputParams, &l.OutputParams, &l.Status, &l.Error,
		&l.ResponseCode, &l.CreatedAt, &l.RepliedAt,
	)
	if err != nil {
		return nil, err
	}
	return &l, nil
}
