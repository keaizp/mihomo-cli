package cfg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManagerCreatesDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	m, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	cfg := m.Config()
	if cfg.Core.HTTPPort != 7890 {
		t.Errorf("expected HTTPPort 7890, got %d", cfg.Core.HTTPPort)
	}
	if cfg.Mode != "rule" {
		t.Errorf("expected mode rule, got %s", cfg.Mode)
	}

	configPath := filepath.Join(dir, "mihomo-cli", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
}

func TestAddAndRemoveSubscription(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	m, _ := NewManager()

	if err := m.AddSubscription("test", "https://example.com/sub"); err != nil {
		t.Fatalf("AddSubscription: %v", err)
	}
	cfg := m.Config()
	if len(cfg.Subscriptions) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(cfg.Subscriptions))
	}
	if cfg.Subscriptions[0].URL != "https://example.com/sub" {
		t.Errorf("wrong URL: %s", cfg.Subscriptions[0].URL)
	}

	if err := m.AddSubscription("test", "https://dup.com"); err == nil {
		t.Error("expected error for duplicate name")
	}

	if err := m.RemoveSubscription("test"); err != nil {
		t.Fatalf("RemoveSubscription: %v", err)
	}
	if len(m.Config().Subscriptions) != 0 {
		t.Error("expected 0 subscriptions after remove")
	}
}

func TestSetMode(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	m, _ := NewManager()

	if err := m.SetMode("invalid"); err == nil {
		t.Error("expected error for invalid mode")
	}
	if err := m.SetMode("global"); err != nil {
		t.Fatalf("SetMode global: %v", err)
	}
	if m.Config().Mode != "global" {
		t.Errorf("expected global, got %s", m.Config().Mode)
	}
}
