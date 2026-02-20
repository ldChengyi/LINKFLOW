package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
)

type AlertRuleRepository struct {
	pool *pgxpool.Pool
}

func NewAlertRuleRepository(pool *pgxpool.Pool) *AlertRuleRepository {
	return &AlertRuleRepository{pool: pool}
}

func (r *AlertRuleRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.QueryRow(ctx, sql, args...)
	}
	return r.pool.QueryRow(ctx, sql, args...)
}

func (r *AlertRuleRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.Query(ctx, sql, args...)
	}
	return r.pool.Query(ctx, sql, args...)
}

func (r *AlertRuleRepository) exec(ctx context.Context, sql string, args ...any) error {
	if conn := database.RLSConn(ctx); conn != nil {
		_, err := conn.Exec(ctx, sql, args...)
		return err
	}
	_, err := r.pool.Exec(ctx, sql, args...)
	return err
}

func scanRule(row pgx.Row) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := row.Scan(
		&rule.ID, &rule.UserID, &rule.Name, &rule.DeviceID, &rule.DeviceName,
		&rule.PropertyID, &rule.Operator, &rule.Threshold, &rule.Severity, &rule.Enabled,
		&rule.CooldownMinutes, &rule.LastTriggeredAt,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func scanRuleFromRows(rows pgx.Rows) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := rows.Scan(
		&rule.ID, &rule.UserID, &rule.Name, &rule.DeviceID, &rule.DeviceName,
		&rule.PropertyID, &rule.Operator, &rule.Threshold, &rule.Severity, &rule.Enabled,
		&rule.CooldownMinutes, &rule.LastTriggeredAt,
		&rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// Create 创建告警规则
func (r *AlertRuleRepository) Create(ctx context.Context, userID string, req *model.CreateAlertRuleRequest) (*model.AlertRule, error) {
	cooldown := 0
	if req.CooldownMinutes != nil {
		cooldown = *req.CooldownMinutes
	}
	row := r.queryRow(ctx, `
		INSERT INTO alert_rules (user_id, name, device_id, property_id, operator, threshold, severity, enabled, cooldown_minutes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, name, device_id, '', property_id, operator, threshold, severity, enabled,
		          cooldown_minutes, last_triggered_at, created_at, updated_at
	`, userID, req.Name, req.DeviceID, req.PropertyID, req.Operator, req.Threshold, req.Severity, req.Enabled, cooldown)
	return scanRule(row)
}

// GetByID 获取告警规则详情
func (r *AlertRuleRepository) GetByID(ctx context.Context, id string) (*model.AlertRule, error) {
	row := r.queryRow(ctx, `
		SELECT r.id, r.user_id, r.name, r.device_id, COALESCE(d.name, '') AS device_name,
		       r.property_id, r.operator, r.threshold, r.severity, r.enabled,
		       r.cooldown_minutes, r.last_triggered_at, r.created_at, r.updated_at
		FROM alert_rules r
		LEFT JOIN devices d ON r.device_id = d.id
		WHERE r.id = $1
	`, id)
	return scanRule(row)
}

// List 获取告警规则列表（可选按设备过滤）
func (r *AlertRuleRepository) List(ctx context.Context, offset, limit int, deviceID string) ([]*model.AlertRule, int, error) {
	var total int
	if deviceID != "" {
		err := r.queryRow(ctx, `SELECT COUNT(*) FROM alert_rules WHERE device_id = $1`, deviceID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.queryRow(ctx, `SELECT COUNT(*) FROM alert_rules`).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	var query string
	var args []any
	if deviceID != "" {
		query = `
			SELECT r.id, r.user_id, r.name, r.device_id, COALESCE(d.name, '') AS device_name,
			       r.property_id, r.operator, r.threshold, r.severity, r.enabled,
			       r.cooldown_minutes, r.last_triggered_at, r.created_at, r.updated_at
			FROM alert_rules r
			LEFT JOIN devices d ON r.device_id = d.id
			WHERE r.device_id = $3
			ORDER BY r.created_at DESC
			OFFSET $1 LIMIT $2`
		args = []any{offset, limit, deviceID}
	} else {
		query = `
			SELECT r.id, r.user_id, r.name, r.device_id, COALESCE(d.name, '') AS device_name,
			       r.property_id, r.operator, r.threshold, r.severity, r.enabled,
			       r.cooldown_minutes, r.last_triggered_at, r.created_at, r.updated_at
			FROM alert_rules r
			LEFT JOIN devices d ON r.device_id = d.id
			ORDER BY r.created_at DESC
			OFFSET $1 LIMIT $2`
		args = []any{offset, limit}
	}

	rows, err := r.query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rules []*model.AlertRule
	for rows.Next() {
		rule, err := scanRuleFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		rules = append(rules, rule)
	}
	return rules, total, nil
}

// Update 更新告警规则
func (r *AlertRuleRepository) Update(ctx context.Context, id string, req *model.UpdateAlertRuleRequest) (*model.AlertRule, error) {
	cooldown := 0
	if req.CooldownMinutes != nil {
		cooldown = *req.CooldownMinutes
	}
	row := r.queryRow(ctx, `
		UPDATE alert_rules
		SET name = $2, device_id = $3, property_id = $4, operator = $5, threshold = $6,
		    severity = $7, enabled = $8, cooldown_minutes = $9, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, name, device_id, '', property_id, operator, threshold, severity, enabled,
		          cooldown_minutes, last_triggered_at, created_at, updated_at
	`, id, req.Name, req.DeviceID, req.PropertyID, req.Operator, req.Threshold, req.Severity, req.Enabled, cooldown)
	return scanRule(row)
}

// Delete 删除告警规则
func (r *AlertRuleRepository) Delete(ctx context.Context, id string) error {
	return r.exec(ctx, `DELETE FROM alert_rules WHERE id = $1`, id)
}

// ListEnabledByDeviceID 获取设备的所有启用规则（MQTT 告警评估用，绕过 RLS）
func (r *AlertRuleRepository) ListEnabledByDeviceID(ctx context.Context, deviceID string) ([]*model.AlertRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.user_id, r.name, r.device_id, COALESCE(d.name, '') AS device_name,
		       r.property_id, r.operator, r.threshold, r.severity, r.enabled,
		       r.cooldown_minutes, r.last_triggered_at, r.created_at, r.updated_at
		FROM alert_rules r
		LEFT JOIN devices d ON r.device_id = d.id
		WHERE r.device_id = $1 AND r.enabled = true
	`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*model.AlertRule
	for rows.Next() {
		rule, err := scanRuleFromRows(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// UpdateLastTriggeredAt 更新规则最后触发时间（绕过 RLS）
func (r *AlertRuleRepository) UpdateLastTriggeredAt(ctx context.Context, ruleID string, t time.Time) error {
	_, err := r.pool.Exec(ctx, `UPDATE alert_rules SET last_triggered_at = $2 WHERE id = $1`, ruleID, t)
	return err
}
