package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/logger"
)

// baselineFiles are migrations already applied via init-db.sql.template.
// New migrations added after this list will be auto-applied on startup.
var baselineFiles = []string{
	"001_init_roles.sql",
	"002_users_table.sql",
	"003_devices_table_rls.sql",
	"004_device_data_hypertable.sql",
	"005_modules.sql",
	"006_audit_logs.sql",
	"007_alert_system.sql",
	"008_scheduled_tasks.sql",
	"009_alert_enhancements.sql",
	"010_ota_system.sql",
	"011_platform_settings.sql",
	"011_scheduled_task_logs.sql",
}

// Migrate runs pending SQL migrations. Existing databases get baseline
// files marked as applied without re-executing them.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	// Create tracking table
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// Seed baseline if this is the first run on an existing DB
	if err := seedBaseline(ctx, pool); err != nil {
		return err
	}

	// Load applied versions
	applied, err := loadApplied(ctx, pool)
	if err != nil {
		return err
	}

	// Read and sort migration files
	files, err := listMigrationFiles(migrationsDir)
	if err != nil {
		return err
	}

	// Apply pending
	for _, f := range files {
		if applied[f] {
			continue
		}
		if err := applyMigration(ctx, pool, migrationsDir, f); err != nil {
			return err
		}
		logger.Log.Infof("Applied migration: %s", f)
	}

	return nil
}

func seedBaseline(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		return fmt.Errorf("check schema_migrations: %w", err)
	}
	if count > 0 {
		return nil // already seeded
	}

	// Check if DB was already initialized (users table exists)
	var exists bool
	err = pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name='users')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("check existing tables: %w", err)
	}
	if !exists {
		return nil // fresh DB, no baseline needed
	}

	// Mark all baseline files as applied
	for _, f := range baselineFiles {
		_, err := pool.Exec(ctx,
			"INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", f)
		if err != nil {
			return fmt.Errorf("seed baseline %s: %w", f, err)
		}
	}
	logger.Log.Info("Seeded baseline migrations for existing database")
	return nil
}

func loadApplied(ctx context.Context, pool *pgxpool.Pool) (map[string]bool, error) {
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, nil
}

func listMigrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, dir, file string) error {
	content, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return fmt.Errorf("read migration %s: %w", file, err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx for %s: %w", file, err)
	}

	if _, err := tx.Exec(ctx, string(content)); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("execute migration %s: %w", file, err)
	}

	if _, err := tx.Exec(ctx,
		"INSERT INTO schema_migrations (version) VALUES ($1)", file); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("record migration %s: %w", file, err)
	}

	return tx.Commit(ctx)
}
