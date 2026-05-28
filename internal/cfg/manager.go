package cfg

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AppConfig is the top-level mihomo-cli configuration.
type AppConfig struct {
	Core          CoreConfig       `yaml:"core"`
	Subscriptions []Subscription   `yaml:"subscriptions"`
	Mode          string           `yaml:"mode"` // rule, global, direct, script
	UserRules     []string         `yaml:"user_rules"`
	UserProxies   []map[string]any `yaml:"user_proxies"`
}

type CoreConfig struct {
	HTTPPort  int    `yaml:"http_port"`
	SOCKSPort int    `yaml:"socks_port"`
	MixedPort int    `yaml:"mixed_port"`
	APIPort   int    `yaml:"api_port"`
	AllowLAN  bool   `yaml:"allow_lan"`
	LogLevel  string `yaml:"log_level"`
}

type Subscription struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	Interval    int    `yaml:"interval"`
	LastUpdated int64  `yaml:"last_updated"`
}

// DefaultAppConfig returns sensible defaults.
func DefaultAppConfig() AppConfig {
	return AppConfig{
		Core: CoreConfig{
			HTTPPort:  7890,
			SOCKSPort: 7891,
			MixedPort: 7892,
			APIPort:   9090,
			AllowLAN:  false,
			LogLevel:  "info",
		},
		Mode: "rule",
	}
}

// Manager reads and writes application configuration.
type Manager struct {
	configDir string
	config    AppConfig
}

// NewManager creates a Manager, loading config from XDG config dir.
func NewManager() (*Manager, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home dir: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}
	configDir = filepath.Join(configDir, "mihomo-cli")

	m := &Manager{configDir: configDir, config: DefaultAppConfig()}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return m, m.Save()
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, &m.config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return m, nil
}

// Config returns a copy of the current app config.
func (m *Manager) Config() AppConfig { return m.config }

// Save writes the current config to disk.
func (m *Manager) Save() error {
	data, err := yaml.Marshal(&m.config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	configPath := filepath.Join(m.configDir, "config.yaml")
	return os.WriteFile(configPath, data, 0644)
}

// SetMode updates the proxy mode and saves.
func (m *Manager) SetMode(mode string) error {
	valid := map[string]bool{"rule": true, "global": true, "direct": true, "script": true}
	if !valid[mode] {
		return fmt.Errorf("invalid mode: %s, must be rule/global/direct/script", mode)
	}
	m.config.Mode = mode
	return m.Save()
}

// AddSubscription adds a subscription and saves.
func (m *Manager) AddSubscription(name, url string) error {
	for _, s := range m.config.Subscriptions {
		if s.Name == name {
			return fmt.Errorf("subscription %q already exists", name)
		}
	}
	m.config.Subscriptions = append(m.config.Subscriptions, Subscription{
		Name: name,
		URL:  url,
	})
	return m.Save()
}

// RemoveSubscription removes a subscription by name.
func (m *Manager) RemoveSubscription(name string) error {
	idx := -1
	for i, s := range m.config.Subscriptions {
		if s.Name == name {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("subscription %q not found", name)
	}
	m.config.Subscriptions = append(m.config.Subscriptions[:idx], m.config.Subscriptions[idx+1:]...)
	return m.Save()
}

// UpdateSubscriptionTimestamp sets the last_updated field for a subscription.
func (m *Manager) UpdateSubscriptionTimestamp(name string, ts int64) error {
	for i, s := range m.config.Subscriptions {
		if s.Name == name {
			m.config.Subscriptions[i].LastUpdated = ts
			return m.Save()
		}
	}
	return fmt.Errorf("subscription %q not found", name)
}

// ConfigDir returns the config directory path.
func (m *Manager) ConfigDir() string { return m.configDir }

// MihomoDir returns the mihomo working directory.
func (m *Manager) MihomoDir() string { return filepath.Join(m.configDir, "mihomo") }

// MihomoConfigPath returns the path to the generated mihomo config file.
func (m *Manager) MihomoConfigPath() string {
	return filepath.Join(m.MihomoDir(), "config.yaml")
}
