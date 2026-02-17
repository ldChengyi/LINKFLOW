package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/model"
)

// AuditLogRepository 审计日志数据访问（直接使用 admin pool，不走 RLS）
type AuditLogRepository struct {
	pool *pgxpool.Pool
}

func NewAuditLogRepository(pool *pgxpool.Pool) *AuditLogRepository {
	return &AuditLogRepository{pool: pool}
}

// Create 插入一条审计日志
func (r *AuditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	var detailJSON []byte
	if log.Detail != nil {
		detailJSON, _ = json.Marshal(log.Detail)
	}

	if log.Category == "" {
		log.Category = model.AuditCategoryAPI
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_logs (user_id, category, action, resource, detail, ip, status_code, latency_ms, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, log.UserID, log.Category, log.Action, log.Resource, detailJSON, log.IP, log.StatusCode, log.LatencyMs, log.UserAgent, log.CreatedAt)
	return err
}

// List 分页查询审计日志（支持筛选）
func (r *AuditLogRepository) List(ctx context.Context, q *model.AuditLogQuery) ([]*model.AuditLog, int, error) {
	var (
		where []string
		args  []any
		idx   = 1
	)

	if q.UserID != "" {
		where = append(where, fmt.Sprintf("user_id = $%d", idx))
		args = append(args, q.UserID)
		idx++
	}
	if q.Category != "" {
		where = append(where, fmt.Sprintf("category = $%d", idx))
		args = append(args, q.Category)
		idx++
	}
	if q.Action != "" {
		where = append(where, fmt.Sprintf("action ILIKE $%d", idx))
		args = append(args, "%"+q.Action+"%")
		idx++
	}
	if q.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, q.StartTime); err == nil {
			where = append(where, fmt.Sprintf("created_at >= $%d", idx))
			args = append(args, t)
			idx++
		}
	}
	if q.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, q.EndTime); err == nil {
			where = append(where, fmt.Sprintf("created_at <= $%d", idx))
			args = append(args, t)
			idx++
		}
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	// 查总数
	var total int
	countSQL := "SELECT COUNT(*) FROM audit_logs " + whereClause
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 分页
	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	dataSQL := fmt.Sprintf(
		"SELECT id, user_id, category, action, resource, detail, ip, status_code, latency_ms, user_agent, created_at FROM audit_logs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		whereClause, idx, idx+1,
	)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*model.AuditLog
	for rows.Next() {
		a := &model.AuditLog{}
		var detailRaw []byte
		if err := rows.Scan(&a.ID, &a.UserID, &a.Category, &a.Action, &a.Resource, &detailRaw, &a.IP, &a.StatusCode, &a.LatencyMs, &a.UserAgent, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		if detailRaw != nil {
			json.Unmarshal(detailRaw, &a.Detail)
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil && err != pgx.ErrNoRows {
		return nil, 0, err
	}

	return list, total, nil
}
