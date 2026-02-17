package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/config"
)

// RoleType 数据库角色类型
type RoleType string

const (
	RoleAdmin RoleType = "admin"
	RoleApp   RoleType = "app"
	RoleRead  RoleType = "read"
)

// DB 多连接池管理器
type DB struct {
	adminPool *pgxpool.Pool
	appPool   *pgxpool.Pool
	readPool  *pgxpool.Pool
}

// New 创建数据库连接池管理器
func New(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	adminPool, err := createPool(ctx, cfg, cfg.AdminUser, cfg.AdminPassword)
	if err != nil {
		return nil, fmt.Errorf("create admin pool: %w", err)
	}

	appPool, err := createPool(ctx, cfg, cfg.AppUser, cfg.AppPassword)
	if err != nil {
		adminPool.Close()
		return nil, fmt.Errorf("create app pool: %w", err)
	}

	readPool, err := createPool(ctx, cfg, cfg.ReadUser, cfg.ReadPassword)
	if err != nil {
		adminPool.Close()
		appPool.Close()
		return nil, fmt.Errorf("create read pool: %w", err)
	}

	return &DB{
		adminPool: adminPool,
		appPool:   appPool,
		readPool:  readPool,
	}, nil
}

func createPool(ctx context.Context, cfg config.DatabaseConfig, user, password string) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, cfg.Host, cfg.Port, cfg.Name,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = 10
	poolCfg.MinConns = 2

	return pgxpool.NewWithConfig(ctx, poolCfg)
}

// Pool 根据角色获取连接池
func (db *DB) Pool(role RoleType) *pgxpool.Pool {
	switch role {
	case RoleAdmin:
		return db.adminPool
	case RoleRead:
		return db.readPool
	default:
		return db.appPool
	}
}

// Admin 获取管理员连接池
func (db *DB) Admin() *pgxpool.Pool {
	return db.adminPool
}

// App 获取应用连接池
func (db *DB) App() *pgxpool.Pool {
	return db.appPool
}

// Read 获取只读连接池
func (db *DB) Read() *pgxpool.Pool {
	return db.readPool
}

// Close 关闭所有连接池
func (db *DB) Close() {
	db.adminPool.Close()
	db.appPool.Close()
	db.readPool.Close()
}

// WithRLS 在连接上设置 RLS 上下文变量
func WithRLS(ctx context.Context, pool *pgxpool.Pool, userID, userRole string) (context.Context, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return ctx, err
	}

	// 设置 RLS 会话变量 (is_local=false 表示会话级别)
	_, err = conn.Exec(ctx, "SELECT set_config('app.current_user_id', $1, false)", userID)
	if err != nil {
		conn.Release()
		return ctx, err
	}

	_, err = conn.Exec(ctx, "SELECT set_config('app.current_user_role', $1, false)", userRole)
	if err != nil {
		conn.Release()
		return ctx, err
	}

	// 将连接存入 context
	return context.WithValue(ctx, rlsConnKey{}, conn), nil
}

// RLSConn 从 context 获取 RLS 连接
func RLSConn(ctx context.Context) *pgxpool.Conn {
	if conn, ok := ctx.Value(rlsConnKey{}).(*pgxpool.Conn); ok {
		return conn
	}
	return nil
}

// ReleaseRLSConn 释放 RLS 连接
func ReleaseRLSConn(ctx context.Context) {
	if conn := RLSConn(ctx); conn != nil {
		conn.Release()
	}
}

type rlsConnKey struct{}
