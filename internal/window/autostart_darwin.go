//go:build darwin

package window

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{.AppName}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.ExecPath}}</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<false/>
</dict>
</plist>
`

func plistPath(appName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("autostart: home dir: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", appName+".plist"), nil
}

// EnableAutostart registers the app as a login item via a LaunchAgent plist.
func EnableAutostart(appName, execPath string) error {
	path, err := plistPath(appName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("autostart.Enable: mkdir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("autostart.Enable: create plist: %w", err)
	}
	defer f.Close()

	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return fmt.Errorf("autostart.Enable: parse template: %w", err)
	}
	return tmpl.Execute(f, struct{ AppName, ExecPath string }{appName, execPath})
}

// DisableAutostart removes the LaunchAgent plist.
func DisableAutostart(appName string) error {
	path, err := plistPath(appName)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("autostart.Disable: %w", err)
	}
	return nil
}

// IsAutostartEnabled reports whether the LaunchAgent plist exists.
func IsAutostartEnabled(appName string) (bool, error) {
	path, err := plistPath(appName)
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
