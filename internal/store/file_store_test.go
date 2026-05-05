package store

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFileStoreReadWriteExistsPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.dat")
	store := NewFileStore(path)

	if store.Path() != path {
		t.Fatalf("Path() = %q, want %q", store.Path(), path)
	}
	if store.Exists() {
		t.Fatal("Exists() = true before file is written, want false")
	}

	want := []byte("encrypted vault bytes")
	if err := store.Write(want); err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}
	if !store.Exists() {
		t.Fatal("Exists() = false after file is written, want true")
	}

	got, err := store.Read()
	if err != nil {
		t.Fatalf("Read() returned error: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("Read() = %q, want %q", got, want)
	}
}

func TestFileStoreOverwriteReplacesExistingContent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.dat")
	store := NewFileStore(path)

	if err := store.Write([]byte("first vault")); err != nil {
		t.Fatalf("first Write() returned error: %v", err)
	}
	if err := store.Write([]byte("second vault")); err != nil {
		t.Fatalf("second Write() returned error: %v", err)
	}

	got, err := store.Read()
	if err != nil {
		t.Fatalf("Read() returned error: %v", err)
	}
	if string(got) != "second vault" {
		t.Fatalf("Read() = %q, want %q", got, "second vault")
	}
}

func TestFileStoreReadMissingReturnsError(t *testing.T) {
	store := NewFileStore(filepath.Join(t.TempDir(), "missing.dat"))

	if _, err := store.Read(); err == nil {
		t.Fatal("Read() error = nil, want an error")
	}
}

func TestFileStoreWriteCreatesParentDirectory(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "nested", "vault")
	path := filepath.Join(parent, "vault.dat")
	store := NewFileStore(path)

	if err := store.Write([]byte("vault")); err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}

	info, err := os.Stat(parent)
	if err != nil {
		t.Fatalf("parent directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("parent path is not a directory")
	}

	if runtime.GOOS != "windows" && info.Mode().Perm() != 0700 {
		t.Fatalf("parent directory mode = %o, want 700", info.Mode().Perm())
	}
}

func TestFileStoreWriteFailsWhenParentPathIsFile(t *testing.T) {
	parentFile := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(parentFile, []byte("file"), 0600); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}
	store := NewFileStore(filepath.Join(parentFile, "vault.dat"))

	err := store.Write([]byte("vault"))

	if err == nil {
		t.Fatal("Write() error = nil, want an error")
	}
	if !strings.Contains(err.Error(), "create parent directory") {
		t.Fatalf("Write() error = %q, want parent directory context", err.Error())
	}
}

func TestFileStoreWriteFailsWhenDestinationIsDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.dat")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatalf("Mkdir() returned error: %v", err)
	}
	store := NewFileStore(path)

	err := store.Write([]byte("vault"))

	if err == nil {
		t.Fatal("Write() error = nil, want an error")
	}
	if !strings.Contains(err.Error(), "rename temp file") {
		t.Fatalf("Write() error = %q, want rename context", err.Error())
	}
}

func TestFileStoreWriteUsesPrivateFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not reliably supported on Windows")
	}

	path := filepath.Join(t.TempDir(), "vault.dat")
	store := NewFileStore(path)

	if err := store.Write([]byte("vault")); err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() returned error: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("file mode = %o, want 600", info.Mode().Perm())
	}
}

func TestFileStoreExistsFalseWhenPathIsMissing(t *testing.T) {
	store := NewFileStore(filepath.Join(t.TempDir(), "missing.dat"))

	if store.Exists() {
		t.Fatal("Exists() = true, want false")
	}
}
