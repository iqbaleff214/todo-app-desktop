package window

import (
	_ "embed"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed tray-icon.png
var trayIcon []byte

// InitTray registers the system tray icon and menu.
// Uses systray.Register (non-blocking) so it integrates with the Wails
// webview event loop instead of starting a competing native loop.
func (m *Manager) InitTray() {
	systray.Register(m.onTrayReady, m.onTrayExit)
}

func (m *Manager) onTrayReady() {
	systray.SetIcon(trayIcon)
	systray.SetTooltip("todo-app")

	mShowHide := systray.AddMenuItem("Hide", "Show or hide the widget")
	mExpand := systray.AddMenuItem("Expand", "Open the expanded window")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit todo-app")

	for {
		select {
		case <-mShowHide.ClickedCh:
			m.ToggleVisibility()
			if m.IsVisible() {
				mShowHide.SetTitle("Hide")
			} else {
				mShowHide.SetTitle("Show")
			}

		case <-mExpand.ClickedCh:
			if !m.IsVisible() {
				runtime.WindowShow(m.ctx)
				m.mu.Lock()
				m.visible = true
				m.mu.Unlock()
				mShowHide.SetTitle("Hide")
			}
			m.SetExpandedMode()

		case <-mQuit.ClickedCh:
			systray.Quit()
			runtime.Quit(m.ctx)
			return
		}
	}
}

func (m *Manager) onTrayExit() {}
