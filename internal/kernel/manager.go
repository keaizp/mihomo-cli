package kernel

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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
		binPath: filepath.Join(binDir, "bin", "mihomo"),
		workDir: workDir,
		apiPort: apiPort,
	}
}

// BinPath returns the expected path of the mihomo binary.
func (m *Manager) BinPath() string { return m.binPath }

// IsInstalled returns true if the mihomo binary exists.
func (m *Manager) IsInstalled() bool {
	_, err := os.Stat(m.binPath)
	return err == nil
}

// InstallFrom copies a local binary to the managed path.
func (m *Manager) InstallFrom(srcPath string) error {
	if err := os.MkdirAll(filepath.Dir(m.binPath), 0755); err != nil {
		return fmt.Errorf("create bin dir: %w", err)
	}
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(m.binPath)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy binary: %w", err)
	}
	return os.Chmod(m.binPath, 0755)
}

// defaultURL builds the default GitHub download URL for the current arch.
func defaultURL() string {
	arch := runtime.GOARCH
	version := "v1.18.10"
	return fmt.Sprintf("%s/download/%s/mihomo-linux-%s-%s.gz",
		mihomoRepo, version, arch, version)
}

// Install downloads the mihomo binary from GitHub with a progress bar.
func (m *Manager) Install() error {
	return m.InstallFromURL(defaultURL())
}

// InstallFromURL downloads the mihomo binary from a custom URL with a progress bar.
func (m *Manager) InstallFromURL(url string) error {
	if err := os.MkdirAll(filepath.Dir(m.binPath), 0755); err != nil {
		return fmt.Errorf("create bin dir: %w", err)
	}

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

	var success bool
	defer func() {
		f.Close()
		if !success {
			os.Remove(tmpPath)
		}
	}()

	// Decompress gzip on the fly
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("decompress gzip: %w", err)
	}
	defer gzReader.Close()

	// Copy with progress bar
	bar := &progressBar{reader: gzReader, total: resp.ContentLength, writer: os.Stderr}
	if _, err := io.Copy(f, bar); err != nil {
		return fmt.Errorf("write binary: %w", err)
	}
	fmt.Fprintln(os.Stderr) // newline after progress bar

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}
	if err := os.Rename(tmpPath, m.binPath); err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	success = true
	return nil
}

// progressBar wraps a reader and prints a progress bar to the writer.
type progressBar struct {
	reader io.Reader
	total  int64
	writer io.Writer
	read   int64
	last   time.Time
}

func (p *progressBar) Read(buf []byte) (int, error) {
	n, err := p.reader.Read(buf)
	p.read += int64(n)

	// Throttle updates to ~10 per second
	if p.last.IsZero() || time.Since(p.last) > 100*time.Millisecond || err != nil {
		p.last = time.Now()
		p.draw()
	}
	return n, err
}

func (p *progressBar) draw() {
	barWidth := 40

	var percent float64
	if p.total > 0 {
		percent = float64(p.read) / float64(p.total)
	} else {
		// Unknown total — use gzipped size as rough guide, or just show bytes
		fmt.Fprintf(p.writer, "\r  %s downloaded\r", formatBytes(p.read))
		return
	}

	filled := int(percent * float64(barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	fmt.Fprintf(p.writer, "\r  [%s] %.0f%%  %s/%s",
		bar, percent*100,
		formatBytes(p.read), formatBytes(p.total))
}

func formatBytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

// Start launches mihomo as a child process.
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd != nil {
		return nil
	}

	// 1. PID file check: if mihomo process is already alive, connect to it
	pid := m.readPID()
	if pid > 0 && isProcessAlive(pid) {
		if m.probe() {
			m.apiClient = api.NewClient(fmt.Sprintf(apiBaseURL, m.apiPort))
			return nil
		}
		// Process is alive but API not ready (still starting up). Wait and retry.
		for i := 0; i < 10; i++ {
			time.Sleep(500 * time.Millisecond)
			if m.probe() {
				m.apiClient = api.NewClient(fmt.Sprintf(apiBaseURL, m.apiPort))
				return nil
			}
		}
		return fmt.Errorf("mihomo (PID %d) is not responding after 5s", pid)
	}

	// 2. Clean up stale PID file
	if pid > 0 {
		m.cleanPID()
	}

	// 3. Probe API — maybe mihomo was started externally
	if m.probe() {
		m.apiClient = api.NewClient(fmt.Sprintf(apiBaseURL, m.apiPort))
		return nil
	}

	// 4. No mihomo running — launch a new process
	if err := os.MkdirAll(m.workDir, 0755); err != nil {
		return fmt.Errorf("create work dir: %w", err)
	}

	logDir := filepath.Join(m.workDir, "logs")
	os.MkdirAll(logDir, 0755)
	logFile, err := os.Create(filepath.Join(logDir, "mihomo.log"))
	if err != nil {
		logFile = nil
	}

	configPath := filepath.Join(m.workDir, "config.yaml")
	m.cmd = exec.Command(m.binPath, "-d", m.workDir, "-f", configPath)
	if logFile != nil {
		m.cmd.Stdout = logFile
		m.cmd.Stderr = logFile
	} else {
		m.cmd.Stdout = os.Stderr
		m.cmd.Stderr = os.Stderr
	}

	if err := m.cmd.Start(); err != nil {
		if logFile != nil {
			logFile.Close()
		}
		m.cmd = nil
		return fmt.Errorf("start mihomo: %w", err)
	}

	// Write PID immediately so other invocations see it
	m.writePID(m.cmd.Process.Pid)

	time.Sleep(startupWait)
	m.apiClient = api.NewClient(fmt.Sprintf(apiBaseURL, m.apiPort))

	return nil
}

// Stop terminates the mihomo process.
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pid := m.readPID()
	if pid > 0 {
		if proc, err := os.FindProcess(pid); err == nil {
			proc.Kill()
		}
		m.cleanPID()
	}

	if m.cmd != nil && m.cmd.Process != nil {
		m.cmd.Process.Kill()
		m.cmd.Wait()
		m.cmd = nil
	}

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

	if m.apiClient != nil && m.apiClient.HealthCheck() == nil {
		return true
	}

	pid := m.readPID()
	if pid > 0 && isProcessAlive(pid) {
		if m.probe() {
			m.apiClient = api.NewClient(fmt.Sprintf(apiBaseURL, m.apiPort))
			return true
		}
	}

	if pid > 0 && !isProcessAlive(pid) {
		m.cleanPID()
	}

	return false
}

// APIClient returns the API client, or nil if mihomo is not running.
func (m *Manager) APIClient() *api.Client {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.apiClient != nil {
		return m.apiClient
	}
	if m.probe() {
		m.apiClient = api.NewClient(fmt.Sprintf(apiBaseURL, m.apiPort))
		return m.apiClient
	}
	return nil
}

// Status returns a human-readable status string.
func (m *Manager) Status() string {
	if m.IsRunning() {
		return "running"
	}
	pid := m.readPID()
	if pid > 0 && isProcessAlive(pid) {
		return "starting"
	}
	return "stopped"
}

// ReadLogs reads the last n lines from the mihomo log file.
func (m *Manager) ReadLogs(n int) ([]string, error) {
	logPath := filepath.Join(m.workDir, "logs", "mihomo.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if n > 0 && len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}

// --- PID file helpers ---

const pidFileName = "mihomo.pid"

func (m *Manager) pidFilePath() string {
	return filepath.Join(m.workDir, pidFileName)
}

func (m *Manager) readPID() int {
	data, err := os.ReadFile(m.pidFilePath())
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return pid
}

func (m *Manager) writePID(pid int) error {
	return os.WriteFile(m.pidFilePath(), []byte(strconv.Itoa(pid)), 0644)
}

func (m *Manager) cleanPID() {
	os.Remove(m.pidFilePath())
}

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	_, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	return err == nil
}

// probe checks if the mihomo API is responding (2s timeout).
func (m *Manager) probe() bool {
	c := &http.Client{Timeout: 2 * time.Second}
	resp, err := c.Get(fmt.Sprintf(apiBaseURL+"/version", m.apiPort))
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}
