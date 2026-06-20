package limeleaf

import (
	"runtime"
	"strings"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	platform, err := detectPlatform()
	if err != nil {
		t.Fatalf("detectPlatform() on %s/%s: %v", runtime.GOOS, runtime.GOARCH, err)
	}

	if !strings.Contains(platform, runtime.GOOS) {
		t.Errorf("platform %q does not contain OS %q", platform, runtime.GOOS)
	}
	if !strings.Contains(platform, runtime.GOARCH) {
		t.Errorf("platform %q does not contain arch %q", platform, runtime.GOARCH)
	}
}

func TestTarballURL(t *testing.T) {
	url := tarballURL("linux-amd64")
	want := "https://github.com/faizalv/limeleaf/releases/download/" + ReleaseTag + "/linux-amd64.tar.gz"
	if url != want {
		t.Errorf("tarballURL = %q, want %q", url, want)
	}
}

func TestChecksumURL(t *testing.T) {
	url := checksumURL("darwin-arm64")
	want := "https://github.com/faizalv/limeleaf/releases/download/" + ReleaseTag + "/darwin-arm64.tar.gz.sha256"
	if url != want {
		t.Errorf("checksumURL = %q, want %q", url, want)
	}
}
