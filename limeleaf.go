package limeleaf

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

// Config controls how the embedded Postgres instance is set up and run.
type Config struct {
	// Required. Directory where Postgres stores its data.
	DataDir string

	// Port to listen on. Default: 0 (random available port).
	Port int

	// Database name to create on first init. Default: "limeleaf".
	Database string

	// Superuser name. Default: "limeleaf".
	Username string

	// Additional Postgres configuration parameters (e.g., "shared_buffers": "'256MB'").
	Settings map[string]string

	// Where to cache downloaded Postgres binaries. Default: ~/.limeleaf/cache/
	CacheDir string

	// Logger for lifecycle events. Default: discard.
	Logger *log.Logger
}

// Postgres represents a running Postgres instance managed by limeleaf.
type Postgres struct {
	port     int
	dataDir  string
	binDir   string
	database string
	username string
	logger   *log.Logger
}

// ConnectionString returns a DSN usable with database/sql, pgx, or lib/pq.
func (p *Postgres) ConnectionString() string {
	return fmt.Sprintf("postgresql://%s@127.0.0.1:%d/%s?sslmode=disable",
		p.username, p.port, p.database)
}

// Port returns the port Postgres is listening on.
func (p *Postgres) Port() int { return p.port }

// DataDir returns the path to the Postgres data directory.
func (p *Postgres) DataDir() string { return p.dataDir }

// Stop shuts down the Postgres instance.
func (p *Postgres) Stop() error {
	return stopPostgres(p.binDir, p.dataDir, p.logger)
}

// Start boots a Postgres instance. On first call with a given DataDir it runs
// initdb, creates the database, and enables pgvector. On subsequent calls it
// starts the existing cluster. It blocks until Postgres is accepting connections.
func Start(ctx context.Context, cfg Config) (*Postgres, error) {
	if cfg.DataDir == "" {
		return nil, fmt.Errorf("limeleaf: DataDir is required")
	}
	if cfg.Database == "" {
		cfg.Database = defaultDatabase
	}
	if cfg.Username == "" {
		cfg.Username = defaultUsername
	}
	if cfg.Logger == nil {
		cfg.Logger = log.New(io.Discard, "", 0)
	}

	binDir, err := EnsureBinary(ctx, cfg.CacheDir)
	if err != nil {
		return nil, fmt.Errorf("limeleaf: %w", err)
	}

	firstRun := isFirstRun(cfg.DataDir)

	if firstRun {
		if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
			return nil, fmt.Errorf("limeleaf: creating data dir: %w", err)
		}
		if err := initDB(ctx, binDir, cfg.DataDir, cfg.Username, cfg.Logger); err != nil {
			return nil, fmt.Errorf("limeleaf: %w", err)
		}
	} else {
		if err := cleanStalePid(cfg.DataDir); err != nil {
			return nil, fmt.Errorf("limeleaf: %w", err)
		}
	}

	port := cfg.Port
	if port == 0 {
		port, err = findFreePort()
		if err != nil {
			return nil, fmt.Errorf("limeleaf: finding free port: %w", err)
		}
	}

	if err := writePostgresConf(cfg.DataDir, port, cfg.Settings); err != nil {
		return nil, fmt.Errorf("limeleaf: writing config: %w", err)
	}

	if err := startPostgres(ctx, binDir, cfg.DataDir, cfg.Logger); err != nil {
		return nil, fmt.Errorf("limeleaf: %w", err)
	}

	if firstRun {
		if cfg.Database != "postgres" {
			sql := fmt.Sprintf("CREATE DATABASE %q", cfg.Database)
			if err := runSetupSQL(ctx, binDir, port, cfg.Username, "postgres", sql); err != nil {
				stopPostgres(binDir, cfg.DataDir, cfg.Logger)
				return nil, fmt.Errorf("limeleaf: creating database: %w", err)
			}
		}
		if err := runSetupSQL(ctx, binDir, port, cfg.Username, cfg.Database,
			"CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
			stopPostgres(binDir, cfg.DataDir, cfg.Logger)
			return nil, fmt.Errorf("limeleaf: creating pgvector extension: %w", err)
		}
		cfg.Logger.Printf("first-run setup complete: database=%s, extension=vector", cfg.Database)
	}

	return &Postgres{
		port:     port,
		dataDir:  cfg.DataDir,
		binDir:   binDir,
		database: cfg.Database,
		username: cfg.Username,
		logger:   cfg.Logger,
	}, nil
}

func findFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port, nil
}
