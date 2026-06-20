package limeleaf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultCacheDir(t *testing.T) {
	dir := defaultCacheDir()
	if dir == "" {
		t.Fatal("defaultCacheDir returned empty string")
	}

	home, err := os.UserHomeDir()
	if err == nil {
		want := filepath.Join(home, ".limeleaf", "cache")
		if dir != want {
			t.Errorf("defaultCacheDir = %q, want %q", dir, want)
		}
	}
}
