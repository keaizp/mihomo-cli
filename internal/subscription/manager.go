package subscription

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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

// Fetch downloads and parses a subscription from the given URL using direct connection.
// Follows Clash Verge Rev's approach: no base64 decode, plain YAML expected.
func (m *Manager) Fetch(subURL string) (*SubscriptionConfig, error) {
	return m.fetch(subURL, nil)
}

// fetch performs the actual HTTP fetch and YAML parsing. If transport is nil,
// a direct-connection transport is used (no proxy).
func (m *Manager) fetch(subURL string, transport *http.Transport) (*SubscriptionConfig, error) {
	cleanURL, err := fixSubscriptionURL(subURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Strip credentials from the request URL — they're sent via Basic Auth header instead.
	// Matches Clash Verge Rev: network.rs get_with_tls_mode lines 193-206.
	reqURL := cleanURL
	if u, err := url.Parse(cleanURL); err == nil && u.User != nil {
		u.User = nil
		reqURL = u.String()
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "clash-verge/v1.0.0")

	// Set Basic Auth from the original URL (url package auto-decodes percent-encoding).
	if u, err := url.Parse(subURL); err == nil && u.User != nil {
		pass, _ := u.User.Password()
		req.SetBasicAuth(u.User.Username(), pass)
	}

	if transport == nil {
		transport = &http.Transport{Proxy: nil} // direct, no proxy
	}
	client := &http.Client{Timeout: 20 * time.Second, Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch subscription: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, fmt.Errorf("subscription returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	// Strip UTF-8 BOM if present (matching Clash Verge Rev: prfitem.rs line 385).
	if len(body) >= 3 && body[0] == 0xEF && body[1] == 0xBB && body[2] == 0xBF {
		body = body[3:]
	}
	data := string(body)

	// Parse as YAML (Clash Verge Rev does NOT base64-decode).
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(data), &root); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}
	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return nil, fmt.Errorf("expected YAML document, got kind %d", root.Kind)
	}
	doc := root.Content[0]

	// Verify the response contains proxies or proxy-providers (prfitem.rs lines 388-392).
	if doc.Kind == yaml.MappingNode {
		hasProxies := false
		for i := 0; i < len(doc.Content)-1; i += 2 {
			key := doc.Content[i]
			if key.Kind == yaml.ScalarNode && (key.Value == "proxies" || key.Value == "proxy-providers") {
				hasProxies = true
				break
			}
		}
		if !hasProxies {
			return nil, fmt.Errorf("subscription does not contain `proxies` or `proxy-providers` key")
		}
	} else if doc.Kind == yaml.SequenceNode {
		// Top-level sequence: treat each entry as a proxy URI string.
		// Deserialize into []any first to get raw entries.
		var rawProxies []any
		if err := yaml.Unmarshal([]byte(data), &rawProxies); err != nil {
			return nil, fmt.Errorf("parse proxy sequence: %w", err)
		}
		subCfg := &SubscriptionConfig{}
		for _, p := range rawProxies {
			switch v := p.(type) {
			case string:
				subCfg.Proxies = append(subCfg.Proxies, map[string]any{"__uri__": v})
			case map[string]any:
				subCfg.Proxies = append(subCfg.Proxies, v)
			}
		}
		if len(subCfg.Proxies) == 0 {
			return nil, fmt.Errorf("no proxies found in subscription")
		}
		return subCfg, nil
	} else {
		return nil, fmt.Errorf("unexpected YAML kind %d at document root", doc.Kind)
	}

	// Now parse into SubscriptionConfig.
	var subCfg SubscriptionConfig
	if err := yaml.Unmarshal([]byte(data), &subCfg); err != nil {
		return nil, fmt.Errorf("parse subscription: %w", err)
	}

	if len(subCfg.Proxies) == 0 {
		// Maybe proxies are URI strings — try that format.
		var wrapper struct {
			Proxies []string `yaml:"proxies"`
		}
		if err := yaml.Unmarshal([]byte(data), &wrapper); err != nil || len(wrapper.Proxies) == 0 {
			return nil, fmt.Errorf("no proxies found in subscription")
		}
		for _, uri := range wrapper.Proxies {
			subCfg.Proxies = append(subCfg.Proxies, map[string]any{"__uri__": uri})
		}
	}

	return &subCfg, nil
}

// fixSubscriptionURL handles malformed URLs where query params appear after & instead of ?.
// Matches Clash Verge Rev: prfitem.rs fix_dirty_url (lines 589-613).
// Example: https://example.com/path&param1=value1 → https://example.com/path?param1=value1
func fixSubscriptionURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}
	// If no query string but path contains '&', extract params from path.
	if u.RawQuery == "" && strings.Contains(u.Path, "&") {
		if idx := strings.Index(u.Path, "&"); idx >= 0 {
			params := u.Path[idx+1:]
			u.Path = u.Path[:idx]
			u.RawQuery = params
		}
	}
	return u.String(), nil
}

// UpdateSubscription fetches a subscription by name with three-tier proxy fallback,
// saves its profile, and updates its timestamp.
// Matches Clash Verge Rev: feat/profile.rs perform_profile_update (lines 100-187).
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

	// Three-tier proxy fallback matching Clash Verge Rev's perform_profile_update:
	// 1. Direct connection (no proxy)
	// 2. Through Clash localhost proxy (mixed port)
	// 3. Through system proxy (environment variables)
	subCfg, err := m.fetchWithFallback(sub.URL, appCfg.Core.MixedPort)
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

// fetchWithFallback tries three proxy methods in order, returning the first success.
// Matches Clash Verge Rev's perform_profile_update three-tier fallback.
func (m *Manager) fetchWithFallback(subURL string, mixedPort int) (*SubscriptionConfig, error) {
	// Tier 1: Direct connection.
	subCfg, err := m.fetch(subURL, nil)
	if err == nil {
		return subCfg, nil
	}
	fmt.Fprintf(os.Stderr, "[订阅] 直连失败: %v\n", err)

	// Tier 2: Through Clash localhost proxy (self_proxy).
	clashProxy := fmt.Sprintf("http://127.0.0.1:%d", mixedPort)
	if pu, err2 := url.Parse(clashProxy); err2 == nil {
		subCfg, err = m.fetch(subURL, &http.Transport{Proxy: http.ProxyURL(pu)})
		if err == nil {
			fmt.Fprintf(os.Stderr, "[订阅] Clash代理(127.0.0.1:%d) 成功\n", mixedPort)
			return subCfg, nil
		}
		fmt.Fprintf(os.Stderr, "[订阅] Clash代理(127.0.0.1:%d) 失败: %v\n", mixedPort, err)
	}

	// Tier 3: Through system proxy (with_proxy, via environment variables).
	subCfg, err = m.fetch(subURL, &http.Transport{Proxy: http.ProxyFromEnvironment})
	if err == nil {
		fmt.Fprintf(os.Stderr, "[订阅] 系统代理 成功\n")
		return subCfg, nil
	}
	fmt.Fprintf(os.Stderr, "[订阅] 系统代理 失败: %v\n", err)

	return nil, fmt.Errorf("all methods failed")
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
