//go:build integration

package limeleaf

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestIntegrationStartStop(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pg, err := Start(ctx, Config{
		DataDir: dataDir,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer pg.Stop()

	if pg.Port() < 1024 {
		t.Errorf("unexpected port: %d", pg.Port())
	}

	db, err := sql.Open("postgres", pg.ConnectionString())
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestIntegrationPgvector(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pg, err := Start(ctx, Config{
		DataDir: dataDir,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer pg.Stop()

	db, err := sql.Open("postgres", pg.ConnectionString())
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, `CREATE TABLE items (
		id SERIAL PRIMARY KEY,
		embedding vector(3)
	)`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}

	_, err = db.ExecContext(ctx, `INSERT INTO items (embedding) VALUES ('[1,2,3]'), ('[4,5,6]'), ('[7,8,9]')`)
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}

	var id int
	err = db.QueryRowContext(ctx, `SELECT id FROM items ORDER BY embedding <-> '[3,3,3]' LIMIT 1`).Scan(&id)
	if err != nil {
		t.Fatalf("vector search query: %v", err)
	}
	if id != 1 {
		t.Errorf("nearest neighbor id = %d, want 1", id)
	}
}

func TestIntegrationRestart(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pg, err := Start(ctx, Config{
		DataDir: dataDir,
	})
	if err != nil {
		t.Fatalf("first Start: %v", err)
	}

	db, err := sql.Open("postgres", pg.ConnectionString())
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}

	_, err = db.ExecContext(ctx, `CREATE TABLE persist_test (val TEXT)`)
	if err != nil {
		t.Fatalf("CREATE TABLE: %v", err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO persist_test (val) VALUES ('survived')`)
	if err != nil {
		t.Fatalf("INSERT: %v", err)
	}
	db.Close()

	if err := pg.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	pg2, err := Start(ctx, Config{
		DataDir: dataDir,
	})
	if err != nil {
		t.Fatalf("second Start: %v", err)
	}
	defer pg2.Stop()

	db2, err := sql.Open("postgres", pg2.ConnectionString())
	if err != nil {
		t.Fatalf("sql.Open after restart: %v", err)
	}
	defer db2.Close()

	var val string
	err = db2.QueryRowContext(ctx, `SELECT val FROM persist_test`).Scan(&val)
	if err != nil {
		t.Fatalf("SELECT after restart: %v", err)
	}
	if val != "survived" {
		t.Errorf("val = %q, want %q", val, "survived")
	}
}

func TestIntegrationCustomConfig(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pg, err := Start(ctx, Config{
		DataDir:  dataDir,
		Database: "mydb",
		Username: "myuser",
		Settings: map[string]string{
			"max_connections": "5",
		},
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer pg.Stop()

	connStr := pg.ConnectionString()
	if want := fmt.Sprintf("postgresql://myuser@127.0.0.1:%d/mydb?sslmode=disable", pg.Port()); connStr != want {
		t.Errorf("ConnectionString = %q, want %q", connStr, want)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	var maxConn string
	err = db.QueryRowContext(ctx, `SHOW max_connections`).Scan(&maxConn)
	if err != nil {
		t.Fatalf("SHOW max_connections: %v", err)
	}
	if maxConn != "5" {
		t.Errorf("max_connections = %q, want %q", maxConn, "5")
	}
}

func TestIntegrationDataDir(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "data")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pg, err := Start(ctx, Config{
		DataDir: dataDir,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer pg.Stop()

	if pg.DataDir() != dataDir {
		t.Errorf("DataDir() = %q, want %q", pg.DataDir(), dataDir)
	}

	if _, err := os.Stat(filepath.Join(dataDir, "PG_VERSION")); err != nil {
		t.Errorf("PG_VERSION not found in data dir: %v", err)
	}
}
