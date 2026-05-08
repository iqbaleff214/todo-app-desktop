package models

import (
	"errors"
	"fmt"
)

var ErrInvalidTheme = errors.New("theme must be one of: light, dark, system")

type Settings struct {
	Theme         string  `json:"theme"`
	Opacity       float64 `json:"opacity"`
	AutoHide      bool    `json:"autoHide"`
	LaunchOnLogin bool    `json:"launchOnLogin"`
	AllowEditPast bool    `json:"allowEditPast"`
	WindowX       int     `json:"windowX"`
	WindowY       int     `json:"windowY"`
	WidgetWidth   int     `json:"widgetWidth"`
	WidgetHeight  int     `json:"widgetHeight"`
}

func Default() Settings {
	return Settings{
		Theme:         "system",
		Opacity:       0.9,
		AutoHide:      false,
		LaunchOnLogin: false,
		AllowEditPast: false,
		WindowX:       -1, // -1 means "use default position"
		WindowY:       -1,
		WidgetWidth:   280,
		WidgetHeight:  420,
	}
}

// ValidateTheme returns an error if Theme is not one of the allowed values.
func (s Settings) ValidateTheme() error {
	switch s.Theme {
	case "light", "dark", "system":
		return nil
	default:
		return fmt.Errorf("%w: %q", ErrInvalidTheme, s.Theme)
	}
}

// ClampOpacity enforces the [0.4, 1.0] opacity range in place.
func (s *Settings) ClampOpacity() {
	if s.Opacity < 0.4 {
		s.Opacity = 0.4
	}
	if s.Opacity > 1.0 {
		s.Opacity = 1.0
	}
}
