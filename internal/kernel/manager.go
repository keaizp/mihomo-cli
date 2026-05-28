package kernel

import (
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
	mu        sync.Mutex
	cmd       *exec.Cmd
	binPath   string
	workDir   string
	apiPort   int
	apiClient *api.Client
}

// NewManager creates a kernel manager.
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

// Install downloads the mihomo binary from GitHub.
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

	tmpPath := m.binPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpPath)

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

// ReadLogs reads log file entries from the mihomo log directory.
func (m *Manager) ReadLogs(n int) ([]string, error) {
	logPath := filepath.Join(m.workDir, "logs")
	entries, err := os.ReadDir(logPath)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	latest := entries[len(entries)-1]
	return []string{latest.Name()}, nil
}
