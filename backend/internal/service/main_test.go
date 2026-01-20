package service_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testPool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}

	postgresURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	testPool, err = pgxpool.New(ctx, postgresURL)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}

	if err := applyMigrations(ctx, testPool); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	exitCode := m.Run()

	testPool.Close()

	if err := container.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}

	os.Exit(exitCode)
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		"000001_init_tables.up.sql",
		"000002_cashbook.up.sql",
	}

	for _, migrationFile := range migrations {
		migrationPath := filepath.Join("..", "..", "migrations", migrationFile)

		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			return err
		}

		_, err = pool.Exec(ctx, string(migration))
		if err != nil {
			return err
		}
	}

	return nil
}
