package models

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrEmptyText   = errors.New("task text must not be empty")
	ErrTextTooLong = errors.New("task text must not exceed 500 characters")
	ErrInvalidDate = errors.New("date must be in YYYY-MM-DD format")
)

type Task struct {
	ID        string
	Date      string // "YYYY-MM-DD"
	Text      string // max 500 chars
	Done      bool
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks that the task's Text field is non-empty and within the 500-char limit.
func (t Task) Validate() error {
	if strings.TrimSpace(t.Text) == "" {
		return ErrEmptyText
	}
	if len(t.Text) > 500 {
		return ErrTextTooLong
	}
	return nil
}

// ParseDate validates and returns s if it is a strictly formatted YYYY-MM-DD date.
func ParseDate(s string) (string, error) {
	if len(s) != 10 {
		return "", fmt.Errorf("%w: %q", ErrInvalidDate, s)
	}
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		return "", fmt.Errorf("%w: %q", ErrInvalidDate, s)
	}
	// round-trip guards against any value time.Parse normalises silently
	if parsed.Format("2006-01-02") != s {
		return "", fmt.Errorf("%w: %q", ErrInvalidDate, s)
	}
	return s, nil
}
