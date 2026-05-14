//go:build windows

package window

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`

// EnableAutostart adds the app to the Windows startup registry key.
func EnableAutostart(appName, execPath string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("autostart.Enable: open registry: %w", err)
	}
	defer k.Close()
	if err := k.SetStringValue(appName, execPath); err != nil {
		return fmt.Errorf("autostart.Enable: set value: %w", err)
	}
	return nil
}

// DisableAutostart removes the app from the Windows startup registry key.
func DisableAutostart(appName string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("autostart.Disable: open registry: %w", err)
	}
	defer k.Close()
	if err := k.DeleteValue(appName); err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("autostart.Disable: delete value: %w", err)
	}
	return nil
}

// IsAutostartEnabled reports whether the app is registered in the startup key.
func IsAutostartEnabled(appName string) (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false, fmt.Errorf("autostart.IsEnabled: open registry: %w", err)
	}
	defer k.Close()
	_, _, err = k.GetStringValue(appName)
	if err == nil {
		return true, nil
	}
	if err == registry.ErrNotExist {
		return false, nil
	}
	return false, fmt.Errorf("autostart.IsEnabled: %w", err)
}
