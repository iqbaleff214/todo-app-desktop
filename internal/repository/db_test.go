package repository_test

import (
	"strings"
	"testing"

	"github.com/iqbaleff214/todo-app/internal/repository"
)

func TestOpen(t *testing.T) {
	db, err := repository.Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	// WAL mode must be active.
	var mode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Errorf("journal_mode = %q, want %q", mode, "wal")
	}

	// Foreign keys must be enabled.
	var fk int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatalf("query foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, want 1", fk)
	}

	// tasks table must exist with the correct columns.
	rows, err := db.Query("PRAGMA table_info(tasks)")
	if err != nil {
		t.Fatalf("table_info: %v", err)
	}
	defer rows.Close()

	cols := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dflt, &pk); err != nil {
			t.Fatalf("scan table_info: %v", err)
		}
		cols[name] = true
	}
	for _, want := range []string{"id", "date", "text", "done", "position", "created_at", "updated_at"} {
		if !cols[want] {
			t.Errorf("tasks table missing column %q", want)
		}
	}

	// idx_tasks_date index must exist.
	var idxName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_tasks_date'").Scan(&idxName)
	if err != nil {
		t.Fatalf("idx_tasks_date not found: %v", err)
	}
}

func TestOpen_InsertAndRead(t *testing.T) {
	db, err := repository.Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer db.Close()

	const (
		wantID       = "abc-123"
		wantDate     = "2024-05-07"
		wantText     = "Buy milk"
		wantDone     = 0
		wantPosition = 0
		wantCreated  = "2024-05-07T10:00:00Z"
		wantUpdated  = "2024-05-07T10:00:00Z"
	)

	_, err = db.Exec(
		`INSERT INTO tasks (id, date, text, done, position, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		wantID, wantDate, wantText, wantDone, wantPosition, wantCreated, wantUpdated,
	)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	var gotID, gotDate, gotText, gotCreated, gotUpdated string
	var gotDone, gotPosition int
	err = db.QueryRow(`SELECT id, date, text, done, position, created_at, updated_at FROM tasks WHERE id = ?`, wantID).
		Scan(&gotID, &gotDate, &gotText, &gotDone, &gotPosition, &gotCreated, &gotUpdated)
	if err != nil {
		t.Fatalf("select: %v", err)
	}

	if gotID != wantID {
		t.Errorf("id = %q, want %q", gotID, wantID)
	}
	if gotDate != wantDate {
		t.Errorf("date = %q, want %q", gotDate, wantDate)
	}
	if gotText != wantText {
		t.Errorf("text = %q, want %q", gotText, wantText)
	}
	if gotDone != wantDone {
		t.Errorf("done = %d, want %d", gotDone, wantDone)
	}
	if gotPosition != wantPosition {
		t.Errorf("position = %d, want %d", gotPosition, wantPosition)
	}
	if gotCreated != wantCreated {
		t.Errorf("created_at = %q, want %q", gotCreated, wantCreated)
	}
	if gotUpdated != wantUpdated {
		t.Errorf("updated_at = %q, want %q", gotUpdated, wantUpdated)
	}
}

func TestOpen_Idempotent(t *testing.T) {
	dir := t.TempDir()

	// Opening the same dir twice must not fail or duplicate the schema.
	for i := 0; i < 2; i++ {
		db, err := repository.Open(dir)
		if err != nil {
			t.Fatalf("Open() attempt %d error: %v", i+1, err)
		}
		db.Close()
	}
}

func TestOpen_CreatesDataDir(t *testing.T) {
	dir := t.TempDir() + "/nested/path"
	db, err := repository.Open(dir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	db.Close()
}

func TestDataDir(t *testing.T) {
	got := repository.DataDir()
	if got == "" {
		t.Fatal("DataDir() returned empty string")
	}
	if !strings.HasSuffix(got, "todo-app") {
		t.Errorf("DataDir() = %q, want suffix %q", got, "todo-app")
	}
}
