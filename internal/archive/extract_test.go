package archive

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func createTestTarGz(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.tar.gz")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	return path
}

func TestExtractTarGz(t *testing.T) {
	files := map[string]string{
		"bin/postgres": "fake-binary",
		"lib/vector.so": "fake-extension",
		"share/extension/vector.sql": "CREATE EXTENSION;",
	}

	src := createTestTarGz(t, files)
	dst := t.TempDir()

	if err := ExtractTarGz(src, dst); err != nil {
		t.Fatalf("ExtractTarGz: %v", err)
	}

	for name, want := range files {
		data, err := os.ReadFile(filepath.Join(dst, name))
		if err != nil {
			t.Errorf("file %s not extracted: %v", name, err)
			continue
		}
		if string(data) != want {
			t.Errorf("file %s content = %q, want %q", name, data, want)
		}
	}
}

func TestExtractTarGzPathTraversal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "evil.tar.gz")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	tw.WriteHeader(&tar.Header{
		Name: "../../../etc/passwd",
		Mode: 0644,
		Size: 5,
	})
	tw.Write([]byte("owned"))
	tw.Close()
	gw.Close()
	f.Close()

	dst := t.TempDir()
	err = ExtractTarGz(path, dst)
	if err == nil {
		t.Fatal("expected path traversal error, got nil")
	}
}

func TestExtractTarGzWithDirs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.tar.gz")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	tw.WriteHeader(&tar.Header{
		Name:     "subdir/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	})
	tw.WriteHeader(&tar.Header{
		Name: "subdir/file.txt",
		Mode: 0644,
		Size: 5,
	})
	tw.Write([]byte("hello"))
	tw.Close()
	gw.Close()
	f.Close()

	dst := t.TempDir()
	if err := ExtractTarGz(path, dst); err != nil {
		t.Fatalf("ExtractTarGz: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "subdir", "file.txt"))
	if err != nil {
		t.Fatalf("file not extracted: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("content = %q, want %q", data, "hello")
	}
}
