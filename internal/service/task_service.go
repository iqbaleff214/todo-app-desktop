package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iqbaleff214/todo-app/internal/models"
	"github.com/iqbaleff214/todo-app/internal/repository"
)

var (
	ErrEmptyText    = errors.New("task text must not be empty")
	ErrTextTooLong  = errors.New("task text must not exceed 500 characters")
	ErrTaskNotFound = errors.New("task not found")
)

type TaskService struct {
	repo repository.TaskRepository
}

func NewTaskService(repo repository.TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) GetTasksForDate(date string) ([]models.Task, error) {
	tasks, err := s.repo.GetByDate(date)
	if err != nil {
		return nil, fmt.Errorf("task_service.GetTasksForDate: %w", err)
	}
	return tasks, nil
}

func (s *TaskService) GetTodayTasks() ([]models.Task, error) {
	return s.GetTasksForDate(time.Now().Format("2006-01-02"))
}

func (s *TaskService) GetDatesWithTasks() ([]string, error) {
	dates, err := s.repo.GetDatesWithTasks()
	if err != nil {
		return nil, fmt.Errorf("task_service.GetDatesWithTasks: %w", err)
	}
	return dates, nil
}

func (s *TaskService) AddTask(date, text string) (models.Task, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return models.Task{}, ErrEmptyText
	}
	if len(text) > 500 {
		return models.Task{}, ErrTextTooLong
	}

	existing, err := s.repo.GetByDate(date)
	if err != nil {
		return models.Task{}, fmt.Errorf("task_service.AddTask: %w", err)
	}

	task := models.Task{
		ID:       models.NewID(),
		Date:     date,
		Text:     text,
		Done:     false,
		Position: len(existing),
	}

	if err := s.repo.Create(task); err != nil {
		return models.Task{}, fmt.Errorf("task_service.AddTask: %w", err)
	}

	created, err := s.repo.GetByID(task.ID)
	if err != nil {
		return models.Task{}, fmt.Errorf("task_service.AddTask: %w", err)
	}
	return created, nil
}

func (s *TaskService) ToggleDone(id string) (models.Task, error) {
	task, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Task{}, ErrTaskNotFound
		}
		return models.Task{}, fmt.Errorf("task_service.ToggleDone: %w", err)
	}
	task.Done = !task.Done
	if err := s.repo.Update(task); err != nil {
		return models.Task{}, fmt.Errorf("task_service.ToggleDone: %w", err)
	}
	return task, nil
}

func (s *TaskService) UpdateText(id, text string) (models.Task, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return models.Task{}, ErrEmptyText
	}
	if len(text) > 500 {
		return models.Task{}, ErrTextTooLong
	}

	task, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Task{}, ErrTaskNotFound
		}
		return models.Task{}, fmt.Errorf("task_service.UpdateText: %w", err)
	}
	task.Text = text
	if err := s.repo.Update(task); err != nil {
		return models.Task{}, fmt.Errorf("task_service.UpdateText: %w", err)
	}
	return task, nil
}

func (s *TaskService) DeleteTask(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("task_service.DeleteTask: %w", err)
	}
	return nil
}

func (s *TaskService) ReorderTasks(ids []string) error {
	if err := s.repo.ReorderPositions(ids); err != nil {
		return fmt.Errorf("task_service.ReorderTasks: %w", err)
	}
	return nil
}
