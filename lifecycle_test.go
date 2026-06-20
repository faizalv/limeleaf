package limeleaf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsFirstRun(t *testing.T) {
	dir := t.TempDir()

	if !isFirstRun(dir) {
		t.Error("empty dir should be first run")
	}

	os.WriteFile(filepath.Join(dir, "PG_VERSION"), []byte("16"), 0600)

	if isFirstRun(dir) {
		t.Error("dir with PG_VERSION should not be first run")
	}
}

func TestCleanStalePidNoPidFile(t *testing.T) {
	dir := t.TempDir()
	if err := cleanStalePid(dir); err != nil {
		t.Errorf("cleanStalePid with no pid file: %v", err)
	}
}

func TestCleanStalePidDeadProcess(t *testing.T) {
	dir := t.TempDir()
	// PID 2147483647 is almost certainly not running
	os.WriteFile(filepath.Join(dir, "postmaster.pid"), []byte("2147483647\n"), 0600)

	if err := cleanStalePid(dir); err != nil {
		t.Errorf("cleanStalePid with dead process: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "postmaster.pid")); !os.IsNotExist(err) {
		t.Error("stale pid file should have been removed")
	}
}

func TestCleanStalePidMalformed(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "postmaster.pid"), []byte("not-a-number\n"), 0600)

	if err := cleanStalePid(dir); err != nil {
		t.Errorf("cleanStalePid with malformed pid: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "postmaster.pid")); !os.IsNotExist(err) {
		t.Error("malformed pid file should have been removed")
	}
}
