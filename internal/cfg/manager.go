package cfg

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AppConfig is the top-level mihomo-cli configuration.
type AppConfig struct {
	Core               CoreConfig       `yaml:"core"`
	ActiveSubscription string           `yaml:"active_subscription"`
	Subscriptions      []Subscription   `yaml:"subscriptions"`
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

// userConfigDir returns the config directory, detecting sudo via SUDO_USER.
func userConfigDir() (string, error) {
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		u, err := user.Lookup(sudoUser)
		if err != nil {
			return "", fmt.Errorf("lookup sudo user %q: %w", sudoUser, err)
		}
		return filepath.Join(u.HomeDir, ".config", "mihomo-cli"), nil
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "mihomo-cli"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".config", "mihomo-cli"), nil
}

// NewManager creates a Manager, loading config from XDG config dir.
func NewManager() (*Manager, error) {
	configDir, err := userConfigDir()
	if err != nil {
		return nil, err
	}

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

// UpdateSubscription updates a subscription's URL by name.
func (m *Manager) UpdateSubscription(name, url string) error {
	for i, s := range m.config.Subscriptions {
		if s.Name == name {
			m.config.Subscriptions[i].URL = url
			return m.Save()
		}
	}
	return fmt.Errorf("subscription %q not found", name)
}

// SetActiveSubscription sets the active subscription and saves.
// Pass empty string to deactivate (use no subscriptions).
func (m *Manager) SetActiveSubscription(name string) error {
	if name != "" {
		found := false
		for _, s := range m.config.Subscriptions {
			if s.Name == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("subscription %q not found", name)
		}
	}
	m.config.ActiveSubscription = name
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

// MihomoDir returns the mihomo working directory (kernel, logs, PID, generated config).
func (m *Manager) MihomoDir() string {
	if d := os.Getenv("MIHOMO_DATA_DIR"); d != "" {
		return d
	}
	return filepath.Join(m.configDir, "data")
}

// MihomoConfigPath returns the path to the generated mihomo config file.
func (m *Manager) MihomoConfigPath() string {
	return filepath.Join(m.MihomoDir(), "config.yaml")
}
