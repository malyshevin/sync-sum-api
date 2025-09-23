package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/malyshevin/sync-sum-api/internal/config"
	"github.com/malyshevin/sync-sum-api/internal/httpapi"
	"github.com/malyshevin/sync-sum-api/internal/repository"
	"github.com/malyshevin/sync-sum-api/internal/service"
)

func TestCounterConcurrentIncrements(t *testing.T) {
	ctx := context.Background()

	t.Logf("start postgres")

	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("syncsum"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	host, err := pg.Host(ctx)
	if err != nil {
		t.Fatalf("host: %v", err)
	}
	port, err := pg.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("port: %v", err)
	}

	t.Logf("host: %s, port: %d", host, port.Int())

	cfg := config.Config{
		HTTP: config.HTTPConfig{Port: 0},
		Database: config.DatabaseConfig{
			Host:     host,
			Port:     port.Int(),
			User:     "postgres",
			Password: "postgres",
			Name:     "syncsum",
			SSLMode:  "disable",
		},
		Migrate: config.MigrateConfig{Dir: "../migrations", Version: 0},
	}

	t.Logf("cfg: %+v", cfg)

	// wait for postgres readiness
	{
		deadline := time.Now().Add(60 * time.Second)
		for {
			pool, err := repository.OpenPostgres(ctx, cfg.Database.DSN(), nil)
			if err == nil {
				pool.Close()
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("postgres not ready: %v", err)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}

	t.Logf("postgres ready")

	// apply migrations
	sourceURL := fmt.Sprintf("file://%s", cfg.Migrate.Dir)
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name, cfg.Database.SSLMode,
	)
	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		t.Fatalf("migrate init: %v", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}

	t.Logf("migrations applied")

	// start app in-memory
	db, err := repository.OpenPostgres(ctx, cfg.Database.DSN(), nil)
	if err != nil {
		t.Fatalf("db: %v", err)
	}
	defer db.Close()
	repo := repository.NewCounterStore(db, nil)
	svc := service.NewCounterService(repo)
	handler := httpapi.NewCounterHandler(svc, nil)
	r := chi.NewRouter()
	httpapi.RegisterRoutes(r, handler)
	srv := httptest.NewServer(r)
	defer srv.Close()

	t.Logf("app started")

	client := &http.Client{Timeout: 5 * time.Second}

	const c, n = 10, 100
	wg := sync.WaitGroup{}
	wg.Add(c)
	for j := 0; j < c; j++ {
		go func() {
			t.Logf("start j=%d", j)

			defer wg.Done()
			for i := 0; i < n; i++ {
				req, _ := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL+"/counter/increment", nil)
				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("request: %v", err)
					return
				}
				_ = resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Errorf("status %d", resp.StatusCode)
				}
			}
		}()
	}
	wg.Wait()

	// get value
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/counter", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	var payload struct {
		Value int64 `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload.Value != n*c {
		t.Fatalf("expected %d got %d", n*c, payload.Value)
	}
}
