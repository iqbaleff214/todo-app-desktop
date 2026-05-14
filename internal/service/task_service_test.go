package service_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/iqbaleff214/todo-app/internal/models"
	"github.com/iqbaleff214/todo-app/internal/service"
)

// mockTaskRepo implements repository.TaskRepository for unit testing.
type mockTaskRepo struct {
	tasks  map[string]models.Task
	byDate map[string][]models.Task
}

func newMockTaskRepo() *mockTaskRepo {
	return &mockTaskRepo{
		tasks:  make(map[string]models.Task),
		byDate: make(map[string][]models.Task),
	}
}

func (m *mockTaskRepo) GetByDate(date string) ([]models.Task, error) {
	tasks := m.byDate[date]
	if tasks == nil {
		return []models.Task{}, nil
	}
	return tasks, nil
}

func (m *mockTaskRepo) GetByID(id string) (models.Task, error) {
	t, ok := m.tasks[id]
	if !ok {
		return models.Task{}, sql.ErrNoRows
	}
	return t, nil
}

func (m *mockTaskRepo) GetDatesWithTasks() ([]string, error) {
	dates := []string{}
	for d, tasks := range m.byDate {
		if len(tasks) > 0 {
			dates = append(dates, d)
		}
	}
	return dates, nil
}

func (m *mockTaskRepo) Create(task models.Task) error {
	m.tasks[task.ID] = task
	m.byDate[task.Date] = append(m.byDate[task.Date], task)
	return nil
}

func (m *mockTaskRepo) Update(task models.Task) error {
	if _, ok := m.tasks[task.ID]; !ok {
		return nil
	}
	m.tasks[task.ID] = task
	for i, t := range m.byDate[task.Date] {
		if t.ID == task.ID {
			m.byDate[task.Date][i] = task
			break
		}
	}
	return nil
}

func (m *mockTaskRepo) Delete(id string) error {
	task, ok := m.tasks[id]
	if !ok {
		return nil
	}
	delete(m.tasks, id)
	tasks := m.byDate[task.Date]
	for i, t := range tasks {
		if t.ID == id {
			m.byDate[task.Date] = append(tasks[:i], tasks[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockTaskRepo) ReorderPositions(ids []string) error {
	for i, id := range ids {
		t, ok := m.tasks[id]
		if !ok {
			continue
		}
		t.Position = i
		m.tasks[id] = t
	}
	return nil
}

// --- AddTask ---

func TestTaskService_AddTask_EmptyText(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	_, err := svc.AddTask("2024-01-01", "")
	if !errors.Is(err, service.ErrEmptyText) {
		t.Errorf("got %v, want ErrEmptyText", err)
	}
}

func TestTaskService_AddTask_WhitespaceOnlyText(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	_, err := svc.AddTask("2024-01-01", "   ")
	if !errors.Is(err, service.ErrEmptyText) {
		t.Errorf("got %v, want ErrEmptyText", err)
	}
}

func TestTaskService_AddTask_TextTooLong(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	_, err := svc.AddTask("2024-01-01", strings.Repeat("x", 501))
	if !errors.Is(err, service.ErrTextTooLong) {
		t.Errorf("got %v, want ErrTextTooLong", err)
	}
}

func TestTaskService_AddTask_Success(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	task, err := svc.AddTask("2024-01-01", "buy milk")
	if err != nil {
		t.Fatal(err)
	}
	if task.ID == "" {
		t.Error("expected non-empty ID")
	}
	if task.Text != "buy milk" {
		t.Errorf("text: got %q, want %q", task.Text, "buy milk")
	}
	if task.Date != "2024-01-01" {
		t.Errorf("date: got %q, want %q", task.Date, "2024-01-01")
	}
	if task.Done {
		t.Error("new task should not be done")
	}
	if task.Position != 0 {
		t.Errorf("position: got %d, want 0", task.Position)
	}
}

func TestTaskService_AddTask_TrimsWhitespace(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	task, err := svc.AddTask("2024-01-01", "  buy milk  ")
	if err != nil {
		t.Fatal(err)
	}
	if task.Text != "buy milk" {
		t.Errorf("text: got %q, want %q", task.Text, "buy milk")
	}
}

func TestTaskService_AddTask_Position(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	svc.AddTask("2024-01-01", "first")  //nolint:errcheck
	svc.AddTask("2024-01-01", "second") //nolint:errcheck
	task, err := svc.AddTask("2024-01-01", "third")
	if err != nil {
		t.Fatal(err)
	}
	if task.Position != 2 {
		t.Errorf("position: got %d, want 2", task.Position)
	}
}

func TestTaskService_AddTask_MaxLengthText(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	_, err := svc.AddTask("2024-01-01", strings.Repeat("x", 500))
	if err != nil {
		t.Errorf("500-char text should be accepted, got: %v", err)
	}
}

// --- ToggleDone ---

func TestTaskService_ToggleDone_NotFound(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	_, err := svc.ToggleDone("nonexistent-id")
	if !errors.Is(err, service.ErrTaskNotFound) {
		t.Errorf("got %v, want ErrTaskNotFound", err)
	}
}

func TestTaskService_ToggleDone_FlipsDone(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	task, _ := svc.AddTask("2024-01-01", "buy milk")
	if task.Done {
		t.Fatal("new task must not be done")
	}

	toggled, err := svc.ToggleDone(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !toggled.Done {
		t.Error("task should be done after first toggle")
	}

	toggled2, err := svc.ToggleDone(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if toggled2.Done {
		t.Error("task should not be done after second toggle")
	}
}

// --- UpdateText ---

func TestTaskService_UpdateText_EmptyText(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	task, _ := svc.AddTask("2024-01-01", "buy milk")
	_, err := svc.UpdateText(task.ID, "")
	if !errors.Is(err, service.ErrEmptyText) {
		t.Errorf("got %v, want ErrEmptyText", err)
	}
}

func TestTaskService_UpdateText_TextTooLong(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	task, _ := svc.AddTask("2024-01-01", "buy milk")
	_, err := svc.UpdateText(task.ID, strings.Repeat("x", 501))
	if !errors.Is(err, service.ErrTextTooLong) {
		t.Errorf("got %v, want ErrTextTooLong", err)
	}
}

func TestTaskService_UpdateText_NotFound(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	_, err := svc.UpdateText("nonexistent", "new text")
	if !errors.Is(err, service.ErrTaskNotFound) {
		t.Errorf("got %v, want ErrTaskNotFound", err)
	}
}

func TestTaskService_UpdateText_Success(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	task, _ := svc.AddTask("2024-01-01", "buy milk")
	updated, err := svc.UpdateText(task.ID, "buy coffee")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Text != "buy coffee" {
		t.Errorf("text: got %q, want %q", updated.Text, "buy coffee")
	}
}

// --- DeleteTask ---

func TestTaskService_DeleteTask_Success(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	task, _ := svc.AddTask("2024-01-01", "buy milk")
	if err := svc.DeleteTask(task.ID); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTaskService_DeleteTask_MissingID_NoError(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	if err := svc.DeleteTask("nonexistent"); err != nil {
		t.Errorf("delete of missing id should not error, got: %v", err)
	}
}

// --- GetTasksForDate ---

func TestTaskService_GetTasksForDate_Empty(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	tasks, err := svc.GetTasksForDate("2024-01-01")
	if err != nil {
		t.Fatal(err)
	}
	if tasks == nil {
		t.Error("expected non-nil slice for empty date")
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestTaskService_GetTasksForDate_ReturnsTasks(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	svc.AddTask("2024-01-01", "task 1") //nolint:errcheck
	svc.AddTask("2024-01-01", "task 2") //nolint:errcheck
	tasks, err := svc.GetTasksForDate("2024-01-01")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

// --- GetTodayTasks ---

func TestTaskService_GetTodayTasks_NonNil(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	tasks, err := svc.GetTodayTasks()
	if err != nil {
		t.Fatal(err)
	}
	if tasks == nil {
		t.Error("expected non-nil slice")
	}
}

// --- GetDatesWithTasks ---

func TestTaskService_GetDatesWithTasks(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	svc.AddTask("2024-01-01", "task") //nolint:errcheck
	svc.AddTask("2024-01-02", "task") //nolint:errcheck
	dates, err := svc.GetDatesWithTasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(dates) != 2 {
		t.Errorf("expected 2 dates, got %d", len(dates))
	}
}

// --- ReorderTasks ---

func TestTaskService_ReorderTasks(t *testing.T) {
	svc := service.NewTaskService(newMockTaskRepo())
	t1, _ := svc.AddTask("2024-01-01", "first")
	t2, _ := svc.AddTask("2024-01-01", "second")
	if err := svc.ReorderTasks([]string{t2.ID, t1.ID}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
