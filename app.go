package main

import (
	"context"
	"log"

	"github.com/iqbaleff214/todo-app/internal/models"
	"github.com/iqbaleff214/todo-app/internal/repository"
	"github.com/iqbaleff214/todo-app/internal/service"
	"github.com/iqbaleff214/todo-app/internal/window"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx             context.Context
	taskService     *service.TaskService
	settingsService *service.SettingsService
	windowManager   *window.Manager
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	dataDir := repository.DataDir()

	if err := repository.MaybeBackup(dataDir); err != nil {
		log.Printf("startup: backup warning: %v", err)
	}

	db, err := repository.Open(dataDir)
	if err != nil {
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{ //nolint:errcheck
			Type:    runtime.ErrorDialog,
			Title:   "Database Error",
			Message: "Could not open database: " + err.Error(),
		})
		runtime.Quit(ctx)
		return
	}

	taskRepo := repository.NewTaskRepository(db)
	settingsRepo := repository.NewSettingsRepository(dataDir)

	a.taskService = service.NewTaskService(taskRepo)
	a.settingsService = service.NewSettingsService(settingsRepo)
	a.windowManager = window.NewManager(ctx)
}

// Task methods — delegates to TaskService.

func (a *App) GetTasksForDate(date string) ([]models.Task, error) {
	return a.taskService.GetTasksForDate(date)
}

func (a *App) GetTodayTasks() ([]models.Task, error) {
	return a.taskService.GetTodayTasks()
}

func (a *App) GetDatesWithTasks() ([]string, error) {
	return a.taskService.GetDatesWithTasks()
}

func (a *App) AddTask(date, text string) (models.Task, error) {
	return a.taskService.AddTask(date, text)
}

func (a *App) ToggleDone(id string) (models.Task, error) {
	return a.taskService.ToggleDone(id)
}

func (a *App) UpdateText(id, text string) (models.Task, error) {
	return a.taskService.UpdateText(id, text)
}

func (a *App) DeleteTask(id string) error {
	return a.taskService.DeleteTask(id)
}

func (a *App) ReorderTasks(ids []string) error {
	return a.taskService.ReorderTasks(ids)
}

// Settings methods — delegates to SettingsService.

func (a *App) GetSettings() (models.Settings, error) {
	return a.settingsService.GetSettings()
}

func (a *App) SaveSettings(settings models.Settings) error {
	return a.settingsService.SaveSettings(settings)
}

func (a *App) ResetWindowPosition() error {
	return a.settingsService.ResetWindowPosition()
}
