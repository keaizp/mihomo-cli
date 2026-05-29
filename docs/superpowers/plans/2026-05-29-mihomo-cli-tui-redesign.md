# TUI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign mihomo-cli TUI with tab-based multi-view architecture, Chinese localization, and comprehensive proxy management features.

**Architecture:** Tab-based multi-view TUI (5 tabs: 代理/连接/日志/规则/订阅) using bubbletea + lipgloss. Model holds all state; view.go routes rendering by tab index; update.go handles tab switching and per-tab keybindings. New API methods added for logs and rules endpoints.

**Tech Stack:** Go 1.24, bubbletea v1.3, lipgloss v1.1, bubbles v1.0

---

### Task 1: Update API Client with Logs and Rules Endpoints

**Files:**
- Modify: `internal/api/client.go`

- [ ] **Step 1: Add LogEntry, RulesResponse types and GetLogs/GetRules methods**

Add to `internal/api/client.go` after the existing type definitions (after line 71):

```go
// LogEntry represents a single log line from mihomo.
type LogEntry struct {
	Type    string `json:"type"`    // "info", "warning", "error", "debug"
	Payload string `json:"payload"`
}

// Rule represents a routing rule.
type Rule struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Proxy   string `json:"proxy"`
}

// RulesResponse is the response from GET /rules.
type RulesResponse struct {
	Rules []Rule `json:"rules"`
}
```

Add after `CloseConnection` method (after line 130):

```go
// GetLogs fetches recent log entries from mihomo.
func (c *Client) GetLogs() ([]LogEntry, error) {
	var logs []LogEntry
	if err := c.do("GET", "/logs", nil, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// GetRules fetches all routing rules.
func (c *Client) GetRules() (*RulesResponse, error) {
	var resp RulesResponse
	if err := c.do("GET", "/rules", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: builds without errors

- [ ] **Step 3: Commit**

```bash
git add internal/api/client.go
git commit -m "feat: add GetLogs and GetRules API methods for TUI"
```

---

### Task 2: Rewrite Styles with Comprehensive Design System

**Files:**
- Modify: `internal/tui/styles.go`

- [ ] **Step 1: Replace styles.go content**

Replace the entire content of `internal/tui/styles.go`:

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
const (
	colorPrimary = "#7c3aed"
	colorSuccess = "#22c55e"
	colorWarning = "#eab308"
	colorDanger  = "#ef4444"
	colorMuted   = "#71717a"
	colorBg      = "#18181b"
	colorBgLight = "#27272a"
	colorBorder  = "#3f3f46"
	colorText    = "#e4e4e7"
	colorTextDim = "#a1a1aa"
	colorCyan    = "#06b6d4"
)

// Text styles
var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorPrimary))

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorText))

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTextDim))

	// Selection highlight
	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color(colorPrimary))

	// Group headers in proxy list
	GroupHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#a78bfa"))

	// Proxy type badge
	TypeBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorCyan)).
			Background(lipgloss.Color(colorBgLight)).
			Padding(0, 1)
)

// Status indicators
var (
	StatusRunningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorSuccess)).
				Bold(true)

	StatusStoppedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorDanger)).
				Bold(true)

	StatusStartingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorWarning)).
				Bold(true)
)

// Tab bar
var (
	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color(colorPrimary)).
			Padding(0, 2)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorMuted)).
				Padding(0, 2)
)

// Panel borders
var (
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Padding(0, 1)

	PanelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorPrimary)).
			Padding(0, 1)
)

// Footer / help bar
var (
	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTextDim)).
			Background(lipgloss.Color(colorBg)).
			Padding(0, 1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPrimary)).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTextDim))

	// Search bar
	SearchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorText)).
			Background(lipgloss.Color(colorBgLight)).
			Padding(0, 1)
)

// Log level colors
var (
	LogInfoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan))
	LogWarnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(colorWarning))
	LogErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDanger))
	LogDebugStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
)

// LatencyColor returns a lipgloss color for a given delay in ms.
func LatencyColor(delay int) lipgloss.Color {
	switch {
	case delay <= 0:
		return lipgloss.Color(colorMuted)
	case delay < 200:
		return lipgloss.Color(colorSuccess)
	case delay < 500:
		return lipgloss.Color(colorWarning)
	default:
		return lipgloss.Color(colorDanger)
	}
}

// LatencyBar returns a visual bar representing relative latency.
func LatencyBar(delay int, maxDelay int, width int) string {
	if delay <= 0 || maxDelay <= 0 {
		return strings.Repeat(" ", width)
	}
	ratio := float64(delay) / float64(maxDelay)
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	if filled < 1 && delay > 0 {
		filled = 1
	}
	bar := strings.Repeat("█", filled) + strings.Repeat(" ", width-filled)
	return LatencyColor(delay).Styled(bar)
}

// FormatBytes formats bytes to human-readable string.
func FormatBytes(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%dB", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1fKB", float64(n)/1024)
	case n < 1024*1024*1024:
		return fmt.Sprintf("%.1fMB", float64(n)/(1024*1024))
	default:
		return fmt.Sprintf("%.1fGB", float64(n)/(1024*1024*1024))
	}
}

// FormatRate formats bytes per second to human-readable string.
func FormatRate(bytesPerSec int64) string {
	return FormatBytes(bytesPerSec) + "/s"
}

// FormatDuration formats seconds to a readable duration.
func FormatDuration(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm", seconds/60)
	}
	if seconds < 86400 {
		return fmt.Sprintf("%dh", seconds/3600)
	}
	return fmt.Sprintf("%dd", seconds/86400)
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: builds without errors

- [ ] **Step 3: Commit**

```bash
git add internal/tui/styles.go
git commit -m "feat: comprehensive TUI style system with latency colors and formatters"
```

---

### Task 3: Rewrite Model with Expanded State and Chinese KeyMap

**Files:**
- Modify: `internal/tui/model.go`

- [ ] **Step 1: Replace model.go content**

Replace the entire content of `internal/tui/model.go`:

```go
package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"mihomo-cli/internal/api"
	"mihomo-cli/internal/cfg"
	"mihomo-cli/internal/kernel"
	"mihomo-cli/internal/subscription"
)

// Message types
type tickMsg time.Time

type proxiesMsg struct {
	proxies *api.ProxiesResponse
	err     error
}

type connectionsMsg struct {
	connections []api.Connection
	err         error
}

type logsMsg struct {
	logs []api.LogEntry
	err  error
}

type rulesMsg struct {
	rules []api.Rule
	err   error
}

type delayResultMsg struct {
	node  string
	delay int
	err   error
}

type speedtestDoneMsg struct {
	results map[string]int
	group   string
}

type notificationMsg string

// Model is the top-level TUI model.
type Model struct {
	apiClient *api.Client
	kernelMgr *kernel.Manager
	subMgr    *subscription.Manager
	cfgMgr    *cfg.Manager

	// Tab: 0=Proxies, 1=Connections, 2=Logs, 3=Rules, 4=Subs
	tabIdx int

	// Proxies tab
	proxies      *api.ProxiesResponse
	proxiesErr   error
	groups       []string
	groupIdx     int
	nodeIdx      int
	delayResults map[string]int // node → latency ms
	collapsed    map[string]bool
	speedtesting bool

	// Connections tab
	connections    []api.Connection
	connectionsErr error
	connIdx        int

	// Logs tab
	logs       []api.LogEntry
	logsErr    error
	logLevel   string // "", "info", "warning", "error", "debug"
	logScroll  int    // scroll offset from bottom (0 = auto-follow)
	autoScroll bool

	// Rules tab
	rules    []api.Rule
	rulesErr error
	ruleIdx  int

	// Subs tab (from cfgMgr)
	subIdx int

	// Search
	searchMode  bool
	searchQuery string

	// Traffic
	trafficUp   int64
	trafficDown int64

	// Window
	width      int
	height     int
	lastUpdate time.Time

	// Notification (transient message in footer area)
	notification string
}

// NewModel creates the TUI model.
func NewModel(apiClient *api.Client, km *kernel.Manager, sm *subscription.Manager, cm *cfg.Manager) Model {
	return Model{
		apiClient:    apiClient,
		kernelMgr:    km,
		subMgr:       sm,
		cfgMgr:       cm,
		groups:       []string{},
		delayResults: make(map[string]int),
		collapsed:    make(map[string]bool),
		autoScroll:   true,
	}
}

// Init starts the ticker and fetches initial data.
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

func fetchConnections(client *api.Client) tea.Msg {
	conns, err := client.GetConnections()
	if err != nil {
		return connectionsMsg{err: err}
	}
	return connectionsMsg{connections: conns.Connections}
}

func fetchLogs(client *api.Client) tea.Msg {
	logs, err := client.GetLogs()
	return logsMsg{logs: logs, err: err}
}

func fetchRules(client *api.Client) tea.Msg {
	resp, err := client.GetRules()
	if err != nil {
		return rulesMsg{err: err}
	}
	return rulesMsg{rules: resp.Rules}
}

// KeyMap with Chinese help text.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	TabPrev  key.Binding
	TabNext  key.Binding
	Test     key.Binding
	TestAll  key.Binding
	Quit     key.Binding
	Reload   key.Binding
	Update   key.Binding
	Mode     key.Binding
	Search   key.Binding
	Collapse key.Binding
	Close    key.Binding
	CloseAll key.Binding
	LogLevel key.Binding
	SubAdd   key.Binding
	SubEdit  key.Binding
	SubDel   key.Binding
	SubToggle key.Binding
}

var Keys = KeyMap{
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "上移")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "下移")),
	Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "选择")),
	TabPrev:  key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("S-Tab", "上一组")),
	TabNext:  key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "下一组")),
	Test:     key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "测速")),
	TestAll:  key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "全组测速")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"), key.WithHelp("q", "退出")),
	Reload:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "重载配置")),
	Update:   key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "更新订阅")),
	Mode:     key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "切换模式")),
	Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "搜索")),
	Collapse: key.NewBinding(key.WithKeys("space"), key.WithHelp("空格", "折叠")),
	Close:    key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "关闭连接")),
	CloseAll: key.NewBinding(key.WithKeys("X"), key.WithHelp("X", "关闭全部")),
	LogLevel: key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "日志级别")),
	SubAdd:   key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "添加订阅")),
	SubEdit:  key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "编辑订阅")),
	SubDel:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "删除订阅")),
	SubToggle: key.NewBinding(key.WithKeys("space"), key.WithHelp("空格", "启用/停用")),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.TabPrev, k.TabNext, k.Quit},
		{k.Test, k.TestAll, k.Search, k.Collapse, k.Reload, k.Update},
	}
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: builds without errors (may have unused import warnings which we'll resolve in later tasks)

- [ ] **Step 3: Commit**

```bash
git add internal/tui/model.go
git commit -m "feat: expanded TUI model with tab state, Chinese keymap, and new message types"
```

---

### Task 4: Wire cfgMgr into TUI Initialization

**Files:**
- Modify: `internal/cli/root.go`

- [ ] **Step 1: Pass cfgMgr to NewModel**

In `internal/cli/root.go`, change line 43 from:
```go
model := tui.NewModel(ac, kernelMgr, subMgr)
```
to:
```go
model := tui.NewModel(ac, kernelMgr, subMgr, cfgMgr)
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: builds without errors

- [ ] **Step 3: Commit**

```bash
git add internal/cli/root.go
git commit -m "feat: pass cfgMgr to TUI model for mode and subscription management"
```

---

### Task 5: Rewrite View with Full Renderers

**Files:**
- Modify: `internal/tui/view.go`

- [ ] **Step 1: Replace view.go with complete implementation**

Replace the entire content of `internal/tui/view.go`:

```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.width == 0 {
		return "加载中..."
	}

	header := m.renderHeader()
	tabBar := m.renderTabBar()
	footer := m.renderFooter()

	var body string
	switch m.tabIdx {
	case 0:
		body = m.renderProxiesTab()
	case 1:
		body = m.renderConnectionsTab()
	case 2:
		body = m.renderLogsTab()
	case 3:
		body = m.renderRulesTab()
	case 4:
		body = m.renderSubsTab()
	}

	mainHeight := m.height - lipgloss.Height(header) - lipgloss.Height(tabBar) - lipgloss.Height(footer) - 2
	if mainHeight < 5 {
		mainHeight = 5
	}
	body = lipgloss.NewStyle().Height(mainHeight).Render(body)

	return lipgloss.JoinVertical(lipgloss.Left, header, tabBar, body, footer)
}

// ─── Header ─────────────────────────────────────────────────

func (m Model) renderHeader() string {
	statusDot := "🔴"
	statusText := StatusStoppedStyle.Render("已停止")
	if m.kernelMgr != nil {
		s := m.kernelMgr.Status()
		switch s {
		case "running":
			statusDot = "🟢"
			statusText = StatusRunningStyle.Render("运行中")
		case "starting":
			statusDot = "🟡"
			statusText = StatusStartingStyle.Render("启动中")
		}
	}

	mode := "未知"
	if m.cfgMgr != nil {
		switch m.cfgMgr.Config().Mode {
		case "rule":
			mode = "规则模式"
		case "global":
			mode = "全局模式"
		case "direct":
			mode = "直连模式"
		case "script":
			mode = "脚本模式"
		}
	}

	traffic := ""
	if m.trafficUp > 0 || m.trafficDown > 0 {
		traffic = fmt.Sprintf("↑ %s  ↓ %s", FormatRate(m.trafficUp), FormatRate(m.trafficDown))
	}

	left := HeaderStyle.Render("mihomo-cli")
	middle := fmt.Sprintf("%s  %s  %s", statusText, MutedStyle.Render(mode), MutedStyle.Render(traffic))

	leftW := lipgloss.Width(left)
	middleW := lipgloss.Width(middle)
	spacing := m.width - leftW - middleW - 2
	if spacing < 0 {
		spacing = 1
	}

	return left + strings.Repeat(" ", spacing) + middle
}

// ─── Tab Bar ────────────────────────────────────────────────

func (m Model) renderTabBar() string {
	tabs := []string{"代理", "连接", "日志", "规则", "订阅"}
	var rendered []string
	for i, t := range tabs {
		label := fmt.Sprintf(" %d.%s ", i+1, t)
		if i == m.tabIdx {
			rendered = append(rendered, TabActiveStyle.Render(label))
		} else {
			rendered = append(rendered, TabInactiveStyle.Render(label))
		}
	}
	separator := MutedStyle.Render("│")
	return lipgloss.JoinHorizontal(lipgloss.Center, rendered...) + " " + separator
}

// ─── Footer / Help Bar ──────────────────────────────────────

func (m Model) renderFooter() string {
	var keys []string
	if m.searchMode {
		keys = append(keys, HelpKeyStyle.Render("输入")+HelpDescStyle.Render(" 搜索节点"))
		keys = append(keys, HelpKeyStyle.Render("Esc")+HelpDescStyle.Render(" 取消"))
	} else {
		switch m.tabIdx {
		case 0: // Proxies
			keys = append(keys, HelpKeyStyle.Render("↑↓/jk")+HelpDescStyle.Render(" 导航"))
			keys = append(keys, HelpKeyStyle.Render("Enter")+HelpDescStyle.Render(" 切换"))
			keys = append(keys, HelpKeyStyle.Render("Tab")+HelpDescStyle.Render(" 换组"))
			keys = append(keys, HelpKeyStyle.Render("t")+HelpDescStyle.Render(" 测速"))
			keys = append(keys, HelpKeyStyle.Render("T")+HelpDescStyle.Render(" 全组测速"))
			keys = append(keys, HelpKeyStyle.Render("/")+HelpDescStyle.Render(" 搜索"))
			keys = append(keys, HelpKeyStyle.Render("m")+HelpDescStyle.Render(" 模式"))
		case 1: // Connections
			keys = append(keys, HelpKeyStyle.Render("↑↓/jk")+HelpDescStyle.Render(" 导航"))
			keys = append(keys, HelpKeyStyle.Render("d")+HelpDescStyle.Render(" 关闭"))
			keys = append(keys, HelpKeyStyle.Render("X")+HelpDescStyle.Render(" 关闭全部"))
		case 2: // Logs
			keys = append(keys, HelpKeyStyle.Render("↑↓/jk")+HelpDescStyle.Render(" 滚动"))
			keys = append(keys, HelpKeyStyle.Render("l")+HelpDescStyle.Render(" 级别"))
			keys = append(keys, HelpKeyStyle.Render("s")+HelpDescStyle.Render(" 自动滚动"))
		case 3: // Rules
			keys = append(keys, HelpKeyStyle.Render("↑↓/jk")+HelpDescStyle.Render(" 滚动"))
		case 4: // Subs
			keys = append(keys, HelpKeyStyle.Render("↑↓/jk")+HelpDescStyle.Render(" 导航"))
			keys = append(keys, HelpKeyStyle.Render("a")+HelpDescStyle.Render(" 添加"))
			keys = append(keys, HelpKeyStyle.Render("d")+HelpDescStyle.Render(" 删除"))
			keys = append(keys, HelpKeyStyle.Render("空格")+HelpDescStyle.Render(" 启停"))
			keys = append(keys, HelpKeyStyle.Render("u")+HelpDescStyle.Render(" 更新"))
		}
		keys = append(keys, HelpKeyStyle.Render("1-5")+HelpDescStyle.Render(" 视图"))
		keys = append(keys, HelpKeyStyle.Render("q")+HelpDescStyle.Render(" 退出"))
	}
	bar := strings.Join(keys, "  ")
	return FooterStyle.Width(m.width).Render(bar)
}

// ─── Tab 1: Proxies ─────────────────────────────────────────

func (m Model) renderProxiesTab() string {
	panelW := m.width - 4
	var b strings.Builder

	if m.searchMode {
		b.WriteString(SearchStyle.Width(panelW).Render(fmt.Sprintf(" 搜索: %s▎", m.searchQuery)))
		b.WriteString("\n\n")
	}

	if m.proxiesErr != nil {
		b.WriteString(MutedStyle.Render(fmt.Sprintf(" 错误: %v", m.proxiesErr)))
		return PanelStyle.Width(panelW).Render(b.String())
	}

	if len(m.groups) == 0 {
		b.WriteString(MutedStyle.Render(" 暂无代理组"))
		return PanelStyle.Width(panelW).Render(b.String())
	}

	for gi, groupName := range m.groups {
		group, ok := m.proxies.Proxies[groupName]
		if !ok {
			continue
		}

		// Group header
		collapsed := m.collapsed[groupName]
		arrow := "▼"
		if collapsed {
			arrow = "▶"
		}
		headerPrefix := " "
		if gi == m.groupIdx {
			headerPrefix = "▸"
		}
		groupTitle := fmt.Sprintf("%s%s [%s]  节点:%d  已选:%s",
			headerPrefix, arrow, groupName, len(group.All), MutedStyle.Render(group.Now))
		b.WriteString(GroupHeaderStyle.Render(groupTitle))
		b.WriteString("\n")

		if collapsed {
			continue
		}

		// Find max delay in group for bar scaling
		maxDelay := 0
		for _, node := range group.All {
			if d, ok := m.delayResults[node]; ok && d > maxDelay {
				maxDelay = d
			}
		}

		// Apply search filter
		var nodes []string
		if m.searchMode && m.searchQuery != "" {
			for _, node := range group.All {
				if strings.Contains(strings.ToLower(node), strings.ToLower(m.searchQuery)) {
					nodes = append(nodes, node)
				}
			}
		} else {
			nodes = group.All
		}

		for ni, node := range nodes {
			marker := "○"
			if node == group.Now {
				marker = "●"
			}

			// Latency display
			latStr := "  -ms"
			if d, ok := m.delayResults[node]; ok && d > 0 {
				latStr = LatencyColor(d).Styled(fmt.Sprintf("%4dms", d))
			}

			// Traffic bar (simplified)
			bar := LatencyBar(m.delayResults[node], maxDelay, 6)

			line := fmt.Sprintf("  %s %s  %s  %s", marker, node, latStr, bar)

			isSelected := gi == m.groupIdx
			// Find if this node is at current cursor
			isCursor := false
			if isSelected {
				originalIdx := -1
				for i, n := range group.All {
					if n == node {
						originalIdx = i
						break
					}
				}
				if originalIdx == m.nodeIdx {
					isCursor = true
				}
			}

			_ = ni // silence unused

			if isCursor {
				b.WriteString(SelectedStyle.Render(line))
			} else {
				b.WriteString(NormalStyle.Render(line))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return PanelStyle.Width(panelW).Render(b.String())
}

// ─── Tab 2: Connections ─────────────────────────────────────

func (m Model) renderConnectionsTab() string {
	panelW := m.width - 4
	var b strings.Builder

	if m.connectionsErr != nil {
		b.WriteString(MutedStyle.Render(fmt.Sprintf(" 错误: %v", m.connectionsErr)))
		return PanelStyle.Width(panelW).Render(b.String())
	}

	b.WriteString(MutedStyle.Render(fmt.Sprintf(" 协议  目标主机                    代理链                 上行      下行\n")))
	b.WriteString(MutedStyle.Render(" " + strings.Repeat("─", panelW-2)))
	b.WriteString("\n")

	if len(m.connections) == 0 {
		b.WriteString(MutedStyle.Render("\n 暂无活跃连接"))
	} else {
		for i, conn := range m.connections {
			host := conn.Metadata.Host
			if len(host) > 28 {
				host = host[:27] + "…"
			}
			chain := strings.Join(conn.Chains, " → ")
			if len(chain) > 24 {
				chain = chain[:23] + "…"
			}
			up := FormatBytes(conn.Upload)
			down := FormatBytes(conn.Download)

			line := fmt.Sprintf(" %-4s  %-28s  %-24s  %6s  %6s",
				conn.Metadata.Network, host, chain, up, down)

			if i == m.connIdx {
				b.WriteString(SelectedStyle.Render(line))
			} else {
				b.WriteString(NormalStyle.Render(line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(fmt.Sprintf(" 共 %d 条连接", len(m.connections))))

	return PanelStyle.Width(panelW).Render(b.String())
}

// ─── Tab 3: Logs ────────────────────────────────────────────

func (m Model) renderLogsTab() string {
	panelW := m.width - 4
	var b strings.Builder

	levelLabel := "全部"
	switch m.logLevel {
	case "info":
		levelLabel = "INFO"
	case "warning":
		levelLabel = "WARNING"
	case "error":
		levelLabel = "ERROR"
	case "debug":
		levelLabel = "DEBUG"
	}
	autoLabel := "关"
	if m.autoScroll {
		autoLabel = "开"
	}
	b.WriteString(MutedStyle.Render(fmt.Sprintf(" 级别:%s  自动滚动:%s", levelLabel, autoLabel)))
	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(" " + strings.Repeat("─", panelW-2)))
	b.WriteString("\n")

	if m.logsErr != nil {
		b.WriteString(MutedStyle.Render(fmt.Sprintf(" 错误: %v", m.logsErr)))
		return PanelStyle.Width(panelW).Render(b.String())
	}

	if len(m.logs) == 0 {
		b.WriteString(MutedStyle.Render("\n 暂无日志"))
	} else {
		for _, entry := range m.logs {
			if m.logLevel != "" && entry.Type != m.logLevel {
				continue
			}
			var style lipgloss.Style
			switch entry.Type {
			case "info":
				style = LogInfoStyle
			case "warning":
				style = LogWarnStyle
			case "error":
				style = LogErrorStyle
			case "debug":
				style = LogDebugStyle
			default:
				style = NormalStyle
			}
			b.WriteString(style.Render(fmt.Sprintf(" %s", entry.Payload)))
			b.WriteString("\n")
		}
	}

	return PanelStyle.Width(panelW).Render(b.String())
}

// ─── Tab 4: Rules ───────────────────────────────────────────

func (m Model) renderRulesTab() string {
	panelW := m.width - 4
	var b strings.Builder

	if m.rulesErr != nil {
		b.WriteString(MutedStyle.Render(fmt.Sprintf(" 错误: %v", m.rulesErr)))
		return PanelStyle.Width(panelW).Render(b.String())
	}

	b.WriteString(MutedStyle.Render(fmt.Sprintf(" 类型        匹配内容                      代理目标\n")))
	b.WriteString(MutedStyle.Render(" " + strings.Repeat("─", panelW-2)))
	b.WriteString("\n")

	if len(m.rules) == 0 {
		b.WriteString(MutedStyle.Render("\n 暂无规则"))
	} else {
		for i, rule := range m.rules {
			payload := rule.Payload
			if len(payload) > 30 {
				payload = payload[:29] + "…"
			}
			proxy := rule.Proxy
			if len(proxy) > 20 {
				proxy = proxy[:19] + "…"
			}
			line := fmt.Sprintf(" %-12s  %-30s  %s", rule.Type, payload, proxy)
			if i == m.ruleIdx {
				b.WriteString(SelectedStyle.Render(line))
			} else {
				b.WriteString(NormalStyle.Render(line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(fmt.Sprintf(" 共 %d 条规则", len(m.rules))))

	return PanelStyle.Width(panelW).Render(b.String())
}

// ─── Tab 5: Subscriptions ───────────────────────────────────

func (m Model) renderSubsTab() string {
	panelW := m.width - 4
	var b strings.Builder

	var subs []struct {
		Name        string
		URL         string
		Enabled     bool
		LastUpdated int64
		Interval    int
	}
	if m.cfgMgr != nil {
		for _, s := range m.cfgMgr.Config().Subscriptions {
			subs = append(subs, struct {
				Name        string
				URL         string
				Enabled     bool
				LastUpdated int64
				Interval    int
			}{s.Name, s.URL, true, s.LastUpdated, s.Interval})
		}
	}

	if len(subs) == 0 {
		b.WriteString(MutedStyle.Render(" 暂无订阅，按 a 添加"))
		return PanelStyle.Width(panelW).Render(b.String())
	}

	for i, s := range subs {
		status := StatusRunningStyle.Render("● 启用")
		if !s.Enabled {
			status = StatusStoppedStyle.Render("○ 停用")
		}

		updated := "从未更新"
		if s.LastUpdated > 0 {
			updated = FormatDuration(time.Now().Unix() - s.LastUpdated) + "前"
		}

		urlDisplay := s.URL
		if len(urlDisplay) > 40 {
			urlDisplay = urlDisplay[:39] + "…"
		}

		line := fmt.Sprintf(" %s  %-12s  %s  更新间隔:%dh  %s",
			status, s.Name, urlDisplay, s.Interval/3600, MutedStyle.Render(updated))

		if i == m.subIdx {
			b.WriteString(SelectedStyle.Render(line))
		} else {
			b.WriteString(NormalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	return PanelStyle.Width(panelW).Render(b.String())
}
```

- [ ] **Step 2: Add missing import**

The file needs `time` import for `FormatDuration`. Let's verify the import block includes it.

The imports used in view.go are: `fmt`, `strings`, `time`, `lipgloss`. The `time` import is needed for `FormatDuration(time.Now().Unix() - s.LastUpdated)`.

Wait, looking at the code again, `time.Now()` requires `time` import. Let me fix the subs renderer to not use `time.Now()` directly but instead use a simpler approach.

Actually I'll just make sure the import block has `time`. Let me check - the renderSubsTab function uses `time.Now()`, and `time` is not yet imported in view.go. We need to add it.

Let me update the view.go imports to include `time`.

- [ ] **Step 2 (revised): Verify compilation with correct imports**

The imports at the top of view.go should be:
```go
import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)
```

Run: `go build ./...`
Expected: builds without errors

- [ ] **Step 3: Commit**

```bash
git add internal/tui/view.go
git commit -m "feat: tab-based TUI views with header, footer, and all 5 tab renderers"
```

---

### Task 6: Rewrite Update Logic

**Files:**
- Modify: `internal/tui/update.go`

- [ ] **Step 1: Replace update.go with complete implementation**

Replace the entire content of `internal/tui/update.go`:

```go
package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Search mode: capture input
		if m.searchMode {
			return m.handleSearchInput(msg)
		}

		// Global keys (all tabs)
		switch {
		case key.Matches(msg, Keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, Keys.Mode):
			m.cycleMode()
			return m, nil

		case key.Matches(msg, Keys.Reload):
			if m.apiClient != nil {
				m.apiClient.ReloadConfig()
				m.notification = "配置已重载"
			}
			return m, nil

		case key.Matches(msg, Keys.Update):
			if m.subMgr != nil {
				m.subMgr.UpdateAll()
				m.notification = "订阅更新已触发"
			}
			return m, nil
		}

		// Tab switching (1-5)
		if len(msg.String()) == 1 {
			ch := msg.String()[0]
			if ch >= '1' && ch <= '5' {
				m.tabIdx = int(ch - '1')
				m.nodeIdx = 0
				m.connIdx = 0
				// Fetch data for newly selected tab
				return m, m.fetchTabData()
			}
		}

		// Per-tab keys
		switch m.tabIdx {
		case 0:
			return m.handleProxiesKeys(msg)
		case 1:
			return m.handleConnectionsKeys(msg)
		case 2:
			return m.handleLogsKeys(msg)
		case 3:
			return m.handleRulesKeys(msg)
		case 4:
			return m.handleSubsKeys(msg)
		}

	case tickMsg:
		return m, m.handleTick()

	case proxiesMsg:
		return m.handleProxiesMsg(msg)

	case connectionsMsg:
		return m.handleConnectionsMsg(msg)

	case logsMsg:
		return m.handleLogsMsg(msg)

	case rulesMsg:
		return m.handleRulesMsg(msg)

	case delayResultMsg:
		return m.handleDelayResult(msg)

	case speedtestDoneMsg:
		return m.handleSpeedtestDone(msg)

	case notificationMsg:
		m.notification = string(msg)
		return m, nil
	}

	return m, nil
}

// ─── Search Input ───────────────────────────────────────────

func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		return m, nil
	case "enter":
		m.searchMode = false
		return m, nil
	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
		}
		return m, nil
	}
}

// ─── Tab Handlers ───────────────────────────────────────────

func (m Model) handleProxiesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, Keys.Up):
		if m.nodeIdx > 0 {
			m.nodeIdx--
		}

	case key.Matches(msg, Keys.Down):
		group := m.currentGroup()
		if group != nil && m.nodeIdx < len(group.All)-1 {
			m.nodeIdx++
		}

	case key.Matches(msg, Keys.TabNext):
		if len(m.groups) > 0 {
			m.groupIdx = (m.groupIdx + 1) % len(m.groups)
			m.nodeIdx = 0
		}

	case key.Matches(msg, Keys.TabPrev):
		if len(m.groups) > 0 {
			m.groupIdx--
			if m.groupIdx < 0 {
				m.groupIdx = len(m.groups) - 1
			}
			m.nodeIdx = 0
		}

	case key.Matches(msg, Keys.Enter):
		m.switchToSelected()

	case key.Matches(msg, Keys.Test):
		return m, m.testSelectedNode()

	case key.Matches(msg, Keys.TestAll):
		return m, m.speedtestCurrentGroup()

	case key.Matches(msg, Keys.Search):
		m.searchMode = true
		m.searchQuery = ""
		return m, nil

	case key.Matches(msg, Keys.Collapse):
		group := m.currentGroup()
		if group != nil {
			m.collapsed[group.Name] = !m.collapsed[group.Name]
		}
	}
	return m, nil
}

func (m Model) handleConnectionsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, Keys.Up):
		if m.connIdx > 0 {
			m.connIdx--
		}

	case key.Matches(msg, Keys.Down):
		if m.connIdx < len(m.connections)-1 {
			m.connIdx++
		}

	case key.Matches(msg, Keys.Close):
		if m.connIdx < len(m.connections) && m.apiClient != nil {
			m.apiClient.CloseConnection(m.connections[m.connIdx].ID)
			m.notification = fmt.Sprintf("已关闭连接 %s", m.connections[m.connIdx].ID)
		}

	case key.Matches(msg, Keys.CloseAll):
		if m.apiClient != nil {
			for _, conn := range m.connections {
				m.apiClient.CloseConnection(conn.ID)
			}
			m.notification = "已关闭全部连接"
		}
	}
	return m, nil
}

func (m Model) handleLogsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, Keys.Up):
		if m.logScroll > 0 {
			m.logScroll--
		}

	case key.Matches(msg, Keys.Down):
		m.logScroll++

	case key.Matches(msg, Keys.LogLevel):
		switch m.logLevel {
		case "":
			m.logLevel = "info"
		case "info":
			m.logLevel = "warning"
		case "warning":
			m.logLevel = "error"
		case "error":
			m.logLevel = "debug"
		case "debug":
			m.logLevel = ""
		}

	case msg.String() == "s":
		m.autoScroll = !m.autoScroll
	}
	return m, nil
}

func (m Model) handleRulesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, Keys.Up):
		if m.ruleIdx > 0 {
			m.ruleIdx--
		}
	case key.Matches(msg, Keys.Down):
		if m.ruleIdx < len(m.rules)-1 {
			m.ruleIdx++
		}
	}
	return m, nil
}

func (m Model) handleSubsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, Keys.Up):
		if m.subIdx > 0 {
			m.subIdx--
		}

	case key.Matches(msg, Keys.Down):
		if m.subIdx < len(m.cfgMgr.Config().Subscriptions)-1 {
			m.subIdx++
		}

	case key.Matches(msg, Keys.SubDel):
		if m.cfgMgr != nil {
			subs := m.cfgMgr.Config().Subscriptions
			if m.subIdx < len(subs) {
				m.cfgMgr.RemoveSubscription(subs[m.subIdx].Name)
				m.notification = fmt.Sprintf("已删除订阅 %s", subs[m.subIdx].Name)
				if m.subIdx >= len(subs)-1 && m.subIdx > 0 {
					m.subIdx--
				}
			}
		}

	case key.Matches(msg, Keys.Update):
		if m.subMgr != nil {
			m.subMgr.UpdateAll()
			m.notification = "正在更新所有订阅..."
		}
	}
	return m, nil
}

// ─── Tick ───────────────────────────────────────────────────

func (m Model) handleTick() tea.Cmd {
	var cmds []tea.Cmd

	// Re-fetch data based on active tab every 2s
	if time.Since(m.lastUpdate) > 2*time.Second {
		m.lastUpdate = time.Now()
		cmds = append(cmds, m.fetchTabData())
	}

	// Always tick
	cmds = append(cmds, func() tea.Msg {
		time.Sleep(1 * time.Second)
		return tickMsg(time.Now())
	})

	return tea.Batch(cmds...)
}

func (m Model) fetchTabData() tea.Cmd {
	switch m.tabIdx {
	case 0:
		return func() tea.Msg { return fetchProxies(m.apiClient) }
	case 1:
		return func() tea.Msg { return fetchConnections(m.apiClient) }
	case 2:
		return func() tea.Msg { return fetchLogs(m.apiClient) }
	case 3:
		return func() tea.Msg { return fetchRules(m.apiClient) }
	default:
		return nil
	}
}

// ─── Message Handlers ───────────────────────────────────────

func (m Model) handleProxiesMsg(msg proxiesMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.proxiesErr = msg.err
	} else {
		m.proxies = msg.proxies
		m.proxiesErr = nil
		m.refreshGroups()
	}
	return m, nil
}

func (m Model) handleConnectionsMsg(msg connectionsMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.connectionsErr = msg.err
	} else {
		m.connections = msg.connections
		m.connectionsErr = nil
		// Calculate traffic rates
		var up, down int64
		for _, c := range msg.connections {
			up += c.Upload
			down += c.Download
		}
		m.trafficUp = up
		m.trafficDown = down
		if m.connIdx >= len(m.connections) && len(m.connections) > 0 {
			m.connIdx = len(m.connections) - 1
		}
	}
	return m, nil
}

func (m Model) handleLogsMsg(msg logsMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.logsErr = msg.err
	} else {
		m.logs = msg.logs
		m.logsErr = nil
	}
	return m, nil
}

func (m Model) handleRulesMsg(msg rulesMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.rulesErr = msg.err
	} else {
		m.rules = msg.rules
		m.rulesErr = nil
		if m.ruleIdx >= len(m.rules) && len(m.rules) > 0 {
			m.ruleIdx = len(m.rules) - 1
		}
	}
	return m, nil
}

func (m Model) handleDelayResult(msg delayResultMsg) (tea.Model, tea.Cmd) {
	if msg.err == nil && msg.delay > 0 {
		m.delayResults[msg.node] = msg.delay
		m.notification = fmt.Sprintf("%s 延迟: %dms", msg.node, msg.delay)
	} else if msg.err != nil {
		m.notification = fmt.Sprintf("%s 测速失败: %v", msg.node, msg.err)
	}
	return m, nil
}

func (m Model) handleSpeedtestDone(msg speedtestDoneMsg) (tea.Model, tea.Cmd) {
	for node, delay := range msg.results {
		m.delayResults[node] = delay
	}
	m.speedtesting = false
	m.notification = fmt.Sprintf("[%s] 全组测速完成 (%d 个节点)", msg.group, len(msg.results))
	return m, nil
}

// ─── Actions ────────────────────────────────────────────────

func (m Model) testSelectedNode() tea.Cmd {
	group := m.currentGroup()
	if group == nil || m.nodeIdx >= len(group.All) {
		return nil
	}
	node := group.All[m.nodeIdx]
	return func() tea.Msg {
		d, err := m.apiClient.TestDelay(node, 5*time.Second)
		return delayResultMsg{node: node, delay: d, err: err}
	}
}

func (m Model) speedtestCurrentGroup() tea.Cmd {
	group := m.currentGroup()
	if group == nil || m.speedtesting {
		return nil
	}
	m.speedtesting = true
	groupName := group.Name
	nodes := make([]string, len(group.All))
	copy(nodes, group.All)

	return func() tea.Msg {
		results := make(map[string]int)
		for _, node := range nodes {
			d, err := m.apiClient.TestDelay(node, 5*time.Second)
			if err == nil && d > 0 {
				results[node] = d
			}
		}
		return speedtestDoneMsg{results: results, group: groupName}
	}
}

func (m *Model) cycleMode() {
	if m.cfgMgr == nil {
		return
	}
	modes := []string{"rule", "global", "direct"}
	cur := m.cfgMgr.Config().Mode
	for i, mode := range modes {
		if mode == cur {
			next := modes[(i+1)%len(modes)]
			m.cfgMgr.SetMode(next)
			if m.apiClient != nil {
				m.apiClient.ReloadConfig()
			}
			m.notification = fmt.Sprintf("已切换到 %s 模式", next)
			return
		}
	}
	// Current mode not in list, default to rule
	m.cfgMgr.SetMode("rule")
	m.notification = "已切换到 rule 模式"
}

func (m *Model) switchToSelected() {
	group := m.currentGroup()
	if group == nil || m.nodeIdx >= len(group.All) {
		return
	}
	node := group.All[m.nodeIdx]
	m.apiClient.SwitchProxy(group.Name, node)
	m.notification = fmt.Sprintf("已切换 [%s] → %s", group.Name, node)
}

// ─── Helpers ────────────────────────────────────────────────

func (m *Model) currentGroup() *api.Proxy {
	if m.proxies == nil || len(m.groups) == 0 {
		return nil
	}
	if m.groupIdx >= len(m.groups) {
		m.groupIdx = 0
	}
	name := m.groups[m.groupIdx]
	p, ok := m.proxies.Proxies[name]
	if !ok {
		return nil
	}
	return &p
}

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
	if m.groupIdx >= len(m.groups) {
		m.groupIdx = 0
	}
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./...`
Expected: builds without errors

- [ ] **Step 3: Commit**

```bash
git add internal/tui/update.go
git commit -m "feat: complete TUI update logic with tab routing, search, speedtest, and mode cycling"
```

---

### Task 7: Integration Verification and Fixes

**Files:**
- No new files; verify and fix all changed files compile and work together

- [ ] **Step 1: Full build**

Run: `go build ./...`
Expected: clean build, no errors or warnings

- [ ] **Step 2: Run go vet**

Run: `go vet ./...`
Expected: no issues

- [ ] **Step 3: Fix any compilation issues**

Common issues to check:
- `time` import in `view.go` (needed for `renderSubsTab`'s `time.Now()`)
- `api` import no longer needed in `view.go` (proxy types accessed through `m.proxies`)
- `fmt` import no longer needed in `update.go` if we use `fmt.Sprintf` in notification (we do)
- Verify `update.go` imports: `fmt`, `time`, `tea`, `key`, and `api` (if still referenced)

Let's verify the imports needed for each file:

**view.go needs:** `fmt`, `strings`, `time`, `lipgloss`
**update.go needs:** `fmt`, `time`, `tea`, `key`

The `api` import is no longer needed in `update.go` since we use `m.currentGroup()` which returns `*api.Proxy` through the Model type. Actually, `currentGroup()` returns `*api.Proxy` but we don't explicitly reference `api.Proxy` type in update.go — it's all through the Model. So `api` import can be removed from update.go.

Wait, actually in the rewritten update.go, we don't reference `api` directly. The `currentGroup()` returns `*api.Proxy` but that's in the method signature which is defined on the Model. So the import can be removed.

Let me double check — in the speedtest function:
```go
func (m Model) speedtestCurrentGroup() tea.Cmd {
	group := m.currentGroup()
	...
	return func() tea.Msg {
		results := make(map[string]int)
		for _, node := range nodes {
			d, err := m.apiClient.TestDelay(node, 5*time.Second)
```

The `m.apiClient` is accessed through the Model, not through the `api` package directly. So no `api` import needed in update.go.

And in view.go, `m.currentGroup()` returns `*api.Proxy` implicitly — Go doesn't need the import for the return type to work as long as it's used through the Model. But actually, the renderProxiesTab accesses `m.proxies.Proxies[groupName]` which returns `api.Proxy` — but again, no explicit reference to the `api` type in the view code.

So both `view.go` and `update.go` should NOT need the `api` import. Let me remove them from the plan.

- [ ] **Step 4: Fix view.go imports**

Ensure `internal/tui/view.go` has exactly these imports:
```go
import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)
```

Ensure `internal/tui/update.go` has exactly these imports:
```go
import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)
```

- [ ] **Step 5: Final build verification**

Run: `go build ./... && go vet ./...`
Expected: clean, no errors

- [ ] **Step 6: Commit any fixes**

```bash
git add -A && git diff --cached --stat
# Only commit if there are changes
git commit -m "fix: resolve compilation issues in TUI rewrite"
```
