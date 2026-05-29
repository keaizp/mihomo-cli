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
	delayResults map[string]int
	collapsed    map[string]bool
	speedtesting bool

	// Connections tab
	connections    []api.Connection
	connectionsErr error
	connIdx        int

	// Logs tab
	logs       []api.LogEntry
	logsErr    error
	logLevel   string
	logScroll  int
	autoScroll bool

	// Rules tab
	rules    []api.Rule
	rulesErr error
	ruleIdx  int

	// Subs tab
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

	// Notification
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
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	TabPrev   key.Binding
	TabNext   key.Binding
	Test      key.Binding
	TestAll   key.Binding
	Quit      key.Binding
	Reload    key.Binding
	Update    key.Binding
	Mode      key.Binding
	Search    key.Binding
	Collapse  key.Binding
	Close     key.Binding
	CloseAll  key.Binding
	LogLevel  key.Binding
	SubAdd    key.Binding
	SubEdit   key.Binding
	SubDel    key.Binding
	SubToggle key.Binding
}

var Keys = KeyMap{
	Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "上移")),
	Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "下移")),
	Enter:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "选择")),
	TabPrev:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("S-Tab", "上一组")),
	TabNext:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "下一组")),
	Test:      key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "测速")),
	TestAll:   key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "全组测速")),
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"), key.WithHelp("q", "退出")),
	Reload:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "重载配置")),
	Update:    key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "更新订阅")),
	Mode:      key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "切换模式")),
	Search:    key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "搜索")),
	Collapse:  key.NewBinding(key.WithKeys("space"), key.WithHelp("空格", "折叠")),
	Close:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "关闭连接")),
	CloseAll:  key.NewBinding(key.WithKeys("X"), key.WithHelp("X", "关闭全部")),
	LogLevel:  key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "日志级别")),
	SubAdd:    key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "添加订阅")),
	SubEdit:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "编辑订阅")),
	SubDel:    key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "删除订阅")),
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
