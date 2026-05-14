package window

import (
	"context"
	"sync"
	"time"

	"github.com/bep/debounce"
	"github.com/iqbaleff214/todo-app/internal/service"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	widgetW     = 280
	widgetH     = 420
	expandedMinW = 420
	expandedMinH = 500
)

type Manager struct {
	ctx             context.Context
	settingsService *service.SettingsService
	mu              sync.Mutex
	expanded        bool
	visible         bool
	debounceSave    func(f func())
}

func NewManager(ctx context.Context, svc *service.SettingsService) *Manager {
	return &Manager{
		ctx:             ctx,
		settingsService: svc,
		visible:         true,
		debounceSave:    debounce.New(500 * time.Millisecond),
	}
}

// SetWidgetMode switches the window to compact floating widget mode.
func (m *Manager) SetWidgetMode() {
	m.mu.Lock()
	m.expanded = false
	m.mu.Unlock()

	runtime.WindowSetAlwaysOnTop(m.ctx, true)
	// Lock size to widget dimensions.
	runtime.WindowSetMinSize(m.ctx, widgetW, widgetH)
	runtime.WindowSetMaxSize(m.ctx, widgetW, widgetH)
	runtime.WindowSetSize(m.ctx, widgetW, widgetH)
	m.restorePosition()
}

// SetExpandedMode switches to the full expanded window.
func (m *Manager) SetExpandedMode() {
	m.mu.Lock()
	m.expanded = true
	m.mu.Unlock()

	runtime.WindowSetAlwaysOnTop(m.ctx, false)
	// Remove max-size lock so the user can resize freely.
	runtime.WindowSetMaxSize(m.ctx, 0, 0)
	runtime.WindowSetMinSize(m.ctx, expandedMinW, expandedMinH)
	runtime.WindowSetSize(m.ctx, 720, 540)
}

// ToggleExpanded flips between widget and expanded mode.
func (m *Manager) ToggleExpanded() {
	m.mu.Lock()
	expanded := m.expanded
	m.mu.Unlock()

	if expanded {
		m.SetWidgetMode()
	} else {
		m.SetExpandedMode()
	}
}

// ToggleVisibility shows or hides the window.
func (m *Manager) ToggleVisibility() {
	m.mu.Lock()
	visible := m.visible
	m.mu.Unlock()

	if visible {
		runtime.WindowHide(m.ctx)
		m.mu.Lock()
		m.visible = false
		m.mu.Unlock()
	} else {
		runtime.WindowShow(m.ctx)
		m.mu.Lock()
		m.visible = true
		m.mu.Unlock()
	}
}

// IsVisible returns whether the window is currently shown.
func (m *Manager) IsVisible() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.visible
}

// IsExpanded returns whether the window is in expanded mode.
func (m *Manager) IsExpanded() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.expanded
}

// SavePosition debounces writes of the current window position to settings.
// Call this on every window-move event to avoid thrashing on drag.
func (m *Manager) SavePosition() {
	m.debounceSave(func() {
		x, y := runtime.WindowGetPosition(m.ctx)
		settings, err := m.settingsService.GetSettings()
		if err != nil {
			return
		}
		settings.WindowX = x
		settings.WindowY = y
		m.settingsService.SaveSettings(settings) //nolint:errcheck
	})
}

// restorePosition sets the window position from settings, computing a
// center-right default on first launch (WindowX == -1).
func (m *Manager) restorePosition() {
	settings, err := m.settingsService.GetSettings()
	if err != nil {
		return
	}

	x, y := settings.WindowX, settings.WindowY
	if x == -1 || y == -1 {
		x, y = m.defaultPosition()
	} else {
		x, y = m.clampToScreen(x, y)
	}
	runtime.WindowSetPosition(m.ctx, x, y)
}

// defaultPosition returns center-right on the primary screen.
func (m *Manager) defaultPosition() (int, int) {
	screens, err := runtime.ScreenGetAll(m.ctx)
	if err != nil || len(screens) == 0 {
		return 40, 100
	}

	sw, sh := 1920, 1080
	for _, s := range screens {
		if s.IsPrimary {
			sw = s.Size.Width
			sh = s.Size.Height
			break
		}
	}
	if sw == 0 {
		sw = screens[0].Size.Width
		sh = screens[0].Size.Height
	}

	x := sw - widgetW - 40
	y := (sh - widgetH) / 2
	return x, y
}

// clampToScreen ensures the widget stays within the combined screen bounds.
// Because the Wails Screen struct only exposes logical size (not origin),
// we use the largest reported screen dimensions as the bounding rectangle.
func (m *Manager) clampToScreen(x, y int) (int, int) {
	screens, err := runtime.ScreenGetAll(m.ctx)
	if err != nil || len(screens) == 0 {
		return x, y
	}

	maxW, maxH := 0, 0
	for _, s := range screens {
		if s.Size.Width > maxW {
			maxW = s.Size.Width
		}
		if s.Size.Height > maxH {
			maxH = s.Size.Height
		}
	}

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if maxW > 0 && x+widgetW > maxW {
		x = maxW - widgetW
	}
	if maxH > 0 && y+widgetH > maxH {
		y = maxH - widgetH
	}
	return x, y
}
