package service

import (
	"fmt"
	"sync"

	"github.com/iqbaleff214/todo-app/internal/models"
	"github.com/iqbaleff214/todo-app/internal/repository"
)

type SettingsService struct {
	repo   repository.SettingsRepository
	mu     sync.RWMutex
	cached *models.Settings
}

func NewSettingsService(repo repository.SettingsRepository) *SettingsService {
	return &SettingsService{repo: repo}
}

func (s *SettingsService) GetSettings() (models.Settings, error) {
	s.mu.RLock()
	if s.cached != nil {
		settings := *s.cached
		s.mu.RUnlock()
		return settings, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached != nil {
		return *s.cached, nil
	}

	settings, err := s.repo.Load()
	if err != nil {
		return models.Settings{}, fmt.Errorf("settings_service.GetSettings: %w", err)
	}
	s.cached = &settings
	return settings, nil
}

func (s *SettingsService) SaveSettings(settings models.Settings) error {
	if err := settings.ValidateTheme(); err != nil {
		return fmt.Errorf("settings_service.SaveSettings: %w", err)
	}
	settings.ClampOpacity()

	if err := s.repo.Save(settings); err != nil {
		return fmt.Errorf("settings_service.SaveSettings: %w", err)
	}

	s.mu.Lock()
	s.cached = &settings
	s.mu.Unlock()
	return nil
}

func (s *SettingsService) ResetWindowPosition() error {
	settings, err := s.GetSettings()
	if err != nil {
		return fmt.Errorf("settings_service.ResetWindowPosition: %w", err)
	}
	settings.WindowX = -1
	settings.WindowY = -1
	return s.SaveSettings(settings)
}
