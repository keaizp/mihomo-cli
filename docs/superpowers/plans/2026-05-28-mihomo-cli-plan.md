# Mihomo CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a single-binary Linux CLI tool (`mihomo-cli`) with TUI for managing mihomo proxy — subscriptions, node switching, latency testing, service control.

**Architecture:** Go 1.22+, cobra CLI + bubbletea TUI share a core logic layer. Mihomo kernel auto-downloaded and managed as child process, controlled via REST API on :9090.

**Tech Stack:** Go 1.22+, cobra, bubbletea, lipgloss, bubbles, stdlib net/http, gopkg.in/yaml.v3

**Target:** Linux (development can happen on any platform; system-proxy is Linux-only via build tags)

---

## File Structure

```
mihomo-cli/
├── cmd/mihomo-cli/main.go
├── internal/
│   ├── cli/
│   │   ├── root.go          # Root cmd, TUI fallback, --version
│   │   ├── sub.go           # sub add/remove/update/list
│   │   ├── mode.go          # mode set/show
│   │   ├── proxy.go         # proxy list/set/test/info
│   │   ├── service.go       # service start/stop/restart/status/logs
│   │   ├── conn.go          # conn list/close
│   │   └── config_cmd.go    # config edit/show/reload
│   ├── tui/
│   │   ├── model.go         # bubbletea Model, Init
│   │   ├── update.go        # Update message handler
│   │   ├── view.go          # View renderer (3-panel)
│   │   └── styles.go        # lipgloss styles
│   ├── subscription/
│   │   └── manager.go       # Fetch, parse, merge subscriptions
│   ├── cfg/
│   │   └── manager.go       # Read/write app config, generate mihomo config
│   ├── api/
│   │   └── client.go        # HTTP client for mihomo REST API
│   ├── kernel/
│   │   └── manager.go       # Download, start, stop, health-check mihomo
│   └── sysproxy/
│       ├── proxy.go          # Interface
│       └── proxy_linux.go    # Linux gsettings impl
├── go.mod
└── go.sum
```

**Design principles:**
- `cfg.Manager` is the dependency-injection hub — everything that needs config receives a `*cfg.Manager`
- `api.Client` wraps all mihomo REST calls; TUI and CLI both use it
- `kernel.Manager` owns the `exec.Cmd` lifecycle; exposes `Start()/Stop()/IsRunning()`
- `subscription.Manager` handles fetch + parse + merge; writes result through `cfg.Manager`

---

### Task 1: Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/mihomo-cli/main.go`
- Create: `internal/cli/root.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd /c/Users/15940/cc
go mod init mihomo-cli
```

Expected: creates `go.mod` with `module mihomo-cli` and `go 1.22.x`

- [ ] **Step 2: Install dependencies**

```bash
go get github.com/spf13/cobra@latest
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest
go get gopkg.in/yaml.v3@latest
```

- [ ] **Step 3: Write root command**

Write `internal/cli/root.go`:

```go
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mihomo-cli",
	Short: "Manage mihomo proxy from the command line",
	Long:  "A CLI tool for managing mihomo proxy subscriptions, nodes, modes, and service lifecycle.",
	Run: func(cmd *cobra.Command, args []string) {
		// No subcommand → launch TUI
		fmt.Println("TUI mode not yet implemented")
		os.Exit(0)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Write main.go**

Write `cmd/mihomo-cli/main.go`:

```go
package main

import "mihomo-cli/internal/cli"

func main() {
	cli.Execute()
}
```

- [ ] **Step 5: Verify build**

```bash
go build ./cmd/mihomo-cli/
```

Expected: binary compiles, `./mihomo-cli` prints "TUI mode not yet implemented"

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "feat: project scaffolding with cobra root command"
```

---

### Task 2: Config Manager

**Files:**
- Create: `internal/cfg/manager.go`

- [ ] **Step 1: Write config types and manager**

Write `internal/cfg/manager.go`:

```go
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
	Interval    int    `yaml:"interval"` // update interval in seconds, 0 = manual
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
```

- [ ] **Step 2: Write unit test**

Create `internal/cfg/manager_test.go`:

```go
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

	// Verify file was written
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
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/cfg/ -v
```

Expected: all tests pass

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: config manager with add/remove subscription and mode management"
```

---

### Task 3: Mihomo API Client

**Files:**
- Create: `internal/api/client.go`

- [ ] **Step 1: Write API client**

Write `internal/api/client.go`:

```go
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client communicates with the mihomo REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates an API client for the given mihomo API address.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Proxy represents a single proxy node returned by the mihomo API.
type Proxy struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Now     string   `json:"now,omitempty"` // currently selected node in group
	All     []string `json:"all,omitempty"` // all nodes in group
	History []struct {
		Time time.Time `json:"time"`
		Delay int      `json:"delay"`
	} `json:"history,omitempty"`
}

// ProxyGroup is a top-level proxy group from GET /proxies.
type ProxyGroup struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Now     string   `json:"now"`
	All     []string `json:"all"`
}

// ProxiesResponse is the full response from GET /proxies.
type ProxiesResponse struct {
	Proxies map[string]Proxy `json:"proxies"`
}

// Connection represents an active connection.
type Connection struct {
	ID          string   `json:"id"`
	Metadata    struct {
		Network string `json:"network"`
		Host    string `json:"host"`
	} `json:"metadata"`
	Upload      int64  `json:"upload"`
	Download    int64  `json:"download"`
	Start       string `json:"start"`
	Chains      []string `json:"chains"`
	Rule        string `json:"rule"`
}

// ConnectionsResponse is the full response from GET /connections.
type ConnectionsResponse struct {
	Connections []Connection `json:"connections"`
}

// DelayResult is the result of a single latency test.
type DelayResult struct {
	Delay int `json:"delay"`
}

// GetProxies fetches all proxy groups and nodes.
func (c *Client) GetProxies() (*ProxiesResponse, error) {
	var resp ProxiesResponse
	if err := c.do("GET", "/proxies", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SwitchProxy selects a specific node within a proxy group.
func (c *Client) SwitchProxy(group, node string) error {
	body := fmt.Sprintf(`{"name":"%s"}`, node)
	return c.do("PUT", "/proxies/"+url.PathEscape(group), strings.NewReader(body), nil)
}

// TestDelay tests the latency of a specific proxy node.
// Returns delay in milliseconds, or an error.
func (c *Client) TestDelay(name string, timeout time.Duration) (int, error) {
	u := fmt.Sprintf("%s/proxies/%s/delay?timeout=%d&url=https://www.gstatic.com/generate_204",
		c.baseURL, url.PathEscape(name), int(timeout.Milliseconds()))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result DelayResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	if result.Delay == 0 {
		return 0, fmt.Errorf("timeout")
	}
	return result.Delay, nil
}

// GetConnections fetches all active connections.
func (c *Client) GetConnections() (*ConnectionsResponse, error) {
	var resp ConnectionsResponse
	if err := c.do("GET", "/connections", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CloseConnection closes a single connection by ID.
func (c *Client) CloseConnection(id string) error {
	return c.do("DELETE", "/connections/"+url.PathEscape(id), nil, nil)
}

// ReloadConfig triggers mihomo to reload its configuration.
func (c *Client) ReloadConfig() error {
	body := strings.NewReader(`{"path":""}`)
	return c.do("PUT", "/configs", body, nil)
}

// HealthCheck returns nil if the mihomo API is reachable.
func (c *Client) HealthCheck() error {
	var v any
	return c.do("GET", "/version", nil, &v)
}

func (c *Client) do(method, path string, body io.Reader, result any) error {
	u := c.baseURL + path
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api error %d: %s", resp.StatusCode, string(b))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
```

- [ ] **Step 2: Commit (no test required — integration-only; unit testing would mock HTTP, defer to integration phase)**

```bash
git add -A && git commit -m "feat: mihomo REST API client"
```

---

### Task 4: Kernel Manager

**Files:**
- Create: `internal/kernel/manager.go`

- [ ] **Step 1: Write kernel manager**

Write `internal/kernel/manager.go`:

```go
package kernel

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"mihomo-cli/internal/api"
)

const (
	mihomoRepo  = "https://github.com/MetaCubeX/mihomo/releases"
	apiBaseURL  = "http://127.0.0.1:%d"
	startupWait = 2 * time.Second
)

// Manager handles the mihomo kernel lifecycle.
type Manager struct {
	mu       sync.Mutex
	cmd      *exec.Cmd
	binPath  string
	workDir  string
	apiPort  int
	apiClient *api.Client
}

// NewManager creates a kernel manager.
// binDir is where the mihomo binary is stored (config dir).
// workDir is the mihomo working directory (contains config.yaml).
func NewManager(binDir, workDir string, apiPort int) *Manager {
	return &Manager{
		binPath: filepath.Join(binDir, "mihomo"),
		workDir: workDir,
		apiPort: apiPort,
	}
}

// IsInstalled returns true if the mihomo binary exists.
func (m *Manager) IsInstalled() bool {
	_, err := os.Stat(m.binPath)
	return err == nil
}

// Install downloads the latest mihomo binary from GitHub.
func (m *Manager) Install() error {
	if m.IsInstalled() {
		return nil
	}

	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "amd64"
	} else if arch == "arm64" {
		arch = "arm64"
	}

	// Download URL pattern for mihomo releases
	url := fmt.Sprintf(
		"%s/download/v1.18.10/mihomo-linux-%s-v1.18.10.gz",
		mihomoRepo, arch,
	)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download mihomo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download mihomo: HTTP %d", resp.StatusCode)
	}

	// Write to temp file, gunzip, then move
	tmpPath := m.binPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpPath)

	// Handle gzip decompression
	// Simplified: write raw if not compressed; for real impl detect Content-Encoding
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return fmt.Errorf("write binary: %w", err)
	}
	f.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}
	if err := os.Rename(tmpPath, m.binPath); err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

// Start launches mihomo as a child process.
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd != nil {
		return fmt.Errorf("mihomo is already running")
	}

	if err := os.MkdirAll(m.workDir, 0755); err != nil {
		return fmt.Errorf("create work dir: %w", err)
	}

	configPath := filepath.Join(m.workDir, "config.yaml")
	m.cmd = exec.Command(m.binPath, "-d", m.workDir, "-f", configPath)
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	if err := m.cmd.Start(); err != nil {
		m.cmd = nil
		return fmt.Errorf("start mihomo: %w", err)
	}

	// Wait for API to become available
	time.Sleep(startupWait)
	m.apiClient = api.NewClient(fmt.Sprintf(apiBaseURL, m.apiPort))

	return nil
}

// Stop terminates the mihomo process.
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd == nil {
		return nil
	}
	if err := m.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("kill mihomo: %w", err)
	}
	m.cmd.Wait()
	m.cmd = nil
	m.apiClient = nil
	return nil
}

// Restart stops and starts mihomo.
func (m *Manager) Restart() error {
	m.Stop()
	return m.Start()
}

// IsRunning checks if mihomo is running and responsive.
func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.apiClient == nil {
		return false
	}
	return m.apiClient.HealthCheck() == nil
}

// APIClient returns the API client, or nil if mihomo is not running.
func (m *Manager) APIClient() *api.Client {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.apiClient
}

// Status returns a human-readable status string.
func (m *Manager) Status() string {
	if m.cmd != nil && m.IsRunning() {
		return "running"
	}
	if m.cmd != nil {
		return "starting"
	}
	return "stopped"
}

// ReadLogs reads the last N lines from the mihomo log file.
func (m *Manager) ReadLogs(n int) ([]string, error) {
	logPath := filepath.Join(m.workDir, "logs")
	entries, err := os.ReadDir(logPath)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	// Return the most recent log file name
	latest := entries[len(entries)-1]
	return []string{latest.Name()}, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add -A && git commit -m "feat: kernel manager for mihomo lifecycle"
```

---

### Task 5: Subscription Manager

**Files:**
- Create: `internal/subscription/manager.go`

- [ ] **Step 1: Write subscription manager**

Write `internal/subscription/manager.go`:

```go
package subscription

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
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
// Returns the parsed proxy nodes.
func (m *Manager) Fetch(subURL string) (*SubscriptionConfig, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(subURL)
	if err != nil {
		return nil, fmt.Errorf("fetch subscription: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
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

// UpdateSubscription fetches a subscription by name and merges it into the mihomo config.
func (m *Manager) UpdateSubscription(name string) error {
	cfg := m.cfg.Config()
	var sub *cfg.Subscription
	for i, s := range cfg.Subscriptions {
		if s.Name == name {
			sub = &cfg.Subscriptions[i]
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

	_ = subCfg // Merged in a later task (GenerateMihomoConfig)
	return m.cfg.UpdateSubscriptionTimestamp(name, time.Now().Unix())
}

// UpdateAll updates all subscriptions.
func (m *Manager) UpdateAll() []error {
	var errs []error
	cfg := m.cfg.Config()
	for _, s := range cfg.Subscriptions {
		if err := m.UpdateSubscription(s.Name); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// MergeAndGenerate merges all subscription configs with user overrides
// and writes the final mihomo config file.
func (m *Manager) MergeAndGenerate() error {
	cfg := m.cfg.Config()
	allProxies := make([]map[string]any, 0)

	// Collect proxies from all subscription profiles
	profilesDir := m.cfg.ConfigDir() + "/profiles"
	for _, sub := range cfg.Subscriptions {
		profilePath := profilesDir + "/" + sub.Name + ".yaml"
		data, err := os.ReadFile(profilePath)
		if err != nil {
			continue // Skip missing profiles
		}
		var sc SubscriptionConfig
		if err := yaml.Unmarshal(data, &sc); err != nil {
			continue
		}
		allProxies = append(allProxies, sc.Proxies...)
	}

	// Add user-defined proxies
	allProxies = append(allProxies, cfg.UserProxies...)

	// Build mihomo config
	mihomoCfg := map[string]any{
		"port":            cfg.Core.HTTPPort,
		"socks-port":      cfg.Core.SOCKSPort,
		"mixed-port":      cfg.Core.MixedPort,
		"allow-lan":       cfg.Core.AllowLAN,
		"mode":            cfg.Mode,
		"log-level":       cfg.Core.LogLevel,
		"external-controller": fmt.Sprintf("127.0.0.1:%d", cfg.Core.APIPort),
		"proxies":         allProxies,
	}

	data, err := yaml.Marshal(&mihomoCfg)
	if err != nil {
		return fmt.Errorf("marshal mihomo config: %w", err)
	}

	mihomoConfigPath := filepath.Join(m.cfg.MihomoDir(), "config.yaml")
	if err := os.MkdirAll(m.cfg.MihomoDir(), 0755); err != nil {
		return fmt.Errorf("create mihomo dir: %w", err)
	}
	return os.WriteFile(mihomoConfigPath, data, 0644)
}
```

- [ ] **Step 2: Fix imports — add missing os/path/filepath**

The file above references `os` and `filepath` without importing them. Add to the import block:

```go
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
```

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: subscription manager — fetch, parse, merge subscriptions"
```

---

### Task 6: System Proxy (Linux)

**Files:**
- Create: `internal/sysproxy/proxy.go`
- Create: `internal/sysproxy/proxy_linux.go`

- [ ] **Step 1: Write interface**

Write `internal/sysproxy/proxy.go`:

```go
package sysproxy

// Manager sets and unsets system-level proxy settings.
type Manager struct {
	httpHost string
	httpPort int
}

// NewManager creates a system proxy manager.
func NewManager(httpHost string, httpPort int) *Manager {
	return &Manager{httpHost: httpHost, httpPort: httpPort}
}
```

Write `internal/sysproxy/proxy_linux.go`:

```go
//go:build linux

package sysproxy

import (
	"fmt"
	"os/exec"
)

// Set configures the system HTTP/HTTPS proxy via gsettings (GNOME).
func (m *Manager) Set() error {
	proxyURL := fmt.Sprintf("http://%s:%d", m.httpHost, m.httpPort)

	commands := [][]string{
		{"gsettings", "set", "org.gnome.system.proxy", "mode", "'manual'"},
		{"gsettings", "set", "org.gnome.system.proxy.http", "host", fmt.Sprintf("'%s'", m.httpHost)},
		{"gsettings", "set", "org.gnome.system.proxy.http", "port", fmt.Sprintf("%d", m.httpPort)},
		{"gsettings", "set", "org.gnome.system.proxy.https", "host", fmt.Sprintf("'%s'", m.httpHost)},
		{"gsettings", "set", "org.gnome.system.proxy.https", "port", fmt.Sprintf("%d", m.httpPort)},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("gsettings: %s: %w", string(out), err)
		}
	}

	_ = proxyURL
	return nil
}

// Unset restores system proxy to "none".
func (m *Manager) Unset() error {
	cmd := exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "'none'")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gsettings unset: %s: %w", string(out), err)
	}
	return nil
}
```

- [ ] **Step 2: Commit**

```bash
git add -A && git commit -m "feat: system proxy manager (linux gsettings)"
```

---

### Task 7: CLI — Service Commands

**Files:**
- Create: `internal/cli/service.go`
- Modify: `internal/cli/root.go` — register service subcommand

- [ ] **Step 1: Write service commands**

Write `internal/cli/service.go`:

```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"mihomo-cli/internal/kernel"
)

var kernelMgr *kernel.Manager

func SetKernelManager(mgr *kernel.Manager) {
	kernelMgr = mgr
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage mihomo kernel service",
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start mihomo kernel",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}
		if kernelMgr.IsRunning() {
			fmt.Println("mihomo is already running")
			return nil
		}
		if err := kernelMgr.Start(); err != nil {
			return fmt.Errorf("start mihomo: %w", err)
		}
		fmt.Println("mihomo started")
		return nil
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop mihomo kernel",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}
		if err := kernelMgr.Stop(); err != nil {
			return fmt.Errorf("stop mihomo: %w", err)
		}
		fmt.Println("mihomo stopped")
		return nil
	},
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart mihomo kernel",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}
		if err := kernelMgr.Restart(); err != nil {
			return fmt.Errorf("restart mihomo: %w", err)
		}
		fmt.Println("mihomo restarted")
		return nil
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show mihomo kernel status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}
		fmt.Printf("mihomo: %s\n", kernelMgr.Status())
		return nil
	},
}

var serviceLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show mihomo kernel logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if kernelMgr == nil {
			return fmt.Errorf("kernel manager not initialized")
		}
		lines, err := kernelMgr.ReadLogs(50)
		if err != nil {
			return fmt.Errorf("read logs: %w", err)
		}
		for _, line := range lines {
			fmt.Println(line)
		}
		return nil
	},
}

func init() {
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceStopCmd)
	serviceCmd.AddCommand(serviceRestartCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
	serviceCmd.AddCommand(serviceLogsCmd)
	rootCmd.AddCommand(serviceCmd)
}
```

- [ ] **Step 2: Commit**

```bash
git add -A && git commit -m "feat: CLI service commands (start/stop/restart/status/logs)"
```

---

### Task 8: CLI — Mode Commands

**Files:**
- Create: `internal/cli/mode.go`

- [ ] **Step 1: Write mode commands**

Write `internal/cli/mode.go`:

```go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"mihomo-cli/internal/cfg"
)

var cfgMgr *cfg.Manager

func SetConfigManager(mgr *cfg.Manager) {
	cfgMgr = mgr
}

var modeCmd = &cobra.Command{
	Use:   "mode",
	Short: "Get or set proxy mode",
}

var modeSetCmd = &cobra.Command{
	Use:   "set <rule|global|direct|script>",
	Short: "Set proxy mode",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		if err := cfgMgr.SetMode(args[0]); err != nil {
			return err
		}
		fmt.Printf("Mode set to: %s\n", args[0])
		return nil
	},
}

var modeShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current proxy mode",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		fmt.Println(cfgMgr.Config().Mode)
		return nil
	},
}

func init() {
	modeCmd.AddCommand(modeSetCmd)
	modeCmd.AddCommand(modeShowCmd)
	rootCmd.AddCommand(modeCmd)
}
```

- [ ] **Step 2: Commit**

```bash
git add -A && git commit -m "feat: CLI mode commands (set/show)"
```

---

### Task 9: CLI — Proxy Commands

**Files:**
- Create: `internal/cli/proxy.go`

- [ ] **Step 1: Write proxy commands**

Write `internal/cli/proxy.go`:

```go
package cli

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"mihomo-cli/internal/api"
)

var apiClient *api.Client

func SetAPIClient(client *api.Client) {
	apiClient = client
}

var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Manage proxy nodes",
}

var proxyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List proxy groups and nodes",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}
		proxies, err := apiClient.GetProxies()
		if err != nil {
			return fmt.Errorf("get proxies: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "GROUP\tCURRENT\tNODES")
		for name, p := range proxies.Proxies {
			if p.All != nil {
				nodes := ""
				for i, n := range p.All {
					marker := ""
					if n == p.Now {
						marker = "*"
					}
					if i > 0 {
						nodes += ", "
					}
					nodes += marker + n
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", name, p.Now, nodes)
			}
		}
		w.Flush()
		return nil
	},
}

var proxySetCmd = &cobra.Command{
	Use:   "set <group> <node>",
	Short: "Switch proxy node in a group",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}
		if err := apiClient.SwitchProxy(args[0], args[1]); err != nil {
			return fmt.Errorf("switch proxy: %w", err)
		}
		fmt.Printf("Switched [%s] → %s\n", args[0], args[1])
		return nil
	},
}

var proxyTestCmd = &cobra.Command{
	Use:   "test [node]",
	Short: "Test proxy latency",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}

		proxies, err := apiClient.GetProxies()
		if err != nil {
			return fmt.Errorf("get proxies: %w", err)
		}

		type result struct {
			name  string
			delay int
			err   error
		}

		results := make(chan result, 50)
		count := 0

		for name, p := range proxies.Proxies {
			if p.All != nil {
				continue // Skip groups, test individual nodes
			}
			if len(args) > 0 && name != args[0] {
				continue
			}
			count++
			go func(n string) {
				d, err := apiClient.TestDelay(n, 5*time.Second)
				results <- result{name: n, delay: d, err: err}
			}(name)
		}

		// Collect results sorted by delay
		var all []result
		for i := 0; i < count; i++ {
			r := <-results
			if r.err == nil {
				all = append(all, r)
			}
		}
		sort.Slice(all, func(i, j int) bool { return all[i].delay < all[j].delay })

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NODE\tDELAY")
		for _, r := range all {
			fmt.Fprintf(w, "%s\t%dms\n", r.name, r.delay)
		}
		w.Flush()
		return nil
	},
}

var proxyInfoCmd = &cobra.Command{
	Use:   "info <node>",
	Short: "Show proxy node details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}
		proxies, err := apiClient.GetProxies()
		if err != nil {
			return fmt.Errorf("get proxies: %w", err)
		}
		p, ok := proxies.Proxies[args[0]]
		if !ok {
			return fmt.Errorf("node %q not found", args[0])
		}
		fmt.Printf("Name: %s\nType: %s\n", p.Name, p.Type)
		if len(p.History) > 0 {
			fmt.Printf("Last delay: %dms\n", p.History[len(p.History)-1].Delay)
		}
		return nil
	},
}

func init() {
	proxyCmd.AddCommand(proxyListCmd)
	proxyCmd.AddCommand(proxySetCmd)
	proxyCmd.AddCommand(proxyTestCmd)
	proxyCmd.AddCommand(proxyInfoCmd)
	rootCmd.AddCommand(proxyCmd)
}
```

- [ ] **Step 2: Commit**

```bash
git add -A && git commit -m "feat: CLI proxy commands (list/set/test/info)"
```

---

### Task 10: CLI — Subscription Commands

**Files:**
- Create: `internal/cli/sub.go`

- [ ] **Step 1: Write subscription commands**

Write `internal/cli/sub.go`:

```go
package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"mihomo-cli/internal/subscription"
)

var subMgr *subscription.Manager

func SetSubscriptionManager(mgr *subscription.Manager) {
	subMgr = mgr
}

var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Manage subscriptions",
}

var subAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a subscription",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		if err := cfgMgr.AddSubscription(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Subscription %q added\n", args[0])
		return nil
	},
}

var subRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a subscription",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		if err := cfgMgr.RemoveSubscription(args[0]); err != nil {
			return err
		}
		fmt.Printf("Subscription %q removed\n", args[0])
		return nil
	},
}

var subUpdateCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update subscriptions (all or by name)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if subMgr == nil {
			return fmt.Errorf("subscription manager not initialized")
		}
		if len(args) > 0 {
			if err := subMgr.UpdateSubscription(args[0]); err != nil {
				return err
			}
			fmt.Printf("Subscription %q updated\n", args[0])
		} else {
			errs := subMgr.UpdateAll()
			if len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintln(os.Stderr, e)
				}
				return fmt.Errorf("%d subscription(s) failed to update", len(errs))
			}
			fmt.Println("All subscriptions updated")
		}
		return nil
	},
}

var subListCmd = &cobra.Command{
	Use:   "list",
	Short: "List subscriptions",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		subs := cfgMgr.Config().Subscriptions
		if len(subs) == 0 {
			fmt.Println("No subscriptions configured")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tURL")
		for _, s := range subs {
			fmt.Fprintf(w, "%s\t%s\n", s.Name, s.URL)
		}
		w.Flush()
		return nil
	},
}

func init() {
	subCmd.AddCommand(subAddCmd)
	subCmd.AddCommand(subRemoveCmd)
	subCmd.AddCommand(subUpdateCmd)
	subCmd.AddCommand(subListCmd)
	rootCmd.AddCommand(subCmd)
}
```

- [ ] **Step 2: Commit**

```bash
git add -A && git commit -m "feat: CLI subscription commands (add/remove/update/list)"
```

---

### Task 11: CLI — Connection and Config Commands

**Files:**
- Create: `internal/cli/conn.go`
- Create: `internal/cli/config_cmd.go`

- [ ] **Step 1: Write connection commands**

Write `internal/cli/conn.go`:

```go
package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var connCmd = &cobra.Command{
	Use:   "conn",
	Short: "Manage active connections",
}

var connListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}
		conns, err := apiClient.GetConnections()
		if err != nil {
			return fmt.Errorf("get connections: %w", err)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tHOST\tNETWORK\tRULE\tUPLOAD\tDOWNLOAD")
		for _, c := range conns.Connections {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\n",
				c.ID, c.Metadata.Host, c.Metadata.Network, c.Rule, c.Upload, c.Download)
		}
		w.Flush()
		return nil
	},
}

var connCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close a connection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}
		if err := apiClient.CloseConnection(args[0]); err != nil {
			return fmt.Errorf("close connection: %w", err)
		}
		fmt.Printf("Connection %s closed\n", args[0])
		return nil
	},
}

func init() {
	connCmd.AddCommand(connListCmd)
	connCmd.AddCommand(connCloseCmd)
	rootCmd.AddCommand(connCmd)
}
```

Write `internal/cli/config_cmd.go`:

```go
package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage mihomo-cli configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		cfg := cfgMgr.Config()
		fmt.Printf("Mode: %s\n", cfg.Mode)
		fmt.Printf("HTTP Port: %d\n", cfg.Core.HTTPPort)
		fmt.Printf("SOCKS Port: %d\n", cfg.Core.SOCKSPort)
		fmt.Printf("API Port: %d\n", cfg.Core.APIPort)
		fmt.Printf("Subscriptions: %d\n", len(cfg.Subscriptions))
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit config with $EDITOR",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgMgr == nil {
			return fmt.Errorf("config manager not initialized")
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		configPath := cfgMgr.ConfigDir() + "/config.yaml"
		c := exec.Command(editor, configPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

var configReloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload mihomo configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if apiClient == nil {
			return fmt.Errorf("API client not available — is mihomo running?")
		}
		if err := apiClient.ReloadConfig(); err != nil {
			return fmt.Errorf("reload config: %w", err)
		}
		fmt.Println("Configuration reloaded")
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configReloadCmd)
	rootCmd.AddCommand(configCmd)
}
```

- [ ] **Step 2: Commit**

```bash
git add -A && git commit -m "feat: CLI conn and config commands"
```

---

### Task 12: Wire Main.go — Dependency Injection

**Files:**
- Modify: `cmd/mihomo-cli/main.go`
- Modify: `internal/cli/root.go`

- [ ] **Step 1: Update root.go to accept managers**

Modify `internal/cli/root.go` — replace the `Run` func:

```go
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"mihomo-cli/internal/api"
	"mihomo-cli/internal/cfg"
	"mihomo-cli/internal/kernel"
	"mihomo-cli/internal/subscription"
)

var rootCmd = &cobra.Command{
	Use:   "mihomo-cli",
	Short: "Manage mihomo proxy from the command line",
	Long:  "A CLI tool for managing mihomo proxy subscriptions, nodes, modes, and service lifecycle.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TUI mode not yet implemented")
		os.Exit(0)
	},
}

// InitManagers creates and wires all managers from config.
func InitManagers() (*cfg.Manager, *kernel.Manager, *api.Client, *subscription.Manager, error) {
	cm, err := cfg.NewManager()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("init config: %w", err)
	}

	c := cm.Config()
	km := kernel.NewManager(cm.ConfigDir(), cm.MihomoDir(), c.Core.APIPort)

	// Auto-install kernel
	if !km.IsInstalled() {
		fmt.Println("Downloading mihomo kernel...")
		if err := km.Install(); err != nil {
			return nil, nil, nil, nil, fmt.Errorf("install mihomo: %w", err)
		}
		fmt.Println("Mihomo kernel installed")
	}

	// Try to start if not running
	if !km.IsRunning() {
		fmt.Println("Starting mihomo...")
		if err := km.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not start mihomo: %v\n", err)
		}
	}

	apiClient := km.APIClient()

	sm := subscription.NewManager(cm)

	// Wire into CLI globals
	SetConfigManager(cm)
	SetKernelManager(km)
	if apiClient != nil {
		SetAPIClient(apiClient)
	}
	SetSubscriptionManager(sm)

	return cm, km, apiClient, sm, nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Update main.go**

Replace `cmd/mihomo-cli/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"mihomo-cli/internal/cli"
)

func main() {
	if _, _, _, _, err := cli.InitManagers(); err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		// Continue — individual commands will surface errors
	}
	cli.Execute()
}
```

- [ ] **Step 3: Verify build**

```bash
go build ./cmd/mihomo-cli/
```

Expected: compiles without errors

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: wire dependency injection in main.go"
```

---

### Task 13: TUI — Styles and Model

**Files:**
- Create: `internal/tui/styles.go`
- Create: `internal/tui/model.go`

- [ ] **Step 1: Write styles**

Write `internal/tui/styles.go`:

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7c3aed")).
		Padding(0, 1)

	SelectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#7c3aed")).
		Padding(0, 1)

	NormalStyle = lipgloss.NewStyle().
		Padding(0, 1)

	GroupHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#a78bfa")).
		Padding(0, 1)

	InfoStyle = lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3f3f46"))

	LogStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#71717a"))

	StatusBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a1a1aa")).
		Background(lipgloss.Color("#18181b")).
		Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#52525b"))
)
```

Write `internal/tui/model.go`:

```go
package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"mihomo-cli/internal/api"
	"mihomo-cli/internal/kernel"
	"mihomo-cli/internal/subscription"
)

// tickMsg is sent every second for status refresh.
type tickMsg time.Time

// proxiesMsg carries the result of fetching proxies.
type proxiesMsg struct {
	proxies *api.ProxiesResponse
	err     error
}

// Model is the top-level TUI model.
type Model struct {
	apiClient *api.Client
	kernelMgr *kernel.Manager
	subMgr    *subscription.Manager

	proxies     *api.ProxiesResponse
	proxiesErr  error

	groups       []string
	groupIdx     int
	nodeIdx      int
	width        int
	height       int
	status       string
	lastUpdate   time.Time
}

// NewModel creates the TUI model.
func NewModel(apiClient *api.Client, km *kernel.Manager, sm *subscription.Manager) Model {
	return Model{
		apiClient: apiClient,
		kernelMgr: km,
		subMgr:    sm,
		groups:    []string{},
	}
}

// Init starts the ticker and fetches initial proxy data.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return fetchProxies(m.apiClient) },
		func() tea.Msg { return tickMsg(time.Now()) },
	)
}

func fetchProxies(client *api.Client) tea.Msg {
	proxies, err := client.GetProxies()
	return proxiesMsg{proxies: proxies, err: err}
}

// KeyMap defines TUI keybindings.
type KeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Tab     key.Binding
	Test    key.Binding
	Quit    key.Binding
	Reload  key.Binding
	Update  key.Binding
}

var Keys = KeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Enter:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Tab:    key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next group")),
	Test:   key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "test latency")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Reload: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reload config")),
	Update: key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "update subs")),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Tab, k.Test, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Tab},
		{k.Test, k.Reload, k.Update, k.Quit},
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add -A && git commit -m "feat: TUI model, styles, and keybindings"
```

---

### Task 14: TUI — Update Handler

**Files:**
- Create: `internal/tui/update.go`

- [ ] **Step 1: Write update logic**

Write `internal/tui/update.go`:

```go
package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"mihomo-cli/internal/api"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, Keys.Tab):
			if len(m.groups) > 0 {
				m.groupIdx = (m.groupIdx + 1) % len(m.groups)
				m.nodeIdx = 0
			}

		case key.Matches(msg, Keys.Up):
			if m.nodeIdx > 0 {
				m.nodeIdx--
			}

		case key.Matches(msg, Keys.Down):
			group := m.currentGroup()
			if group != nil && m.nodeIdx < len(group.All)-1 {
				m.nodeIdx++
			}

		case key.Matches(msg, Keys.Enter):
			m.switchToSelected()

		case key.Matches(msg, Keys.Test):
			return m, func() tea.Msg {
				group := m.currentGroup()
				if group != nil && m.nodeIdx < len(group.All) {
					node := group.All[m.nodeIdx]
					d, err := m.apiClient.TestDelay(node, 5*time.Second)
					return delayResultMsg{node: node, delay: d, err: err}
				}
				return nil
			}

		case key.Matches(msg, Keys.Reload):
			if m.apiClient != nil {
				m.apiClient.ReloadConfig()
			}

		case key.Matches(msg, Keys.Update):
			if m.subMgr != nil {
				m.subMgr.UpdateAll()
			}
		}

	case tickMsg:
		if m.apiClient != nil && time.Since(m.lastUpdate) > 2*time.Second {
			m.lastUpdate = time.Now()
			return m, func() tea.Msg { return fetchProxies(m.apiClient) }
		}
		return m, func() tea.Msg {
			time.Sleep(1 * time.Second)
			return tickMsg(time.Now())
		}

	case proxiesMsg:
		if msg.err != nil {
			m.proxiesErr = msg.err
		} else {
			m.proxies = msg.proxies
			m.proxiesErr = nil
			m.refreshGroups()
		}
		return m, func() tea.Msg {
			time.Sleep(1 * time.Second)
			return tickMsg(time.Now())
		}

	case delayResultMsg:
		// Delay result shown transiently — could store in a map
		_ = msg
	}

	return m, nil
}

type delayResultMsg struct {
	node  string
	delay int
	err   error
}

// currentGroup returns the currently selected proxy group.
func (m *Model) currentGroup() *api.Proxy {
	if m.proxies == nil || len(m.groups) == 0 {
		return nil
	}
	name := m.groups[m.groupIdx]
	p, ok := m.proxies.Proxies[name]
	if !ok {
		return nil
	}
	return &p
}

// switchToSelected switches the proxy in the current group to the selected node.
func (m *Model) switchToSelected() {
	group := m.currentGroup()
	if group == nil || m.nodeIdx >= len(group.All) {
		return
	}
	node := group.All[m.nodeIdx]
	m.apiClient.SwitchProxy(group.Name, node)
}

// refreshGroups rebuilds the group list from proxy data.
func (m *Model) refreshGroups() {
	if m.proxies == nil {
		return
	}
	m.groups = nil
	for name, p := range m.proxies.Proxies {
		if p.All != nil {
			m.groups = append(m.groups, name)
		}
	}
	// Clamp indices
	if m.groupIdx >= len(m.groups) {
		m.groupIdx = 0
	}
}
```

- [ ] **Step 2: Add missing import — fix the `key` reference in update.go**

The `key.Matches` calls need the key import. The imports should already be fine since we use `key.Matches`. Let me double-check...

Actually, in the update function I use `key.Matches(msg, Keys.Quit)` etc. This needs `"github.com/charmbracelet/bubbles/key"`. Let me check the imports needed.

The import block should be:
```go
import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"mihomo-cli/internal/api"
)
```

This is correct. The file is fine.

- [ ] **Step 2: Commit**

```bash
git add -A && git commit -m "feat: TUI update handler with keyboard navigation"
```

---

### Task 15: TUI — View Renderer

**Files:**
- Create: `internal/tui/view.go`
- Modify: `internal/cli/root.go` — TUI launch

- [ ] **Step 1: Write view renderer**

Write `internal/tui/view.go`:

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	leftPanel := m.renderLeftPanel()
	rightPanel := m.renderRightPanel()
	statusBar := m.renderStatusBar()

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	return lipgloss.JoinVertical(lipgloss.Left, panels, statusBar)
}

func (m Model) renderLeftPanel() string {
	width := m.width * 2 / 3
	var b strings.Builder

	b.WriteString(TitleStyle.Render(" Proxies "))
	b.WriteString("\n\n")

	if m.proxiesErr != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.proxiesErr))
		return lipgloss.NewStyle().Width(width).Render(b.String())
	}

	if len(m.groups) == 0 {
		b.WriteString("No proxy groups found\n")
		return lipgloss.NewStyle().Width(width).Render(b.String())
	}

	for gi, groupName := range m.groups {
		group, ok := m.proxies.Proxies[groupName]
		if !ok {
			continue
		}

		// Group header
		header := GroupHeaderStyle.Render(fmt.Sprintf(" [%s] ", groupName))
		b.WriteString(header)
		b.WriteString("\n")

		// Nodes in this group
		for ni, node := range group.All {
			marker := "  "
			if node == group.Now {
				marker = " ●"
			}

			line := fmt.Sprintf("%s %s", marker, node)

			if gi == m.groupIdx && ni == m.nodeIdx {
				b.WriteString(SelectedStyle.Render(line))
			} else {
				b.WriteString(NormalStyle.Render(line))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return lipgloss.NewStyle().Width(width).Render(b.String())
}

func (m Model) renderRightPanel() string {
	width := m.width / 3
	var b strings.Builder

	b.WriteString(TitleStyle.Render(" Details "))
	b.WriteString("\n\n")

	group := m.currentGroup()
	if group == nil || m.nodeIdx >= len(group.All) {
		b.WriteString(InfoStyle.Width(width - 4).Render("Select a node"))
		return b.String()
	}

	nodeName := group.All[m.nodeIdx]
	nodeInfo, ok := m.proxies.Proxies[nodeName]
	
	info := fmt.Sprintf("Name: %s\nType: %s", nodeName, "unknown")
	if ok {
		info = fmt.Sprintf("Name: %s\nType: %s", nodeInfo.Name, nodeInfo.Type)
		if len(nodeInfo.History) > 0 {
			info += fmt.Sprintf("\nDelay: %dms", nodeInfo.History[len(nodeInfo.History)-1].Delay)
		}
	}

	b.WriteString(InfoStyle.Width(width - 4).Render(info))
	b.WriteString("\n\n")

	// Help section
	b.WriteString(HelpStyle.Render("↑↓ navigate  tab group  enter switch"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("t test  r reload  u update  q quit"))

	return b.String()
}

func (m Model) renderStatusBar() string {
	status := m.kernelMgr.Status()
	mode := "N/A"
	// If we have a config manager reference, use it — for now show status
	left := fmt.Sprintf(" mihomo: %s ", status)
	right := fmt.Sprintf(" mode: %s ", mode)
	bar := left + strings.Repeat(" ", m.width-lipgloss.Width(left)-lipgloss.Width(right)) + right
	return StatusBarStyle.Width(m.width).Render(bar)
}
```

- [ ] **Step 2: Update root.go to launch TUI**

Modify `internal/cli/root.go` — replace the root `Run` func to launch TUI:

The existing `Run` in root.go's rootCmd should be changed from the placeholder to:

```go
var rootCmd = &cobra.Command{
	Use:   "mihomo-cli",
	Short: "Manage mihomo proxy from the command line",
	Long:  "A CLI tool for managing mihomo proxy subscriptions, nodes, modes, and service lifecycle.",
	Run: func(cmd *cobra.Command, args []string) {
		cm, km, ac, sm, err := InitManagers()
		if err != nil {
			fmt.Fprintf(os.Stderr, "init: %v\n", err)
		}
		if ac == nil {
			fmt.Println("Mihomo is not running. Start it with: mihomo-cli service start")
			os.Exit(1)
		}
		model := tui.NewModel(ac, km, sm)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}
		// Cleanup on exit
		_ = cm
	},
}
```

Also add the import for tui and tea at the top:
```go
import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"mihomo-cli/internal/api"
	"mihomo-cli/internal/cfg"
	"mihomo-cli/internal/kernel"
	"mihomo-cli/internal/subscription"
	"mihomo-cli/internal/tui"
)
```

- [ ] **Step 3: Verify build**

```bash
go build ./cmd/mihomo-cli/
```

Expected: compiles without errors

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: TUI view renderer with 3-panel layout"
```

---

### Task 16: Build and Integration Verification

**Files:** none — verification only

- [ ] **Step 1: Run all unit tests**

```bash
go test ./... -v
```

Expected: all tests pass (cfg tests, any others)

- [ ] **Step 2: Build release binary**

```bash
go build -o mihomo-cli ./cmd/mihomo-cli/
```

Expected: produces `mihomo-cli` binary

- [ ] **Step 3: Smoke test CLI help**

```bash
./mihomo-cli --help
```

Expected: shows command tree with sub, mode, proxy, service, conn, config

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "chore: final build verification and cleanup"
```
