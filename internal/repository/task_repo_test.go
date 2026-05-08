package repository_test

import (
	"testing"
	"time"

	"github.com/iqbaleff214/todo-app/internal/models"
	"github.com/iqbaleff214/todo-app/internal/repository"
)

func newTestTaskRepo(t *testing.T) repository.TaskRepository {
	t.Helper()
	db, err := repository.Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return repository.NewTaskRepository(db)
}

func makeTask(date, text string, position int) models.Task {
	return models.Task{Date: date, Text: text, Position: position}
}

// --- GetByDate ---

func TestTaskRepo_GetByDate_EmptyReturnsNonNilSlice(t *testing.T) {
	repo := newTestTaskRepo(t)
	tasks, err := repo.GetByDate("2024-01-01")
	if err != nil {
		t.Fatalf("GetByDate: %v", err)
	}
	if tasks == nil {
		t.Error("GetByDate returned nil, want empty (non-nil) slice")
	}
	if len(tasks) != 0 {
		t.Errorf("GetByDate returned %d tasks, want 0", len(tasks))
	}
}

func TestTaskRepo_GetByDate_ReturnsOnlyMatchingDate(t *testing.T) {
	repo := newTestTaskRepo(t)

	if err := repo.Create(makeTask("2024-01-01", "task A", 0)); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.Create(makeTask("2024-01-02", "task B", 0)); err != nil {
		t.Fatalf("Create: %v", err)
	}

	tasks, err := repo.GetByDate("2024-01-01")
	if err != nil {
		t.Fatalf("GetByDate: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	if tasks[0].Text != "task A" {
		t.Errorf("text = %q, want %q", tasks[0].Text, "task A")
	}
}

func TestTaskRepo_GetByDate_OrderByPositionThenDone(t *testing.T) {
	repo := newTestTaskRepo(t)
	const date = "2024-01-15"

	tasks := []models.Task{
		{Date: date, Text: "pos2", Position: 2},
		{Date: date, Text: "pos0", Position: 0},
		{Date: date, Text: "pos1-done", Position: 1, Done: true},
	}
	for _, task := range tasks {
		if err := repo.Create(task); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	got, err := repo.GetByDate(date)
	if err != nil {
		t.Fatalf("GetByDate: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d tasks, want 3", len(got))
	}

	wantOrder := []string{"pos0", "pos1-done", "pos2"}
	for i, want := range wantOrder {
		if got[i].Text != want {
			t.Errorf("got[%d].Text = %q, want %q", i, got[i].Text, want)
		}
	}
}

// --- Create ---

func TestTaskRepo_Create_GeneratesID(t *testing.T) {
	repo := newTestTaskRepo(t)
	task := makeTask("2024-01-15", "Buy milk", 0)

	if err := repo.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	tasks, _ := repo.GetByDate("2024-01-15")
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	if tasks[0].ID == "" {
		t.Error("ID was not generated")
	}
}

func TestTaskRepo_Create_UsesExplicitID(t *testing.T) {
	repo := newTestTaskRepo(t)
	task := models.Task{ID: "explicit-id", Date: "2024-01-15", Text: "Walk dog", Position: 0}

	if err := repo.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	tasks, _ := repo.GetByDate("2024-01-15")
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	if tasks[0].ID != "explicit-id" {
		t.Errorf("ID = %q, want %q", tasks[0].ID, "explicit-id")
	}
}

func TestTaskRepo_Create_SetsTimestamps(t *testing.T) {
	repo := newTestTaskRepo(t)
	before := time.Now().UTC().Truncate(time.Second)

	if err := repo.Create(makeTask("2024-01-15", "task", 0)); err != nil {
		t.Fatalf("Create: %v", err)
	}

	tasks, _ := repo.GetByDate("2024-01-15")
	got := tasks[0]

	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt is zero")
	}
	if got.CreatedAt.Before(before) {
		t.Errorf("CreatedAt %v is before test start %v", got.CreatedAt, before)
	}
	if !got.CreatedAt.Equal(got.UpdatedAt) {
		t.Errorf("CreatedAt %v != UpdatedAt %v on newly created task", got.CreatedAt, got.UpdatedAt)
	}
}

// --- Update ---

func TestTaskRepo_Update_ChangesFields(t *testing.T) {
	repo := newTestTaskRepo(t)
	original := models.Task{ID: "u1", Date: "2024-01-15", Text: "original", Position: 0}
	if err := repo.Create(original); err != nil {
		t.Fatalf("Create: %v", err)
	}

	tasks, _ := repo.GetByDate("2024-01-15")
	saved := tasks[0]
	savedCreatedAt := saved.CreatedAt

	// Record time before update; RFC3339 is second-precision so truncate accordingly.
	beforeUpdate := time.Now().UTC().Truncate(time.Second)

	saved.Text = "updated"
	saved.Done = true
	saved.Position = 5
	if err := repo.Update(saved); err != nil {
		t.Fatalf("Update: %v", err)
	}

	tasks, _ = repo.GetByDate("2024-01-15")
	got := tasks[0]

	if got.Text != "updated" {
		t.Errorf("Text = %q, want %q", got.Text, "updated")
	}
	if !got.Done {
		t.Error("Done = false, want true")
	}
	if got.Position != 5 {
		t.Errorf("Position = %d, want 5", got.Position)
	}
	// created_at must never change on Update.
	if !got.CreatedAt.Equal(savedCreatedAt) {
		t.Errorf("CreatedAt changed: got %v, want %v", got.CreatedAt, savedCreatedAt)
	}
	// updated_at must be set to a time >= the moment before Update was called.
	if got.UpdatedAt.Before(beforeUpdate) {
		t.Errorf("UpdatedAt %v is before pre-update timestamp %v", got.UpdatedAt, beforeUpdate)
	}
}

// --- Delete ---

func TestTaskRepo_Delete_RemovesTask(t *testing.T) {
	repo := newTestTaskRepo(t)
	task := models.Task{ID: "del-1", Date: "2024-01-15", Text: "to delete", Position: 0}
	if err := repo.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete("del-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	tasks, _ := repo.GetByDate("2024-01-15")
	if len(tasks) != 0 {
		t.Errorf("got %d tasks after delete, want 0", len(tasks))
	}
}

func TestTaskRepo_Delete_NonExistentIDNoError(t *testing.T) {
	repo := newTestTaskRepo(t)
	if err := repo.Delete("does-not-exist"); err != nil {
		t.Errorf("Delete non-existent ID returned error: %v", err)
	}
}

// --- GetDatesWithTasks ---

func TestTaskRepo_GetDatesWithTasks_Empty(t *testing.T) {
	repo := newTestTaskRepo(t)
	dates, err := repo.GetDatesWithTasks()
	if err != nil {
		t.Fatalf("GetDatesWithTasks: %v", err)
	}
	if dates == nil {
		t.Error("GetDatesWithTasks returned nil, want empty slice")
	}
}

func TestTaskRepo_GetDatesWithTasks_DescendingDistinct(t *testing.T) {
	repo := newTestTaskRepo(t)

	for _, date := range []string{"2024-01-01", "2024-01-03", "2024-01-02"} {
		if err := repo.Create(makeTask(date, "task", 0)); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	// second task on same date — should not create duplicate entry
	if err := repo.Create(makeTask("2024-01-01", "task2", 1)); err != nil {
		t.Fatalf("Create: %v", err)
	}

	dates, err := repo.GetDatesWithTasks()
	if err != nil {
		t.Fatalf("GetDatesWithTasks: %v", err)
	}
	if len(dates) != 3 {
		t.Fatalf("got %d dates, want 3", len(dates))
	}
	wantOrder := []string{"2024-01-03", "2024-01-02", "2024-01-01"}
	for i, want := range wantOrder {
		if dates[i] != want {
			t.Errorf("dates[%d] = %q, want %q", i, dates[i], want)
		}
	}
}

// --- ReorderPositions ---

func TestTaskRepo_ReorderPositions(t *testing.T) {
	repo := newTestTaskRepo(t)
	const date = "2024-01-15"

	ids := []string{"r1", "r2", "r3"}
	for i, id := range ids {
		task := models.Task{ID: id, Date: date, Text: "task " + id, Position: i}
		if err := repo.Create(task); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	// reverse the order
	reversed := []string{"r3", "r2", "r1"}
	if err := repo.ReorderPositions(reversed); err != nil {
		t.Fatalf("ReorderPositions: %v", err)
	}

	tasks, err := repo.GetByDate(date)
	if err != nil {
		t.Fatalf("GetByDate: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("got %d tasks, want 3", len(tasks))
	}

	wantOrder := []string{"r3", "r2", "r1"}
	for i, wantID := range wantOrder {
		if tasks[i].ID != wantID {
			t.Errorf("tasks[%d].ID = %q, want %q", i, tasks[i].ID, wantID)
		}
		if tasks[i].Position != i {
			t.Errorf("tasks[%d].Position = %d, want %d", i, tasks[i].Position, i)
		}
	}
}

func TestTaskRepo_ReorderPositions_EmptySlice(t *testing.T) {
	repo := newTestTaskRepo(t)
	if err := repo.ReorderPositions([]string{}); err != nil {
		t.Errorf("ReorderPositions with empty slice returned error: %v", err)
	}
}
