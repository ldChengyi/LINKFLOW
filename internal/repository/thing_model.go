package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
)

type ThingModelRepository struct {
	pool *pgxpool.Pool
}

func NewThingModelRepository(pool *pgxpool.Pool) *ThingModelRepository {
	return &ThingModelRepository{pool: pool}
}

// queryRow 执行查询（优先使用 RLS 连接）
func (r *ThingModelRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.QueryRow(ctx, sql, args...)
	}
	return r.pool.QueryRow(ctx, sql, args...)
}

// query 执行查询（优先使用 RLS 连接）
func (r *ThingModelRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.Query(ctx, sql, args...)
	}
	return r.pool.Query(ctx, sql, args...)
}

// exec 执行命令（优先使用 RLS 连接）
func (r *ThingModelRepository) exec(ctx context.Context, sql string, args ...any) error {
	if conn := database.RLSConn(ctx); conn != nil {
		_, err := conn.Exec(ctx, sql, args...)
		return err
	}
	_, err := r.pool.Exec(ctx, sql, args...)
	return err
}

// Create 创建物模型
func (r *ThingModelRepository) Create(ctx context.Context, userID string, req *model.CreateThingModelRequest) (*model.ThingModel, error) {
	properties, _ := json.Marshal(req.Properties)
	events, _ := json.Marshal(req.Events)
	services, _ := json.Marshal(req.Services)
	modules, _ := json.Marshal(req.Modules)

	if req.Properties == nil {
		properties = []byte("[]")
	}
	if req.Events == nil {
		events = []byte("[]")
	}
	if req.Services == nil {
		services = []byte("[]")
	}
	if req.Modules == nil {
		modules = []byte("[]")
	}

	var db model.ThingModelDB
	err := r.queryRow(ctx, `
		INSERT INTO thing_models (user_id, name, description, properties, events, services, modules)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, name, description, properties, events, services, modules, created_at, updated_at
	`, userID, req.Name, nullString(req.Description), properties, events, services, modules).Scan(
		&db.ID, &db.UserID, &db.Name, &db.Description,
		&db.Properties, &db.Events, &db.Services, &db.Modules,
		&db.CreatedAt, &db.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return db.ToModel()
}

// GetByID 根据ID获取物模型
func (r *ThingModelRepository) GetByID(ctx context.Context, id string) (*model.ThingModel, error) {
	var db model.ThingModelDB
	err := r.queryRow(ctx, `
		SELECT id, user_id, name, description, properties, events, services, modules, created_at, updated_at
		FROM thing_models WHERE id = $1
	`, id).Scan(
		&db.ID, &db.UserID, &db.Name, &db.Description,
		&db.Properties, &db.Events, &db.Services, &db.Modules,
		&db.CreatedAt, &db.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return db.ToModel()
}

// List 获取用户的物模型列表
func (r *ThingModelRepository) List(ctx context.Context, offset, limit int) ([]*model.ThingModel, int, error) {
	// 获取总数
	var total int
	err := r.queryRow(ctx, `SELECT COUNT(*) FROM thing_models`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	rows, err := r.query(ctx, `
		SELECT id, user_id, name, description, properties, events, services, modules, created_at, updated_at
		FROM thing_models
		ORDER BY created_at DESC
		OFFSET $1 LIMIT $2
	`, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var models []*model.ThingModel
	for rows.Next() {
		var db model.ThingModelDB
		err := rows.Scan(
			&db.ID, &db.UserID, &db.Name, &db.Description,
			&db.Properties, &db.Events, &db.Services, &db.Modules,
			&db.CreatedAt, &db.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		m, err := db.ToModel()
		if err != nil {
			return nil, 0, err
		}
		models = append(models, m)
	}
	return models, total, nil
}

// Update 更新物模型
func (r *ThingModelRepository) Update(ctx context.Context, id string, req *model.UpdateThingModelRequest) (*model.ThingModel, error) {
	properties, _ := json.Marshal(req.Properties)
	events, _ := json.Marshal(req.Events)
	services, _ := json.Marshal(req.Services)
	modules, _ := json.Marshal(req.Modules)

	if req.Properties == nil {
		properties = []byte("[]")
	}
	if req.Events == nil {
		events = []byte("[]")
	}
	if req.Services == nil {
		services = []byte("[]")
	}
	if req.Modules == nil {
		modules = []byte("[]")
	}

	var db model.ThingModelDB
	err := r.queryRow(ctx, `
		UPDATE thing_models
		SET name = $2, description = $3, properties = $4, events = $5, services = $6, modules = $7, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, name, description, properties, events, services, modules, created_at, updated_at
	`, id, req.Name, nullString(req.Description), properties, events, services, modules).Scan(
		&db.ID, &db.UserID, &db.Name, &db.Description,
		&db.Properties, &db.Events, &db.Services, &db.Modules,
		&db.CreatedAt, &db.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return db.ToModel()
}

// Delete 删除物模型
func (r *ThingModelRepository) Delete(ctx context.Context, id string) error {
	return r.exec(ctx, `DELETE FROM thing_models WHERE id = $1`, id)
}

// Count 获取物模型总数（受 RLS 约束）
func (r *ThingModelRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.queryRow(ctx, `SELECT COUNT(*) FROM thing_models`).Scan(&count)
	return count, err
}

// nullString 空字符串转nil
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
