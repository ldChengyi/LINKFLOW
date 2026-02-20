package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
)

type OTATaskRepository struct {
	pool *pgxpool.Pool
}

func NewOTATaskRepository(pool *pgxpool.Pool) *OTATaskRepository {
	return &OTATaskRepository{pool: pool}
}

func (r *OTATaskRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.QueryRow(ctx, sql, args...)
	}
	return r.pool.QueryRow(ctx, sql, args...)
}

func (r *OTATaskRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.Query(ctx, sql, args...)
	}
	return r.pool.Query(ctx, sql, args...)
}

func (r *OTATaskRepository) exec(ctx context.Context, sql string, args ...any) error {
	if conn := database.RLSConn(ctx); conn != nil {
		_, err := conn.Exec(ctx, sql, args...)
		return err
	}
	_, err := r.pool.Exec(ctx, sql, args...)
	return err
}

const otaTaskCols = `id, user_id, device_id, device_name, firmware_id, firmware_version, status, progress, error_msg, created_at, updated_at`

func scanOTATask(row pgx.Row) (*model.OTATask, error) {
	var t model.OTATask
	err := row.Scan(&t.ID, &t.UserID, &t.DeviceID, &t.DeviceName, &t.FirmwareID, &t.FirmwareVersion, &t.Status, &t.Progress, &t.ErrorMsg, &t.CreatedAt, &t.UpdatedAt)
	return &t, err
}

func (r *OTATaskRepository) Create(ctx context.Context, t *model.OTATask) (*model.OTATask, error) {
	row := r.queryRow(ctx, `
		INSERT INTO ota_tasks (user_id, device_id, device_name, firmware_id, firmware_version)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+otaTaskCols,
		t.UserID, t.DeviceID, t.DeviceName, t.FirmwareID, t.FirmwareVersion)
	return scanOTATask(row)
}

func (r *OTATaskRepository) GetByID(ctx context.Context, id string) (*model.OTATask, error) {
	row := r.queryRow(ctx, `SELECT `+otaTaskCols+` FROM ota_tasks WHERE id = $1`, id)
	return scanOTATask(row)
}

func (r *OTATaskRepository) List(ctx context.Context, offset, limit int, deviceID string) ([]*model.OTATask, int, error) {
	var total int
	if deviceID != "" {
		r.queryRow(ctx, `SELECT COUNT(*) FROM ota_tasks WHERE device_id = $1`, deviceID).Scan(&total)
	} else {
		r.queryRow(ctx, `SELECT COUNT(*) FROM ota_tasks`).Scan(&total)
	}

	var q string
	var args []any
	if deviceID != "" {
		q = `SELECT ` + otaTaskCols + ` FROM ota_tasks WHERE device_id = $3 ORDER BY created_at DESC OFFSET $1 LIMIT $2`
		args = []any{offset, limit, deviceID}
	} else {
		q = `SELECT ` + otaTaskCols + ` FROM ota_tasks ORDER BY created_at DESC OFFSET $1 LIMIT $2`
		args = []any{offset, limit}
	}

	rows, err := r.query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*model.OTATask
	for rows.Next() {
		var t model.OTATask
		if err := rows.Scan(&t.ID, &t.UserID, &t.DeviceID, &t.DeviceName, &t.FirmwareID, &t.FirmwareVersion, &t.Status, &t.Progress, &t.ErrorMsg, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &t)
	}
	return list, total, nil
}

// UpdateProgress 更新任务进度（MQTT hook 调用，绕过 RLS）
func (r *OTATaskRepository) UpdateProgress(ctx context.Context, id, status string, progress int, errorMsg string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ota_tasks SET status = $2, progress = $3, error_msg = $4 WHERE id = $1
	`, id, status, progress, errorMsg)
	return err
}

// Cancel 取消任务
func (r *OTATaskRepository) Cancel(ctx context.Context, id string) error {
	return r.exec(ctx, `UPDATE ota_tasks SET status = 'cancelled' WHERE id = $1 AND status IN ('pending', 'pushing')`, id)
}

// ListPendingByDeviceID 获取设备的 pending 任务（设备上线时检查，绕过 RLS）
func (r *OTATaskRepository) ListPendingByDeviceID(ctx context.Context, deviceID string) ([]*model.OTATask, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+otaTaskCols+` FROM ota_tasks
		WHERE device_id = $1 AND status = 'pending'
		ORDER BY created_at ASC LIMIT 1
	`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.OTATask
	for rows.Next() {
		var t model.OTATask
		if err := rows.Scan(&t.ID, &t.UserID, &t.DeviceID, &t.DeviceName, &t.FirmwareID, &t.FirmwareVersion, &t.Status, &t.Progress, &t.ErrorMsg, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, nil
}
