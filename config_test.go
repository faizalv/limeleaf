package limeleaf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWritePostgresConf(t *testing.T) {
	dir := t.TempDir()

	err := writePostgresConf(dir, 15432, map[string]string{
		"shared_buffers": "'256MB'",
	})
	if err != nil {
		t.Fatalf("writePostgresConf: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "postgresql.conf"))
	if err != nil {
		t.Fatalf("reading conf: %v", err)
	}
	content := string(data)

	checks := []struct {
		name, want string
	}{
		{"port", "port = 15432"},
		{"listen_addresses", "listen_addresses = '127.0.0.1'"},
		{"unix_socket_directories", "unix_socket_directories = ''"},
		{"shared_buffers override", "shared_buffers = '256MB'"},
		{"max_connections", "max_connections = 20"},
	}

	for _, c := range checks {
		if !strings.Contains(content, c.want) {
			t.Errorf("%s: conf does not contain %q", c.name, c.want)
		}
	}
}

func TestWritePostgresConfDefaults(t *testing.T) {
	dir := t.TempDir()

	err := writePostgresConf(dir, 5433, nil)
	if err != nil {
		t.Fatalf("writePostgresConf: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "postgresql.conf"))
	if err != nil {
		t.Fatalf("reading conf: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "shared_buffers = '128MB'") {
		t.Error("default shared_buffers not set")
	}
}
