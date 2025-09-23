package main

import (
	"context"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/malyshevin/sync-sum-api/internal/config"
	"github.com/malyshevin/sync-sum-api/internal/repository"
)

func main() {
	_ = godotenv.Overload(".env")
	_ = godotenv.Overload(".env.local")

	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	pool, err := repository.OpenPostgres(ctx, cfg.Database.DSN(), nil)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	migrationsDir := cfg.Migrate.Dir
	sourceURL := fmt.Sprintf("file://%s", migrationsDir)
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}
	defer m.Close()

	targetVersion := cfg.Migrate.Version

	if targetVersion == 0 {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up: %v", err)
		}
		fmt.Println("migrations applied: up to latest")
		return
	}

	if err := m.Migrate(uint(targetVersion)); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrate to %d: %v", targetVersion, err)
	}
	fmt.Printf("migrated to version: %d\n", targetVersion)
}
