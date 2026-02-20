package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
)

type AlertLogRepository struct {
	pool *pgxpool.Pool
}

func NewAlertLogRepository(pool *pgxpool.Pool) *AlertLogRepository {
	return &AlertLogRepository{pool: pool}
}

func (r *AlertLogRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.QueryRow(ctx, sql, args...)
	}
	return r.pool.QueryRow(ctx, sql, args...)
}

func (r *AlertLogRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.Query(ctx, sql, args...)
	}
	return r.pool.Query(ctx, sql, args...)
}

func (r *AlertLogRepository) exec(ctx context.Context, sql string, args ...any) error {
	if conn := database.RLSConn(ctx); conn != nil {
		_, err := conn.Exec(ctx, sql, args...)
		return err
	}
	_, err := r.pool.Exec(ctx, sql, args...)
	return err
}

// Create 写入告警日志（MQTT 触发，绕过 RLS）
func (r *AlertLogRepository) Create(ctx context.Context, log *model.AlertLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO alert_logs (rule_id, user_id, device_id, device_name, property_id, property_name, operator, threshold, actual_value, severity, rule_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, log.RuleID, log.UserID, log.DeviceID, log.DeviceName, log.PropertyID, log.PropertyName, log.Operator, log.Threshold, log.ActualValue, log.Severity, log.RuleName)
	return err
}

// List 查询告警日志（分页，可选按设备过滤）
func (r *AlertLogRepository) List(ctx context.Context, offset, limit int, deviceID string) ([]*model.AlertLog, int, error) {
	var total int
	if deviceID != "" {
		err := r.queryRow(ctx, `SELECT COUNT(*) FROM alert_logs WHERE device_id = $1`, deviceID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.queryRow(ctx, `SELECT COUNT(*) FROM alert_logs`).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	var query string
	var args []any
	if deviceID != "" {
		query = `
			SELECT id, COALESCE(rule_id::text, ''), user_id, device_id, device_name,
			       property_id, property_name, operator, threshold, actual_value,
			       severity, rule_name, acknowledged, acknowledged_at, created_at
			FROM alert_logs
			WHERE device_id = $3
			ORDER BY created_at DESC
			OFFSET $1 LIMIT $2`
		args = []any{offset, limit, deviceID}
	} else {
		query = `
			SELECT id, COALESCE(rule_id::text, ''), user_id, device_id, device_name,
			       property_id, property_name, operator, threshold, actual_value,
			       severity, rule_name, acknowledged, acknowledged_at, created_at
			FROM alert_logs
			ORDER BY created_at DESC
			OFFSET $1 LIMIT $2`
		args = []any{offset, limit}
	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*model.AlertLog
	for rows.Next() {
		var l model.AlertLog
		err := rows.Scan(
			&l.ID, &l.RuleID, &l.UserID, &l.DeviceID, &l.DeviceName,
			&l.PropertyID, &l.PropertyName, &l.Operator, &l.Threshold, &l.ActualValue,
			&l.Severity, &l.RuleName, &l.Acknowledged, &l.AcknowledgedAt, &l.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, &l)
	}
	return logs, total, nil
}

// Acknowledge 确认告警（RLS 连接，只能确认自己的）
func (r *AlertLogRepository) Acknowledge(ctx context.Context, id int64) error {
	return r.exec(ctx, `
		UPDATE alert_logs
		SET acknowledged = true, acknowledged_at = NOW()
		WHERE id = $1 AND acknowledged = false
	`, id)
}

// CountUnacknowledged 统计未确认告警数（RLS 连接）
func (r *AlertLogRepository) CountUnacknowledged(ctx context.Context) (int, error) {
	var count int
	err := r.queryRow(ctx, `SELECT COUNT(*) FROM alert_logs WHERE acknowledged = false`).Scan(&count)
	return count, err
}

// CountToday 统计今日告警数（RLS 连接）
func (r *AlertLogRepository) CountToday(ctx context.Context) (int, error) {
	var count int
	err := r.queryRow(ctx, `SELECT COUNT(*) FROM alert_logs WHERE created_at >= CURRENT_DATE`).Scan(&count)
	return count, err
}
