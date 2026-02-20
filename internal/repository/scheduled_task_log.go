package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
)

type ScheduledTaskLogRepository struct {
	pool *pgxpool.Pool
}

func NewScheduledTaskLogRepository(pool *pgxpool.Pool) *ScheduledTaskLogRepository {
	return &ScheduledTaskLogRepository{pool: pool}
}

func (r *ScheduledTaskLogRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.QueryRow(ctx, sql, args...)
	}
	return r.pool.QueryRow(ctx, sql, args...)
}

func (r *ScheduledTaskLogRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.Query(ctx, sql, args...)
	}
	return r.pool.Query(ctx, sql, args...)
}

// Create 写入执行日志（Scheduler 触发，使用 Admin pool 绕过 RLS）
func (r *ScheduledTaskLogRepository) Create(ctx context.Context, log *model.ScheduledTaskLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO scheduled_task_logs
		    (task_id, user_id, device_id, device_name, task_name, action_type, topic, payload, status, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, log.TaskID, log.UserID, log.DeviceID, log.DeviceName, log.TaskName,
		log.ActionType, log.Topic, log.Payload, log.Status, log.Error)
	return err
}

// List 查询执行日志（分页，可选按 task_id 或 device_id 过滤，RLS 生效）
func (r *ScheduledTaskLogRepository) List(ctx context.Context, offset, limit int, taskID, deviceID string) ([]*model.ScheduledTaskLog, int, error) {
	var total int
	switch {
	case taskID != "":
		err := r.queryRow(ctx, `SELECT COUNT(*) FROM scheduled_task_logs WHERE task_id = $1`, taskID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	case deviceID != "":
		err := r.queryRow(ctx, `SELECT COUNT(*) FROM scheduled_task_logs WHERE device_id = $1`, deviceID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	default:
		err := r.queryRow(ctx, `SELECT COUNT(*) FROM scheduled_task_logs`).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	var (
		q    string
		args []any
	)
	switch {
	case taskID != "":
		q = `SELECT id, task_id, user_id, device_id, device_name, task_name, action_type, topic, payload, status, error, created_at
		     FROM scheduled_task_logs WHERE task_id = $3 ORDER BY created_at DESC OFFSET $1 LIMIT $2`
		args = []any{offset, limit, taskID}
	case deviceID != "":
		q = `SELECT id, task_id, user_id, device_id, device_name, task_name, action_type, topic, payload, status, error, created_at
		     FROM scheduled_task_logs WHERE device_id = $3 ORDER BY created_at DESC OFFSET $1 LIMIT $2`
		args = []any{offset, limit, deviceID}
	default:
		q = `SELECT id, task_id, user_id, device_id, device_name, task_name, action_type, topic, payload, status, error, created_at
		     FROM scheduled_task_logs ORDER BY created_at DESC OFFSET $1 LIMIT $2`
		args = []any{offset, limit}
	}

	rows, err := r.query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*model.ScheduledTaskLog
	for rows.Next() {
		var l model.ScheduledTaskLog
		if err := rows.Scan(
			&l.ID, &l.TaskID, &l.UserID, &l.DeviceID, &l.DeviceName,
			&l.TaskName, &l.ActionType, &l.Topic, &l.Payload,
			&l.Status, &l.Error, &l.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, &l)
	}
	return logs, total, nil
}
