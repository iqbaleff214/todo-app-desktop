package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/iqbaleff214/todo-app/internal/models"
)

type TaskRepository interface {
	GetByDate(date string) ([]models.Task, error)
	GetByID(id string) (models.Task, error)
	GetDatesWithTasks() ([]string, error)
	Create(task models.Task) error
	Update(task models.Task) error
	Delete(id string) error
	ReorderPositions(ids []string) error
}

type sqliteTaskRepo struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) TaskRepository {
	return &sqliteTaskRepo{db: db}
}

func (r *sqliteTaskRepo) GetByDate(date string) ([]models.Task, error) {
	rows, err := r.db.Query(`
		SELECT id, date, text, done, position, created_at, updated_at
		FROM tasks
		WHERE date = ?
		ORDER BY position ASC, done ASC
	`, date)
	if err != nil {
		return nil, fmt.Errorf("task_repo.GetByDate: %w", err)
	}
	defer rows.Close()

	tasks := []models.Task{} // always non-nil
	for rows.Next() {
		var t models.Task
		var done int
		var createdAt, updatedAt string
		if err := rows.Scan(&t.ID, &t.Date, &t.Text, &done, &t.Position, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("task_repo.GetByDate: scan: %w", err)
		}
		t.Done = done == 1
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task_repo.GetByDate: rows: %w", err)
	}
	return tasks, nil
}

func (r *sqliteTaskRepo) GetByID(id string) (models.Task, error) {
	var t models.Task
	var done int
	var createdAt, updatedAt string
	err := r.db.QueryRow(`
		SELECT id, date, text, done, position, created_at, updated_at
		FROM tasks
		WHERE id = ?
	`, id).Scan(&t.ID, &t.Date, &t.Text, &done, &t.Position, &createdAt, &updatedAt)
	if err != nil {
		return models.Task{}, fmt.Errorf("task_repo.GetByID: %w", err)
	}
	t.Done = done == 1
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return t, nil
}

func (r *sqliteTaskRepo) GetDatesWithTasks() ([]string, error) {
	rows, err := r.db.Query(`
		SELECT DISTINCT date FROM tasks
		ORDER BY date DESC
		LIMIT 365
	`)
	if err != nil {
		return nil, fmt.Errorf("task_repo.GetDatesWithTasks: %w", err)
	}
	defer rows.Close()

	dates := []string{} // always non-nil
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, fmt.Errorf("task_repo.GetDatesWithTasks: scan: %w", err)
		}
		dates = append(dates, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task_repo.GetDatesWithTasks: rows: %w", err)
	}
	return dates, nil
}

func (r *sqliteTaskRepo) Create(task models.Task) error {
	if task.ID == "" {
		task.ID = models.NewID()
	}
	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now

	_, err := r.db.Exec(`
		INSERT INTO tasks (id, date, text, done, position, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		task.ID,
		task.Date,
		task.Text,
		boolToInt(task.Done),
		task.Position,
		task.CreatedAt.Format(time.RFC3339),
		task.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("task_repo.Create: %w", err)
	}
	return nil
}

func (r *sqliteTaskRepo) Update(task models.Task) error {
	task.UpdatedAt = time.Now().UTC()

	_, err := r.db.Exec(`
		UPDATE tasks
		SET date = ?, text = ?, done = ?, position = ?, updated_at = ?
		WHERE id = ?
	`,
		task.Date,
		task.Text,
		boolToInt(task.Done),
		task.Position,
		task.UpdatedAt.Format(time.RFC3339),
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("task_repo.Update: %w", err)
	}
	return nil
}

func (r *sqliteTaskRepo) Delete(id string) error {
	if _, err := r.db.Exec(`DELETE FROM tasks WHERE id = ?`, id); err != nil {
		return fmt.Errorf("task_repo.Delete: %w", err)
	}
	return nil
}

func (r *sqliteTaskRepo) ReorderPositions(ids []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("task_repo.ReorderPositions: begin: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.Prepare(`UPDATE tasks SET position = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("task_repo.ReorderPositions: prepare: %w", err)
	}
	defer stmt.Close()

	now := time.Now().UTC().Format(time.RFC3339)
	for i, id := range ids {
		if _, err := stmt.Exec(i, now, id); err != nil {
			return fmt.Errorf("task_repo.ReorderPositions: exec %q: %w", id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("task_repo.ReorderPositions: commit: %w", err)
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
