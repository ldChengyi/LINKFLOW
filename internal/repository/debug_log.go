package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
)

type DebugLogRepository struct {
	pool *pgxpool.Pool
}

func NewDebugLogRepository(pool *pgxpool.Pool) *DebugLogRepository {
	return &DebugLogRepository{pool: pool}
}

func (r *DebugLogRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.QueryRow(ctx, sql, args...)
	}
	return r.pool.QueryRow(ctx, sql, args...)
}

func (r *DebugLogRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.Query(ctx, sql, args...)
	}
	return r.pool.Query(ctx, sql, args...)
}

func (r *DebugLogRepository) Create(ctx context.Context, log *model.DebugLog) error {
	reqJSON, _ := json.Marshal(log.Request)
	respJSON, _ := json.Marshal(log.Response)

	query := `INSERT INTO debug_logs (user_id, device_id, device_name, connection_type, action_type, request, response, success, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at`

	return r.queryRow(ctx, query, log.UserID, log.DeviceID, log.DeviceName, log.ConnectionType,
		log.ActionType, reqJSON, respJSON, log.Success, log.ErrorMessage).Scan(&log.ID, &log.CreatedAt)
}

func (r *DebugLogRepository) List(ctx context.Context, deviceID string, page, pageSize int) ([]model.DebugLog, int, error) {
	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM debug_logs WHERE device_id = $1`
	var total int
	if err := r.queryRow(ctx, countQuery, deviceID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, user_id, device_id, device_name, connection_type, action_type, request, response, success, error_message, created_at
		FROM debug_logs WHERE device_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.query(ctx, query, deviceID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []model.DebugLog
	for rows.Next() {
		var log model.DebugLog
		var reqJSON, respJSON []byte
		if err := rows.Scan(&log.ID, &log.UserID, &log.DeviceID, &log.DeviceName, &log.ConnectionType,
			&log.ActionType, &reqJSON, &respJSON, &log.Success, &log.ErrorMessage, &log.CreatedAt); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(reqJSON, &log.Request)
		json.Unmarshal(respJSON, &log.Response)
		logs = append(logs, log)
	}
	return logs, total, nil
}
