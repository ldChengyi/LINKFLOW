//go:build integration

package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ldchengyi/linkflow/internal/cache"
	"github.com/ldchengyi/linkflow/internal/logger"
)

var (
	testPool *pgxpool.Pool
	testRdb  *cache.Redis
)

func TestMain(m *testing.M) {
	logger.Init(logger.Config{Level: "error"})
	ctx := context.Background()
	db := testCfg.Integration.Database
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		db.AdminUser, db.AdminPassword, db.Host, db.Port, db.Name)

	var err error
	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect pg: %v\n", err)
		os.Exit(1)
	}

	r := testCfg.Integration.Redis
	testRdb, err = cache.NewRedis(r.Addr, r.Password, r.DB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect redis: %v\n", err)
		os.Exit(1)
	}

	setupRouter()

	code := m.Run()

	testPool.Exec(ctx, "DELETE FROM scheduled_tasks WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@linkflow.dev')")
	testPool.Exec(ctx, "DELETE FROM alert_logs WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@linkflow.dev')")
	testPool.Exec(ctx, "DELETE FROM alert_rules WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@linkflow.dev')")
	testPool.Exec(ctx, "DELETE FROM device_data WHERE device_id IN (SELECT id FROM devices WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@linkflow.dev'))")
	testPool.Exec(ctx, "DELETE FROM devices WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@linkflow.dev')")
	testPool.Exec(ctx, "DELETE FROM thing_models WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@linkflow.dev')")
	testPool.Exec(ctx, "DELETE FROM users WHERE email LIKE '%@linkflow.dev'")
	testRdb.Client().FlushDB(ctx)
	testPool.Close()
	testRdb.Close()

	os.Exit(code)
}
