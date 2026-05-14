//go:build linux

package window

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const desktopTemplate = `[Desktop Entry]
Type=Application
Version=1.0
Name={{.AppName}}
Exec={{.ExecPath}}
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
`

func desktopPath(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("autostart: home dir: %w", err)
	}
	return filepath.Join(home, ".config", "autostart", appName+".desktop"), nil
}

// EnableAutostart writes a .desktop file to ~/.config/autostart/.
func EnableAutostart(appName, execPath string) error {
	path, err := desktopPath(appName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("autostart.Enable: mkdir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("autostart.Enable: create desktop file: %w", err)
	}
	defer f.Close()

	tmpl, err := template.New("desktop").Parse(desktopTemplate)
	if err != nil {
		return fmt.Errorf("autostart.Enable: parse template: %w", err)
	}
	return tmpl.Execute(f, struct{ AppName, ExecPath string }{appName, execPath})
}

// DisableAutostart removes the .desktop autostart entry.
func DisableAutostart(appName string) error {
	path, err := desktopPath(appName)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("autostart.Disable: %w", err)
	}
	return nil
}

// IsAutostartEnabled reports whether the .desktop autostart entry exists.
func IsAutostartEnabled(appName string) (bool, error) {
	path, err := desktopPath(appName)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("autostart.IsEnabled: %w", err)
}
