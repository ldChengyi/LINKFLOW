package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
)

type ScheduledTaskRepository struct {
	pool *pgxpool.Pool
}

func NewScheduledTaskRepository(pool *pgxpool.Pool) *ScheduledTaskRepository {
	return &ScheduledTaskRepository{pool: pool}
}

func (r *ScheduledTaskRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.QueryRow(ctx, sql, args...)
	}
	return r.pool.QueryRow(ctx, sql, args...)
}

func (r *ScheduledTaskRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.Query(ctx, sql, args...)
	}
	return r.pool.Query(ctx, sql, args...)
}

func (r *ScheduledTaskRepository) exec(ctx context.Context, sql string, args ...any) error {
	if conn := database.RLSConn(ctx); conn != nil {
		_, err := conn.Exec(ctx, sql, args...)
		return err
	}
	_, err := r.pool.Exec(ctx, sql, args...)
	return err
}

func scanTask(row pgx.Row) (*model.ScheduledTask, error) {
	var t model.ScheduledTask
	err := row.Scan(
		&t.ID, &t.UserID, &t.DeviceID, &t.DeviceName, &t.Name, &t.CronExpr,
		&t.ActionType, &t.PropertyID, &t.ServiceID, &t.Value,
		&t.Enabled, &t.LastRunAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

const taskSelectFields = `t.id, t.user_id, t.device_id, COALESCE(d.name,'') AS device_name,
	t.name, t.cron_expr, t.action_type, COALESCE(t.property_id,''), COALESCE(t.service_id,''),
	t.value, t.enabled, t.last_run_at, t.created_at, t.updated_at`

func (r *ScheduledTaskRepository) Create(ctx context.Context, userID string, req *model.CreateScheduledTaskRequest) (*model.ScheduledTask, error) {
	return scanTask(r.queryRow(ctx, `
		INSERT INTO scheduled_tasks (user_id, device_id, name, cron_expr, action_type, property_id, service_id, value, enabled)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, user_id, device_id, '', name, cron_expr, action_type, COALESCE(property_id,''), COALESCE(service_id,''),
		          value, enabled, last_run_at, created_at, updated_at
	`, userID, req.DeviceID, req.Name, req.CronExpr, req.ActionType, req.PropertyID, req.ServiceID, req.Value, req.Enabled))
}

func (r *ScheduledTaskRepository) GetByID(ctx context.Context, id string) (*model.ScheduledTask, error) {
	return scanTask(r.queryRow(ctx, `
		SELECT `+taskSelectFields+`
		FROM scheduled_tasks t LEFT JOIN devices d ON t.device_id = d.id
		WHERE t.id = $1
	`, id))
}

func (r *ScheduledTaskRepository) List(ctx context.Context, offset, limit int, deviceID string) ([]*model.ScheduledTask, int, error) {
	var total int
	if deviceID != "" {
		if err := r.queryRow(ctx, `SELECT COUNT(*) FROM scheduled_tasks WHERE device_id=$1`, deviceID).Scan(&total); err != nil {
			return nil, 0, err
		}
	} else {
		if err := r.queryRow(ctx, `SELECT COUNT(*) FROM scheduled_tasks`).Scan(&total); err != nil {
			return nil, 0, err
		}
	}

	var q string
	var args []any
	base := `SELECT ` + taskSelectFields + ` FROM scheduled_tasks t LEFT JOIN devices d ON t.device_id = d.id`
	if deviceID != "" {
		q = base + ` WHERE t.device_id=$3 ORDER BY t.created_at DESC OFFSET $1 LIMIT $2`
		args = []any{offset, limit, deviceID}
	} else {
		q = base + ` ORDER BY t.created_at DESC OFFSET $1 LIMIT $2`
		args = []any{offset, limit}
	}

	rows, err := r.query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*model.ScheduledTask
	for rows.Next() {
		var t model.ScheduledTask
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.DeviceID, &t.DeviceName, &t.Name, &t.CronExpr,
			&t.ActionType, &t.PropertyID, &t.ServiceID, &t.Value,
			&t.Enabled, &t.LastRunAt, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, total, nil
}

func (r *ScheduledTaskRepository) Update(ctx context.Context, id string, req *model.UpdateScheduledTaskRequest) (*model.ScheduledTask, error) {
	return scanTask(r.queryRow(ctx, `
		UPDATE scheduled_tasks
		SET device_id=$2, name=$3, cron_expr=$4, action_type=$5, property_id=$6, service_id=$7, value=$8, enabled=$9, updated_at=NOW()
		WHERE id=$1
		RETURNING id, user_id, device_id, '', name, cron_expr, action_type, COALESCE(property_id,''), COALESCE(service_id,''),
		          value, enabled, last_run_at, created_at, updated_at
	`, id, req.DeviceID, req.Name, req.CronExpr, req.ActionType, req.PropertyID, req.ServiceID, req.Value, req.Enabled))
}

func (r *ScheduledTaskRepository) Delete(ctx context.Context, id string) error {
	return r.exec(ctx, `DELETE FROM scheduled_tasks WHERE id=$1`, id)
}

// ListAllEnabled 获取所有启用的任务（Scheduler 用，Admin pool 绕过 RLS）
func (r *ScheduledTaskRepository) ListAllEnabled(ctx context.Context) ([]*model.ScheduledTask, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+taskSelectFields+`
		FROM scheduled_tasks t LEFT JOIN devices d ON t.device_id = d.id
		WHERE t.enabled = true
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.ScheduledTask
	for rows.Next() {
		var t model.ScheduledTask
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.DeviceID, &t.DeviceName, &t.Name, &t.CronExpr,
			&t.ActionType, &t.PropertyID, &t.ServiceID, &t.Value,
			&t.Enabled, &t.LastRunAt, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

func (r *ScheduledTaskRepository) UpdateLastRunAt(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE scheduled_tasks SET last_run_at=NOW() WHERE id=$1`, id)
	return err
}
