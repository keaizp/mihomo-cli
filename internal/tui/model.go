package tui

import (
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

	proxies    *api.ProxiesResponse
	proxiesErr error

	groups     []string
	groupIdx   int
	nodeIdx    int
	width      int
	height     int
	status     string
	lastUpdate time.Time
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
	Up     key.Binding
	Down   key.Binding
	Enter  key.Binding
	Tab    key.Binding
	Test   key.Binding
	Quit   key.Binding
	Reload key.Binding
	Update key.Binding
}

var Keys = KeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "up")),
	Down:   key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "down")),
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
