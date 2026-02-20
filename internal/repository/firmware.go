package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/database"
	"github.com/ldchengyi/linkflow/internal/model"
)

type FirmwareRepository struct {
	pool *pgxpool.Pool
}

func NewFirmwareRepository(pool *pgxpool.Pool) *FirmwareRepository {
	return &FirmwareRepository{pool: pool}
}

func (r *FirmwareRepository) queryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.QueryRow(ctx, sql, args...)
	}
	return r.pool.QueryRow(ctx, sql, args...)
}

func (r *FirmwareRepository) query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if conn := database.RLSConn(ctx); conn != nil {
		return conn.Query(ctx, sql, args...)
	}
	return r.pool.Query(ctx, sql, args...)
}

func (r *FirmwareRepository) exec(ctx context.Context, sql string, args ...any) error {
	if conn := database.RLSConn(ctx); conn != nil {
		_, err := conn.Exec(ctx, sql, args...)
		return err
	}
	_, err := r.pool.Exec(ctx, sql, args...)
	return err
}

func scanFirmware(row pgx.Row) (*model.Firmware, error) {
	var f model.Firmware
	err := row.Scan(&f.ID, &f.UserID, &f.Name, &f.Version, &f.FilePath, &f.FileSize, &f.Checksum, &f.Description, &f.CreatedAt)
	return &f, err
}

func (r *FirmwareRepository) Create(ctx context.Context, f *model.Firmware) (*model.Firmware, error) {
	row := r.queryRow(ctx, `
		INSERT INTO firmwares (user_id, name, version, file_path, file_size, checksum, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, name, version, file_path, file_size, checksum, description, created_at
	`, f.UserID, f.Name, f.Version, f.FilePath, f.FileSize, f.Checksum, f.Description)
	return scanFirmware(row)
}

func (r *FirmwareRepository) GetByID(ctx context.Context, id string) (*model.Firmware, error) {
	row := r.queryRow(ctx, `
		SELECT id, user_id, name, version, file_path, file_size, checksum, description, created_at
		FROM firmwares WHERE id = $1
	`, id)
	return scanFirmware(row)
}

func (r *FirmwareRepository) List(ctx context.Context, offset, limit int) ([]*model.Firmware, int, error) {
	var total int
	if err := r.queryRow(ctx, `SELECT COUNT(*) FROM firmwares`).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.query(ctx, `
		SELECT id, user_id, name, version, file_path, file_size, checksum, description, created_at
		FROM firmwares ORDER BY created_at DESC OFFSET $1 LIMIT $2
	`, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []*model.Firmware
	for rows.Next() {
		var f model.Firmware
		if err := rows.Scan(&f.ID, &f.UserID, &f.Name, &f.Version, &f.FilePath, &f.FileSize, &f.Checksum, &f.Description, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, &f)
	}
	return list, total, nil
}

func (r *FirmwareRepository) Delete(ctx context.Context, id string) error {
	return r.exec(ctx, `DELETE FROM firmwares WHERE id = $1`, id)
}
