package limeleaf

import (
	"context"
	"testing"
)

func TestFindFreePort(t *testing.T) {
	port, err := findFreePort()
	if err != nil {
		t.Fatalf("findFreePort: %v", err)
	}
	if port < 1024 || port > 65535 {
		t.Errorf("port %d outside expected range", port)
	}
}

func TestFindFreePortUnique(t *testing.T) {
	seen := make(map[int]bool)
	for range 10 {
		port, err := findFreePort()
		if err != nil {
			t.Fatalf("findFreePort: %v", err)
		}
		if seen[port] {
			t.Errorf("duplicate port %d", port)
		}
		seen[port] = true
	}
}

func TestStartMissingDataDir(t *testing.T) {
	_, err := Start(context.Background(), Config{})
	if err == nil {
		t.Fatal("expected error for empty DataDir")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := Config{DataDir: "/tmp/test"}
	if cfg.Database != "" {
		t.Errorf("Database should be zero value before Start fills defaults")
	}
	if cfg.Username != "" {
		t.Errorf("Username should be zero value before Start fills defaults")
	}
}
