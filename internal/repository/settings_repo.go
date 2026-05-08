package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iqbaleff214/todo-app/internal/models"
)

const (
	settingsFilename = "settings.json"
	settingsTmpFile  = "settings.json.tmp"
)

type SettingsRepository interface {
	Load() (models.Settings, error)
	Save(s models.Settings) error
}

type jsonSettingsRepo struct {
	path    string // absolute path to settings.json
	tmpPath string // absolute path to settings.json.tmp
}

func NewSettingsRepository(dataDir string) SettingsRepository {
	return &jsonSettingsRepo{
		path:    filepath.Join(dataDir, settingsFilename),
		tmpPath: filepath.Join(dataDir, settingsTmpFile),
	}
}

// Load reads settings.json and unmarshals it.
// Returns models.Default() when the file does not exist (first launch).
func (r *jsonSettingsRepo) Load() (models.Settings, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return models.Default(), nil
		}
		return models.Settings{}, fmt.Errorf("settings_repo.Load: %w", err)
	}

	var s models.Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return models.Settings{}, fmt.Errorf("settings_repo.Load: unmarshal: %w", err)
	}
	return s, nil
}

// Save writes s to settings.json atomically via a temporary file and os.Rename.
func (r *jsonSettingsRepo) Save(s models.Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("settings_repo.Save: marshal: %w", err)
	}

	if err := os.WriteFile(r.tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("settings_repo.Save: write tmp: %w", err)
	}

	if err := os.Rename(r.tmpPath, r.path); err != nil {
		return fmt.Errorf("settings_repo.Save: rename: %w", err)
	}

	return nil
}
