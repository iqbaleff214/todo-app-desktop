package repository_test

import (
	"sync"
	"testing"

	"github.com/iqbaleff214/todo-app/internal/models"
	"github.com/iqbaleff214/todo-app/internal/repository"
)

func newTestSettingsRepo(t *testing.T) repository.SettingsRepository {
	t.Helper()
	return repository.NewSettingsRepository(t.TempDir())
}

// --- Load ---

func TestSettingsRepo_Load_MissingFileReturnsDefaults(t *testing.T) {
	repo := newTestSettingsRepo(t)

	got, err := repo.Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	want := models.Default()
	if got != want {
		t.Errorf("Load() = %+v, want defaults %+v", got, want)
	}
}

func TestSettingsRepo_Load_DefaultsMatchModel(t *testing.T) {
	repo := newTestSettingsRepo(t)
	got, _ := repo.Load()

	if got.Theme != "system" {
		t.Errorf("Theme = %q, want %q", got.Theme, "system")
	}
	if got.Opacity != 0.9 {
		t.Errorf("Opacity = %v, want 0.9", got.Opacity)
	}
	if got.WindowX != -1 {
		t.Errorf("WindowX = %d, want -1", got.WindowX)
	}
	if got.WindowY != -1 {
		t.Errorf("WindowY = %d, want -1", got.WindowY)
	}
	if got.WidgetWidth != 280 {
		t.Errorf("WidgetWidth = %d, want 280", got.WidgetWidth)
	}
	if got.WidgetHeight != 420 {
		t.Errorf("WidgetHeight = %d, want 420", got.WidgetHeight)
	}
}

// --- Save + Load round-trip ---

func TestSettingsRepo_RoundTrip(t *testing.T) {
	repo := newTestSettingsRepo(t)

	want := models.Settings{
		Theme:         "dark",
		Opacity:       0.6,
		AutoHide:      true,
		LaunchOnLogin: true,
		AllowEditPast: true,
		WindowX:       120,
		WindowY:       240,
		WidgetWidth:   300,
		WidgetHeight:  500,
	}

	if err := repo.Save(want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := repo.Load()
	if err != nil {
		t.Fatalf("Load() after Save() error: %v", err)
	}

	if got != want {
		t.Errorf("round-trip mismatch:\ngot  %+v\nwant %+v", got, want)
	}
}

func TestSettingsRepo_RoundTrip_AllThemes(t *testing.T) {
	for _, theme := range []string{"light", "dark", "system"} {
		t.Run(theme, func(t *testing.T) {
			repo := newTestSettingsRepo(t)
			s := models.Default()
			s.Theme = theme

			if err := repo.Save(s); err != nil {
				t.Fatalf("Save: %v", err)
			}
			got, err := repo.Load()
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got.Theme != theme {
				t.Errorf("Theme = %q, want %q", got.Theme, theme)
			}
		})
	}
}

func TestSettingsRepo_Save_OverwritesPreviousValue(t *testing.T) {
	repo := newTestSettingsRepo(t)

	first := models.Default()
	first.Theme = "light"
	if err := repo.Save(first); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	second := models.Default()
	second.Theme = "dark"
	if err := repo.Save(second); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	got, err := repo.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Theme != "dark" {
		t.Errorf("Theme = %q, want %q", got.Theme, "dark")
	}
}

// --- Concurrent saves ---

func TestSettingsRepo_ConcurrentSaves_NoCorruption(t *testing.T) {
	repo := newTestSettingsRepo(t)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			s := models.Default()
			s.WindowX = i
			// Ignore errors: a concurrent rename race on Windows may transiently
			// fail, but the final file must always be valid JSON.
			_ = repo.Save(s)
		}(i)
	}
	wg.Wait()

	// After all concurrent writes, Load must succeed and return valid data.
	got, err := repo.Load()
	if err != nil {
		t.Fatalf("Load() after concurrent saves error: %v", err)
	}
	// Theme must still be one of the allowed values — not garbled.
	if err := got.ValidateTheme(); err != nil {
		t.Errorf("theme corrupted after concurrent saves: %v", err)
	}
}
