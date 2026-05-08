package models_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/iqbaleff214/todo-app/internal/models"
)

func TestTask_Validate_EmptyText(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			task := models.Task{Text: tc.text}
			err := task.Validate()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !errors.Is(err, models.ErrEmptyText) {
				t.Errorf("got %v, want ErrEmptyText", err)
			}
		})
	}
}

func TestTask_Validate_TextTooLong(t *testing.T) {
	task := models.Task{Text: strings.Repeat("x", 501)}
	err := task.Validate()
	if err == nil {
		t.Fatal("expected error for text > 500 chars, got nil")
	}
	if !errors.Is(err, models.ErrTextTooLong) {
		t.Errorf("got %v, want ErrTextTooLong", err)
	}
}

func TestTask_Validate_BoundaryLength(t *testing.T) {
	// exactly 500 chars must pass
	task := models.Task{Text: strings.Repeat("x", 500)}
	if err := task.Validate(); err != nil {
		t.Errorf("unexpected error for 500-char text: %v", err)
	}
}

func TestTask_Validate_ValidText(t *testing.T) {
	task := models.Task{Text: "Buy milk"}
	if err := task.Validate(); err != nil {
		t.Errorf("unexpected error for valid text: %v", err)
	}
}

func TestParseDate(t *testing.T) {
	valid := []string{
		"2024-01-15",
		"2000-12-31",
		"2024-02-29", // 2024 is a leap year
		"1999-01-01",
	}
	for _, s := range valid {
		t.Run("valid/"+s, func(t *testing.T) {
			got, err := models.ParseDate(s)
			if err != nil {
				t.Fatalf("ParseDate(%q) unexpected error: %v", s, err)
			}
			if got != s {
				t.Errorf("ParseDate(%q) = %q, want %q", s, got, s)
			}
		})
	}

	invalid := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no zero-padding month", "2024-1-15"},
		{"no zero-padding day", "2024-01-5"},
		{"short year", "24-01-15"},
		{"slash separator", "2024/01/15"},
		{"dot separator", "2024.01.15"},
		{"invalid month", "2024-13-01"},
		{"invalid day", "2024-01-32"},
		{"non-leap feb 29", "2023-02-29"},
		{"random string", "not-a-date"},
		{"datetime suffix", "2024-01-15T00:00:00"},
		{"reversed", "15-01-2024"},
	}
	for _, tc := range invalid {
		t.Run("invalid/"+tc.name, func(t *testing.T) {
			got, err := models.ParseDate(tc.input)
			if err == nil {
				t.Errorf("ParseDate(%q) = %q, expected error", tc.input, got)
			}
			if !errors.Is(err, models.ErrInvalidDate) {
				t.Errorf("ParseDate(%q) error = %v, want ErrInvalidDate", tc.input, err)
			}
		})
	}
}
