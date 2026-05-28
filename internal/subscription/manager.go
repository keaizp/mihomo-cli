package subscription

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"mihomo-cli/internal/cfg"
)

// SubscriptionConfig represents a parsed Clash subscription config.
type SubscriptionConfig struct {
	Proxies     []map[string]any `yaml:"proxies"`
	ProxyGroups []map[string]any `yaml:"proxy-groups"`
	Rules       []string         `yaml:"rules"`
}

// Manager handles fetching and parsing subscriptions.
type Manager struct {
	cfg *cfg.Manager
}

// NewManager creates a subscription manager.
func NewManager(cfgMgr *cfg.Manager) *Manager {
	return &Manager{cfg: cfgMgr}
}

// Fetch downloads and decodes a subscription from the given URL.
func (m *Manager) Fetch(subURL string) (*SubscriptionConfig, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(subURL)
	if err != nil {
		return nil, fmt.Errorf("fetch subscription: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	// Try base64 decode first (standard Clash subscription format)
	decoded, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		// Not base64, try raw YAML
		decoded = body
	}

	var subCfg SubscriptionConfig
	if err := yaml.Unmarshal(decoded, &subCfg); err != nil {
		return nil, fmt.Errorf("parse subscription YAML: %w", err)
	}

	return &subCfg, nil
}

// UpdateSubscription fetches a subscription by name, saves its profile, and
// updates its timestamp.
func (m *Manager) UpdateSubscription(name string) error {
	appCfg := m.cfg.Config()
	var sub *cfg.Subscription
	for i, s := range appCfg.Subscriptions {
		if s.Name == name {
			sub = &appCfg.Subscriptions[i]
			break
		}
	}
	if sub == nil {
		return fmt.Errorf("subscription %q not found", name)
	}

	subCfg, err := m.Fetch(sub.URL)
	if err != nil {
		return fmt.Errorf("fetch %q: %w", name, err)
	}

	// Save the fetched profile to disk for later merging.
	profilesDir := filepath.Join(m.cfg.ConfigDir(), "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return fmt.Errorf("create profiles dir: %w", err)
	}

	profilePath := filepath.Join(profilesDir, name+".yaml")
	data, err := yaml.Marshal(subCfg)
	if err != nil {
		return fmt.Errorf("marshal profile: %w", err)
	}
	if err := os.WriteFile(profilePath, data, 0644); err != nil {
		return fmt.Errorf("save profile: %w", err)
	}

	if err := m.cfg.UpdateSubscriptionTimestamp(name, time.Now().Unix()); err != nil {
		return fmt.Errorf("update timestamp: %w", err)
	}

	return m.MergeAndGenerate()
}

// UpdateAll updates all subscriptions.
func (m *Manager) UpdateAll() []error {
	var errs []error
	appCfg := m.cfg.Config()
	for _, s := range appCfg.Subscriptions {
		if err := m.UpdateSubscription(s.Name); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// MergeAndGenerate merges all subscription configs with user overrides and
// writes the final mihomo config file.
func (m *Manager) MergeAndGenerate() error {
	appCfg := m.cfg.Config()
	allProxies := make([]map[string]any, 0)

	profilesDir := filepath.Join(m.cfg.ConfigDir(), "profiles")
	for _, sub := range appCfg.Subscriptions {
		profilePath := filepath.Join(profilesDir, sub.Name+".yaml")
		data, err := os.ReadFile(profilePath)
		if err != nil {
			continue
		}
		var sc SubscriptionConfig
		if err := yaml.Unmarshal(data, &sc); err != nil {
			continue
		}
		allProxies = append(allProxies, sc.Proxies...)
	}

	allProxies = append(allProxies, appCfg.UserProxies...)

	mihomoCfg := map[string]any{
		"port":                appCfg.Core.HTTPPort,
		"socks-port":          appCfg.Core.SOCKSPort,
		"mixed-port":          appCfg.Core.MixedPort,
		"allow-lan":           appCfg.Core.AllowLAN,
		"mode":                appCfg.Mode,
		"log-level":           appCfg.Core.LogLevel,
		"external-controller": fmt.Sprintf("127.0.0.1:%d", appCfg.Core.APIPort),
		"proxies":             allProxies,
	}

	data, err := yaml.Marshal(&mihomoCfg)
	if err != nil {
		return fmt.Errorf("marshal mihomo config: %w", err)
	}

	mihomoConfigPath := m.cfg.MihomoConfigPath()
	if err := os.MkdirAll(filepath.Dir(mihomoConfigPath), 0755); err != nil {
		return fmt.Errorf("create mihomo dir: %w", err)
	}
	return os.WriteFile(mihomoConfigPath, data, 0644)
}
