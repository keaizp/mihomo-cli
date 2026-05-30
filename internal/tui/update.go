package tui

import (
	"fmt"
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
				m.notification = "正在更新所有订阅..."
			}
			return m, nil
		}

		// Tab switching (1-5)
		s := msg.String()
		if len(s) == 1 {
			ch := s[0]
			if ch >= '1' && ch <= '5' {
				m.tabIdx = int(ch - '1')
				m.nodeIdx = 0
				m.connIdx = 0
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
		if len(m.groups) > 0 {
			name := m.groups[m.groupIdx]
			m.collapsed[name] = !m.collapsed[name]
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
			id := m.connections[m.connIdx].ID
			m.apiClient.CloseConnection(id)
			m.notification = fmt.Sprintf("已关闭连接 %s", id)
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
		if m.logScroll < len(m.logs)-1 {
			m.logScroll++
		}

	case key.Matches(msg, Keys.Down):
		if m.logScroll > 0 {
			m.logScroll--
		}

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
	subCount := 0
	if m.cfgMgr != nil {
		subCount = len(m.cfgMgr.Config().Subscriptions)
	}

	switch {
	case key.Matches(msg, Keys.Up):
		if m.subIdx > 0 {
			m.subIdx--
		}

	case key.Matches(msg, Keys.Down):
		if m.subIdx < subCount-1 {
			m.subIdx++
		}

	case key.Matches(msg, Keys.Enter):
		// Set selected subscription as active
		if m.cfgMgr != nil && subCount > 0 {
			subs := m.cfgMgr.Config().Subscriptions
			if m.subIdx < len(subs) {
				name := subs[m.subIdx].Name
				cur := m.cfgMgr.Config().ActiveSubscription
				if cur == name {
					// Toggle off — use all subscriptions
					m.cfgMgr.SetActiveSubscription("")
					m.subMgr.MergeAndGenerate()
					m.notification = "已切换为使用全部订阅"
				} else {
					m.cfgMgr.SetActiveSubscription(name)
					m.subMgr.MergeAndGenerate()
					m.notification = fmt.Sprintf("已切换激活订阅: %s", name)
				}
				if m.apiClient != nil {
					m.apiClient.ReloadConfig()
				}
			}
		}

	case key.Matches(msg, Keys.SubAdd):
		m.notification = "请使用命令行添加: mihomo-cli sub add <名称> <URL>"

	case key.Matches(msg, Keys.SubEdit):
		if m.cfgMgr != nil && subCount > 0 {
			subs := m.cfgMgr.Config().Subscriptions
			if m.subIdx < len(subs) {
				m.notification = fmt.Sprintf("请使用命令行编辑: mihomo-cli sub edit %s <新URL>", subs[m.subIdx].Name)
			}
		}

	case key.Matches(msg, Keys.SubDel):
		if m.cfgMgr != nil && subCount > 0 {
			subs := m.cfgMgr.Config().Subscriptions
			if m.subIdx < len(subs) {
				name := subs[m.subIdx].Name
				m.cfgMgr.RemoveSubscription(name)
				// If we removed the active sub, reset
				if m.cfgMgr.Config().ActiveSubscription == name {
					m.cfgMgr.SetActiveSubscription("")
				}
				m.subMgr.MergeAndGenerate()
				if m.apiClient != nil {
					m.apiClient.ReloadConfig()
				}
				m.notification = fmt.Sprintf("已删除订阅 %s", name)
				if m.subIdx >= len(subs)-1 && m.subIdx > 0 {
					m.subIdx--
				}
			}
		}

	case key.Matches(msg, Keys.Update):
		// Update single selected subscription
		if m.subMgr != nil && subCount > 0 {
			subs := m.cfgMgr.Config().Subscriptions
			if m.subIdx < len(subs) {
				name := subs[m.subIdx].Name
				if err := m.subMgr.UpdateSubscription(name); err != nil {
					m.notification = fmt.Sprintf("更新失败: %v", err)
				} else {
					m.notification = fmt.Sprintf("已更新订阅: %s", name)
				}
				if m.apiClient != nil {
					m.apiClient.ReloadConfig()
				}
			}
		}
	}
	return m, nil
}

// ─── Tick ───────────────────────────────────────────────────

func (m Model) handleTick() tea.Cmd {
	var cmds []tea.Cmd

	if time.Since(m.lastUpdate) > 2*time.Second {
		m.lastUpdate = time.Now()
		cmds = append(cmds, m.fetchTabData())
	}

	cmds = append(cmds, func() tea.Msg {
		time.Sleep(1 * time.Second)
		return tickMsg(time.Now())
	})

	return tea.Batch(cmds...)
}

func (m Model) fetchTabData() tea.Cmd {
	if m.apiClient == nil {
		return nil
	}
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
		m.notification = fmt.Sprintf("%s 测速失败", msg.node)
	}
	return m, nil
}

func (m Model) handleSpeedtestDone(msg speedtestDoneMsg) (tea.Model, tea.Cmd) {
	for node, delay := range msg.results {
		m.delayResults[node] = delay
	}
	m.speedtesting = false
	m.notification = fmt.Sprintf("[%s] 全组测速完成 (%d 节点)", msg.group, len(msg.results))
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
	m.cfgMgr.SetMode("rule")
	if m.apiClient != nil {
		m.apiClient.ReloadConfig()
	}
	m.notification = "已切换到规则模式"
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
