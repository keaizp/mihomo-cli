package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
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
		// Delay result stored for display
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
	if m.groupIdx >= len(m.groups) {
		m.groupIdx = 0
	}
}
