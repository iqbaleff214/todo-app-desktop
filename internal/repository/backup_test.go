package repository_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/iqbaleff214/todo-app/internal/repository"
)

const (
	testDBFile  = "data.db"
	testBakFile = "data.db.bak"
)

func writeDB(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, testDBFile), []byte(content), 0o644); err != nil {
		t.Fatalf("writeDB: %v", err)
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readFile(%q): %v", path, err)
	}
	return b
}

// --- Core done criteria ---

func TestMaybeBackup_CreatesBackupAndContentMatches(t *testing.T) {
	dir := t.TempDir()
	const content = "fake sqlite data"
	writeDB(t, dir, content)

	if err := repository.MaybeBackup(dir); err != nil {
		t.Fatalf("MaybeBackup: %v", err)
	}

	bakPath := filepath.Join(dir, testBakFile)
	if _, err := os.Stat(bakPath); err != nil {
		t.Fatalf("data.db.bak not created: %v", err)
	}

	got := readFile(t, bakPath)
	if !bytes.Equal(got, []byte(content)) {
		t.Errorf("backup content = %q, want %q", got, content)
	}
}

func TestMaybeBackup_SecondCallUnchangedFileDoesNotRecopy(t *testing.T) {
	dir := t.TempDir()
	writeDB(t, dir, "original")

	// First call — creates .bak.
	if err := repository.MaybeBackup(dir); err != nil {
		t.Fatalf("first MaybeBackup: %v", err)
	}

	bakPath := filepath.Join(dir, testBakFile)
	bakInfoBefore, err := os.Stat(bakPath)
	if err != nil {
		t.Fatalf("stat bak before second call: %v", err)
	}

	// Second call — data.db is unchanged (not newer than .bak).
	if err := repository.MaybeBackup(dir); err != nil {
		t.Fatalf("second MaybeBackup: %v", err)
	}

	bakInfoAfter, err := os.Stat(bakPath)
	if err != nil {
		t.Fatalf("stat bak after second call: %v", err)
	}

	// Modification time must not have changed — no re-copy occurred.
	if !bakInfoAfter.ModTime().Equal(bakInfoBefore.ModTime()) {
		t.Errorf("bak mtime changed on second call (%.9fs → %.9fs); file was re-copied unexpectedly",
			float64(bakInfoBefore.ModTime().UnixNano())/1e9,
			float64(bakInfoAfter.ModTime().UnixNano())/1e9)
	}
}

// --- Additional scenarios ---

func TestMaybeBackup_NoDBFileReturnsNil(t *testing.T) {
	dir := t.TempDir()
	// No data.db present.
	if err := repository.MaybeBackup(dir); err != nil {
		t.Errorf("MaybeBackup with no db = %v, want nil", err)
	}
	// .bak must not be created either.
	if _, err := os.Stat(filepath.Join(dir, testBakFile)); !os.IsNotExist(err) {
		t.Error("data.db.bak should not exist when data.db is absent")
	}
}

func TestMaybeBackup_UpdatedDBOverwritesBackup(t *testing.T) {
	dir := t.TempDir()
	writeDB(t, dir, "version-1")

	// First call — create initial backup.
	if err := repository.MaybeBackup(dir); err != nil {
		t.Fatalf("first MaybeBackup: %v", err)
	}

	// Advance mtime so data.db is strictly newer than .bak.
	dbPath := filepath.Join(dir, testDBFile)
	future := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(dbPath, future, future); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}
	// Update content to confirm overwrite.
	if err := os.WriteFile(dbPath, []byte("version-2"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.Chtimes(dbPath, future, future); err != nil {
		t.Fatalf("Chtimes after write: %v", err)
	}

	// Second call — should overwrite backup.
	if err := repository.MaybeBackup(dir); err != nil {
		t.Fatalf("second MaybeBackup: %v", err)
	}

	got := readFile(t, filepath.Join(dir, testBakFile))
	if !bytes.Equal(got, []byte("version-2")) {
		t.Errorf("backup content = %q, want %q", got, "version-2")
	}
}

func TestMaybeBackup_BackupContentIsIdentical(t *testing.T) {
	dir := t.TempDir()
	// Use a larger payload to exercise io.Copy properly.
	content := bytes.Repeat([]byte("abcdefgh"), 4096) // 32 KiB
	if err := os.WriteFile(filepath.Join(dir, testDBFile), content, 0o644); err != nil {
		t.Fatalf("write large db: %v", err)
	}

	if err := repository.MaybeBackup(dir); err != nil {
		t.Fatalf("MaybeBackup: %v", err)
	}

	got := readFile(t, filepath.Join(dir, testBakFile))
	if !bytes.Equal(got, content) {
		t.Errorf("large backup mismatch: got %d bytes, want %d bytes", len(got), len(content))
	}
}
