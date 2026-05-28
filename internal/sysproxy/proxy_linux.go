//go:build linux

package sysproxy

import (
	"fmt"
	"os/exec"
)

// Set configures the system HTTP/HTTPS proxy via gsettings (GNOME).
func (m *Manager) Set() error {
	host := m.httpHost
	port := fmt.Sprintf("%d", m.httpPort)

	commands := [][]string{
		{"gsettings", "set", "org.gnome.system.proxy", "mode", "manual"},
		{"gsettings", "set", "org.gnome.system.proxy.http", "host", host},
		{"gsettings", "set", "org.gnome.system.proxy.http", "port", port},
		{"gsettings", "set", "org.gnome.system.proxy.https", "host", host},
		{"gsettings", "set", "org.gnome.system.proxy.https", "port", port},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("gsettings: %s: %w", string(out), err)
		}
	}

	return nil
}

// Unset restores system proxy to "none".
func (m *Manager) Unset() error {
	cmd := exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gsettings unset: %s: %w", string(out), err)
	}
	return nil
}
