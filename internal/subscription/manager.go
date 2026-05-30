package subscription

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"mihomo-cli/internal/cfg"
)

// SubscriptionConfig represents a parsed Clash subscription config.
// Proxies can be either structured maps (Clash YAML) or URI strings (stored with __uri__ key).
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
	req, err := http.NewRequest("GET", subURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "clash-verge/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
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

	// Try raw body first (many servers return plain YAML / URI list).
	if subCfg, ok := parseSubscription(body); ok {
		return subCfg, nil
	}

	// If raw body wasn't valid, try base64 decoding (classic Clash subscription format).
	if decoded := decodeSubscriptionBody(body); decoded != nil {
		if subCfg, ok := parseSubscription(decoded); ok {
			return subCfg, nil
		}
	}

	return nil, fmt.Errorf("no proxies found in subscription (body starts with: %.80s)", string(body))
}

// parseSubscription tries to parse a byte slice as a subscription config.
func parseSubscription(data []byte) (*SubscriptionConfig, bool) {
	var subCfg SubscriptionConfig

	// Try structured Clash YAML: proxies as list of maps.
	if err := yaml.Unmarshal(data, &subCfg); err == nil && len(subCfg.Proxies) > 0 {
		return &subCfg, true
	}

	// Try "proxies" as list of URI strings (ss://..., vmess://...).
	var wrapper struct {
		Proxies []string `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err == nil && len(wrapper.Proxies) > 0 {
		for _, uri := range wrapper.Proxies {
			subCfg.Proxies = append(subCfg.Proxies, map[string]any{"__uri__": uri})
		}
		return &subCfg, true
	}

	// Try bare list of URI strings.
	var proxiesRaw []string
	if err := yaml.Unmarshal(data, &proxiesRaw); err == nil && len(proxiesRaw) > 0 {
		for _, uri := range proxiesRaw {
			subCfg.Proxies = append(subCfg.Proxies, map[string]any{"__uri__": uri})
		}
		return &subCfg, true
	}

	return nil, false
}

// decodeSubscriptionBody tries multiple base64 encodings.
// Returns nil if the body doesn't look like base64 at all.
func decodeSubscriptionBody(body []byte) []byte {
	raw := strings.TrimSpace(string(body))

	// Skip empty or obviously-not-base64 bodies.
	if len(raw) == 0 || raw[0] == '{' || raw[0] == '-' {
		return nil
	}

	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}
	for _, enc := range encodings {
		if decoded, err := enc.DecodeString(raw); err == nil && looksLikeSubscription(decoded) {
			return decoded
		}
	}
	return nil
}

// looksLikeSubscription checks if data appears to be a subscription config.
func looksLikeSubscription(data []byte) bool {
	s := strings.TrimSpace(string(data))
	return strings.Contains(s, "proxies") ||
		strings.Contains(s, "ss://") ||
		strings.Contains(s, "vmess://") ||
		strings.Contains(s, "trojan://")
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
	allProxies := make([]any, 0) // []any to hold both maps and URI strings

	profilesDir := filepath.Join(m.cfg.ConfigDir(), "profiles")
	for _, sub := range appCfg.Subscriptions {
		// If an active subscription is set, only merge that one.
		if appCfg.ActiveSubscription != "" && sub.Name != appCfg.ActiveSubscription {
			continue
		}
		profilePath := filepath.Join(profilesDir, sub.Name+".yaml")
		data, err := os.ReadFile(profilePath)
		if err != nil {
			continue
		}
		var sc SubscriptionConfig
		if err := yaml.Unmarshal(data, &sc); err != nil {
			continue
		}
		for _, p := range sc.Proxies {
			if uri, ok := p["__uri__"].(string); ok {
				// URI string proxy — pass through as raw string for mihomo to parse.
				allProxies = append(allProxies, uri)
			} else {
				allProxies = append(allProxies, p)
			}
		}
	}

	for _, p := range appCfg.UserProxies {
		allProxies = append(allProxies, p)
	}

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
