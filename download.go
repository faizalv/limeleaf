package limeleaf

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/faizalv/limeleaf/internal/archive"
)

func defaultCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "limeleaf", "cache")
	}
	return filepath.Join(home, ".limeleaf", "cache")
}

// EnsureBinary downloads and extracts the Postgres binary if it is not already
// cached. Returns the path to the directory containing bin/, lib/, share/.
func EnsureBinary(ctx context.Context, cacheDir string) (string, error) {
	if cacheDir == "" {
		cacheDir = defaultCacheDir()
	}

	binDir := filepath.Join(cacheDir, ReleaseTag)

	if _, err := os.Stat(filepath.Join(binDir, "bin", "postgres")); err == nil {
		return binDir, nil
	}

	platform, err := detectPlatform()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("creating cache dir: %w", err)
	}

	tarballPath := filepath.Join(cacheDir, fmt.Sprintf("%s-%s.tar.gz", ReleaseTag, platform))
	if err := downloadFile(ctx, tarballURL(platform), tarballPath); err != nil {
		return "", fmt.Errorf("downloading tarball: %w", err)
	}

	if err := os.MkdirAll(binDir, 0755); err != nil {
		os.Remove(tarballPath)
		return "", fmt.Errorf("creating bin dir: %w", err)
	}

	if err := archive.ExtractTarGz(tarballPath, binDir); err != nil {
		os.RemoveAll(binDir)
		os.Remove(tarballPath)
		return "", fmt.Errorf("extracting tarball: %w", err)
	}

	os.Remove(tarballPath)
	return binDir, nil
}

func downloadFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
