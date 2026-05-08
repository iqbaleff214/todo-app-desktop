package models_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/iqbaleff214/todo-app/internal/models"
)

func TestDefault(t *testing.T) {
	s := models.Default()

	if s.Theme != "system" {
		t.Errorf("Theme = %q, want %q", s.Theme, "system")
	}
	if s.Opacity != 0.9 {
		t.Errorf("Opacity = %v, want 0.9", s.Opacity)
	}
	if s.AutoHide != false {
		t.Error("AutoHide = true, want false")
	}
	if s.LaunchOnLogin != false {
		t.Error("LaunchOnLogin = true, want false")
	}
	if s.AllowEditPast != false {
		t.Error("AllowEditPast = true, want false")
	}
	if s.WindowX != -1 {
		t.Errorf("WindowX = %d, want -1", s.WindowX)
	}
	if s.WindowY != -1 {
		t.Errorf("WindowY = %d, want -1", s.WindowY)
	}
	if s.WidgetWidth != 280 {
		t.Errorf("WidgetWidth = %d, want 280", s.WidgetWidth)
	}
	if s.WidgetHeight != 420 {
		t.Errorf("WidgetHeight = %d, want 420", s.WidgetHeight)
	}
}

func TestSettings_ValidateTheme(t *testing.T) {
	valid := []string{"light", "dark", "system"}
	for _, theme := range valid {
		t.Run("valid/"+theme, func(t *testing.T) {
			s := models.Settings{Theme: theme}
			if err := s.ValidateTheme(); err != nil {
				t.Errorf("ValidateTheme() unexpected error: %v", err)
			}
		})
	}

	invalid := []string{"Light", "DARK", "auto", "none", "", "solarized"}
	for _, theme := range invalid {
		t.Run("invalid/"+theme, func(t *testing.T) {
			s := models.Settings{Theme: theme}
			err := s.ValidateTheme()
			if err == nil {
				t.Errorf("ValidateTheme() = nil, want error for theme %q", theme)
			}
			if !errors.Is(err, models.ErrInvalidTheme) {
				t.Errorf("ValidateTheme() error = %v, want ErrInvalidTheme", err)
			}
		})
	}
}

func TestSettings_ClampOpacity(t *testing.T) {
	cases := []struct {
		name  string
		input float64
		want  float64
	}{
		{"below minimum", 0.0, 0.4},
		{"negative", -1.5, 0.4},
		{"at minimum", 0.4, 0.4},
		{"mid range", 0.7, 0.7},
		{"default", 0.9, 0.9},
		{"at maximum", 1.0, 1.0},
		{"above maximum", 1.5, 1.0},
		{"way above maximum", 99.0, 1.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := models.Settings{Opacity: tc.input}
			s.ClampOpacity()
			if s.Opacity != tc.want {
				t.Errorf("ClampOpacity() with input %v = %v, want %v", tc.input, s.Opacity, tc.want)
			}
		})
	}
}

func TestSettings_JSONRoundTrip(t *testing.T) {
	original := models.Default()
	original.Theme = "dark"
	original.Opacity = 0.75

	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// verify expected camelCase key names are present
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("Unmarshal to map error: %v", err)
	}
	for _, key := range []string{"theme", "opacity", "autoHide", "launchOnLogin", "allowEditPast", "windowX", "windowY", "widgetWidth", "widgetHeight"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("JSON output missing key %q", key)
		}
	}

	// verify round-trip fidelity
	var got models.Settings
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal to Settings error: %v", err)
	}
	if got != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, original)
	}
}
