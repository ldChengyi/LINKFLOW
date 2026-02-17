package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/model"
)

type ModuleRepository struct {
	pool *pgxpool.Pool
}

func NewModuleRepository(pool *pgxpool.Pool) *ModuleRepository {
	return &ModuleRepository{pool: pool}
}

// List 获取所有模块
func (r *ModuleRepository) List(ctx context.Context) ([]*model.Module, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, description, config_schema, created_at, updated_at
		FROM modules ORDER BY created_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var modules []*model.Module
	for rows.Next() {
		var db model.ModuleDB
		err := rows.Scan(&db.ID, &db.Name, &db.Description, &db.ConfigSchema, &db.CreatedAt, &db.UpdatedAt)
		if err != nil {
			return nil, err
		}
		m := &model.Module{
			ID:        db.ID,
			Name:      db.Name,
			CreatedAt: db.CreatedAt,
			UpdatedAt: db.UpdatedAt,
		}
		if db.Description != nil {
			m.Description = *db.Description
		}
		if db.ConfigSchema != nil {
			json.Unmarshal(db.ConfigSchema, &m.ConfigSchema)
		}
		modules = append(modules, m)
	}
	return modules, nil
}

// GetByID 根据ID获取模块
func (r *ModuleRepository) GetByID(ctx context.Context, id string) (*model.Module, error) {
	var db model.ModuleDB
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, config_schema, created_at, updated_at
		FROM modules WHERE id = $1
	`, id).Scan(&db.ID, &db.Name, &db.Description, &db.ConfigSchema, &db.CreatedAt, &db.UpdatedAt)
	if err != nil {
		return nil, err
	}
	m := &model.Module{
		ID:        db.ID,
		Name:      db.Name,
		CreatedAt: db.CreatedAt,
		UpdatedAt: db.UpdatedAt,
	}
	if db.Description != nil {
		m.Description = *db.Description
	}
	if db.ConfigSchema != nil {
		json.Unmarshal(db.ConfigSchema, &m.ConfigSchema)
	}
	return m, nil
}
