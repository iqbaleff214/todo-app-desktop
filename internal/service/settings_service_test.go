package service_test

import (
	"sync"
	"testing"

	"github.com/iqbaleff214/todo-app/internal/models"
	"github.com/iqbaleff214/todo-app/internal/service"
)

// mockSettingsRepo implements repository.SettingsRepository for unit testing.
type mockSettingsRepo struct {
	mu        sync.Mutex
	settings  models.Settings
	loadCount int
}

func newMockSettingsRepo() *mockSettingsRepo {
	return &mockSettingsRepo{settings: models.Default()}
}

func (m *mockSettingsRepo) Load() (models.Settings, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadCount++
	return m.settings, nil
}

func (m *mockSettingsRepo) Save(s models.Settings) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = s
	return nil
}

func TestSettingsService_GetSettings_ReturnsDefaults(t *testing.T) {
	svc := service.NewSettingsService(newMockSettingsRepo())
	got, err := svc.GetSettings()
	if err != nil {
		t.Fatal(err)
	}
	want := models.Default()
	if got.Theme != want.Theme {
		t.Errorf("theme: got %q, want %q", got.Theme, want.Theme)
	}
	if got.Opacity != want.Opacity {
		t.Errorf("opacity: got %f, want %f", got.Opacity, want.Opacity)
	}
	if got.WindowX != want.WindowX || got.WindowY != want.WindowY {
		t.Errorf("window pos: got (%d,%d), want (%d,%d)", got.WindowX, got.WindowY, want.WindowX, want.WindowY)
	}
}

func TestSettingsService_GetSettings_CachesAfterFirstLoad(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := service.NewSettingsService(repo)

	svc.GetSettings() //nolint:errcheck
	svc.GetSettings() //nolint:errcheck
	svc.GetSettings() //nolint:errcheck

	repo.mu.Lock()
	count := repo.loadCount
	repo.mu.Unlock()

	if count != 1 {
		t.Errorf("Load called %d times, want exactly 1", count)
	}
}

func TestSettingsService_SaveSettings_InvalidTheme(t *testing.T) {
	svc := service.NewSettingsService(newMockSettingsRepo())
	s := models.Default()
	s.Theme = "rainbow"
	if err := svc.SaveSettings(s); err == nil {
		t.Error("expected error for invalid theme, got nil")
	}
}

func TestSettingsService_SaveSettings_ClampsOpacityLow(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := service.NewSettingsService(repo)
	s := models.Default()
	s.Opacity = 0.1

	if err := svc.SaveSettings(s); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.GetSettings()
	if got.Opacity != 0.4 {
		t.Errorf("opacity: got %f, want 0.4 (clamped from 0.1)", got.Opacity)
	}
}

func TestSettingsService_SaveSettings_ClampsOpacityHigh(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := service.NewSettingsService(repo)
	s := models.Default()
	s.Opacity = 1.5

	if err := svc.SaveSettings(s); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.GetSettings()
	if got.Opacity != 1.0 {
		t.Errorf("opacity: got %f, want 1.0 (clamped from 1.5)", got.Opacity)
	}
}

func TestSettingsService_SaveSettings_UpdatesCache(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := service.NewSettingsService(repo)

	// Prime cache
	svc.GetSettings() //nolint:errcheck

	s := models.Default()
	s.Theme = "dark"
	svc.SaveSettings(s) //nolint:errcheck

	got, _ := svc.GetSettings()
	if got.Theme != "dark" {
		t.Errorf("theme: got %q, want %q", got.Theme, "dark")
	}

	// Load should still have been called only once (cache primed, then updated by Save)
	repo.mu.Lock()
	count := repo.loadCount
	repo.mu.Unlock()
	if count != 1 {
		t.Errorf("Load called %d times, want 1", count)
	}
}

func TestSettingsService_ResetWindowPosition(t *testing.T) {
	repo := newMockSettingsRepo()
	svc := service.NewSettingsService(repo)

	s := models.Default()
	s.WindowX = 200
	s.WindowY = 300
	svc.SaveSettings(s) //nolint:errcheck

	if err := svc.ResetWindowPosition(); err != nil {
		t.Fatal(err)
	}

	got, _ := svc.GetSettings()
	if got.WindowX != -1 || got.WindowY != -1 {
		t.Errorf("window pos after reset: got (%d,%d), want (-1,-1)", got.WindowX, got.WindowY)
	}
}

func TestSettingsService_ConcurrentReads(t *testing.T) {
	svc := service.NewSettingsService(newMockSettingsRepo())

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svc.GetSettings() //nolint:errcheck
		}()
	}
	wg.Wait()
}

func TestSettingsService_ConcurrentReadWrite(t *testing.T) {
	svc := service.NewSettingsService(newMockSettingsRepo())

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			svc.GetSettings() //nolint:errcheck
		}()
		go func() {
			defer wg.Done()
			s := models.Default()
			s.Theme = "dark"
			svc.SaveSettings(s) //nolint:errcheck
		}()
	}
	wg.Wait()
}
