package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SettingsRepository 平台设置数据访问（使用 admin pool，无 RLS）
type SettingsRepository struct {
	pool *pgxpool.Pool
}

// NewSettingsRepository 创建设置 Repository
func NewSettingsRepository(pool *pgxpool.Pool) *SettingsRepository {
	return &SettingsRepository{pool: pool}
}

// GetAll 获取所有平台设置，返回 key→value 映射
func (r *SettingsRepository) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT key, value FROM platform_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		result[k] = v
	}
	return result, rows.Err()
}

// SetMany 批量更新设置（UPSERT）
func (r *SettingsRepository) SetMany(ctx context.Context, settings map[string]string) error {
	for k, v := range settings {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO platform_settings (key, value, updated_at)
			 VALUES ($1, $2, NOW())
			 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()`,
			k, v,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
