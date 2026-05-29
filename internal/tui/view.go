package tui

import (
	"fmt"
	"strings"
	"time"

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

	mainHeight := m.height - lipgloss.Height(header) - lipgloss.Height(tabBar) - lipgloss.Height(footer)
	if mainHeight < 5 {
		mainHeight = 5
	}
	body = lipgloss.NewStyle().Height(mainHeight).Render(body)

	return lipgloss.JoinVertical(lipgloss.Left, header, tabBar, body, footer)
}

// ─── Header ─────────────────────────────────────────────────

func (m Model) renderHeader() string {
	statusText := StatusStoppedStyle.Render("已停止")
	if m.kernelMgr != nil {
		switch m.kernelMgr.Status() {
		case "running":
			statusText = StatusRunningStyle.Render("● 运行中")
		case "starting":
			statusText = StatusStartingStyle.Render("● 启动中")
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

	notif := ""
	if m.notification != "" {
		notif = MutedStyle.Render(" | " + m.notification)
	}

	left := HeaderStyle.Render("mihomo-cli")
	middle := fmt.Sprintf("%s  %s  %s%s",
		statusText, MutedStyle.Render(mode), MutedStyle.Render(traffic), notif)

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
			keys = append(keys, HelpKeyStyle.Render("空格")+HelpDescStyle.Render(" 折叠"))
		case 1: // Connections
			keys = append(keys, HelpKeyStyle.Render("↑↓/jk")+HelpDescStyle.Render(" 导航"))
			keys = append(keys, HelpKeyStyle.Render("d")+HelpDescStyle.Render(" 关闭连接"))
			keys = append(keys, HelpKeyStyle.Render("X")+HelpDescStyle.Render(" 关闭全部"))
		case 2: // Logs
			keys = append(keys, HelpKeyStyle.Render("↑↓/jk")+HelpDescStyle.Render(" 滚动"))
			keys = append(keys, HelpKeyStyle.Render("l")+HelpDescStyle.Render(" 切换级别"))
			keys = append(keys, HelpKeyStyle.Render("s")+HelpDescStyle.Render(" 自动滚动"))
		case 3: // Rules
			keys = append(keys, HelpKeyStyle.Render("↑↓/jk")+HelpDescStyle.Render(" 滚动"))
		case 4: // Subs
			keys = append(keys, HelpKeyStyle.Render("↑↓/jk")+HelpDescStyle.Render(" 导航"))
			keys = append(keys, HelpKeyStyle.Render("a")+HelpDescStyle.Render(" 添加"))
			keys = append(keys, HelpKeyStyle.Render("d")+HelpDescStyle.Render(" 删除"))
			keys = append(keys, HelpKeyStyle.Render("u")+HelpDescStyle.Render(" 更新"))
		}
		keys = append(keys, HelpKeyStyle.Render("1-5")+HelpDescStyle.Render(" 视图"))
		keys = append(keys, HelpKeyStyle.Render("m")+HelpDescStyle.Render(" 模式"))
		keys = append(keys, HelpKeyStyle.Render("r")+HelpDescStyle.Render(" 重载"))
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

		collapsed := m.collapsed[groupName]
		arrow := "▼"
		if collapsed {
			arrow = "▶"
		}
		isActiveGroup := gi == m.groupIdx

		groupTitle := fmt.Sprintf(" %s [%s]  节点:%d  已选:%s",
			arrow, groupName, len(group.All), MutedStyle.Render(group.Now))
		if isActiveGroup {
			b.WriteString(SelectedStyle.Render(groupTitle))
		} else {
			b.WriteString(GroupHeaderStyle.Render(groupTitle))
		}
		b.WriteString("\n")

		if collapsed {
			continue
		}

		// Only show nodes for active group (ccswitch-style: one group at a time)
		if !isActiveGroup {
			continue
		}

		// Find max delay for bar scaling
		maxDelay := 0
		for _, node := range group.All {
			if d, ok := m.delayResults[node]; ok && d > maxDelay {
				maxDelay = d
			}
		}
		if maxDelay < 50 {
			maxDelay = 500 // default scale
		}

		// Apply search filter
		nodes := group.All
		if m.searchMode && m.searchQuery != "" {
			filtered := make([]string, 0)
			for _, node := range group.All {
				if strings.Contains(strings.ToLower(node), strings.ToLower(m.searchQuery)) {
					filtered = append(filtered, node)
				}
			}
			nodes = filtered
		}

		for ni, node := range nodes {
			marker := "○"
			if node == group.Now {
				marker = "●"
			}

			latStr := "   -ms"
			if d, ok := m.delayResults[node]; ok && d > 0 {
				latStr = lipgloss.NewStyle().Foreground(LatencyColor(d)).Render(fmt.Sprintf("%4dms", d))
			}

			bar := LatencyBar(m.delayResults[node], maxDelay, 6)

			line := fmt.Sprintf("  %s %-20s %s  %s", marker, node, latStr, bar)

			if ni == m.nodeIdx {
				b.WriteString(SelectedStyle.Render(line))
			} else {
				b.WriteString(NormalStyle.Render(line))
			}
			b.WriteString("\n")
		}

		if m.speedtesting {
			b.WriteString(MutedStyle.Render("  ⏳ 正在测速..."))
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

	b.WriteString(MutedStyle.Render(fmt.Sprintf(" 协议  目标主机                          代理链                       上行      下行\n")))
	b.WriteString(MutedStyle.Render(" " + strings.Repeat("─", panelW-1)))
	b.WriteString("\n")

	if len(m.connections) == 0 {
		b.WriteString(MutedStyle.Render("\n 暂无活跃连接"))
	} else {
		for i, conn := range m.connections {
			host := conn.Metadata.Host
			if len(host) > 32 {
				host = host[:31] + "…"
			}
			chain := strings.Join(conn.Chains, " → ")
			if len(chain) > 26 {
				chain = chain[:25] + "…"
			}
			up := FormatBytes(conn.Upload)
			down := FormatBytes(conn.Download)

			line := fmt.Sprintf(" %-4s  %-32s  %-26s  %6s  %6s",
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
	b.WriteString(MutedStyle.Render(fmt.Sprintf(" 级别:%s  自动滚动:%s  共 %d 条", levelLabel, autoLabel, len(m.logs))))
	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(" " + strings.Repeat("─", panelW-1)))
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

	b.WriteString(MutedStyle.Render(fmt.Sprintf(" 类型          匹配内容                            代理目标\n")))
	b.WriteString(MutedStyle.Render(" " + strings.Repeat("─", panelW-1)))
	b.WriteString("\n")

	if len(m.rules) == 0 {
		b.WriteString(MutedStyle.Render("\n 暂无规则"))
	} else {
		for i, rule := range m.rules {
			payload := rule.Payload
			if len(payload) > 36 {
				payload = payload[:35] + "…"
			}
			proxy := rule.Proxy
			if len(proxy) > 24 {
				proxy = proxy[:23] + "…"
			}
			line := fmt.Sprintf(" %-14s %-36s %s", rule.Type, payload, proxy)
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

	type subInfo struct {
		Name        string
		URL         string
		Enabled     bool
		LastUpdated int64
		Interval    int
	}
	var subs []subInfo
	if m.cfgMgr != nil {
		for _, s := range m.cfgMgr.Config().Subscriptions {
			subs = append(subs, subInfo{
				Name:        s.Name,
				URL:         s.URL,
				Enabled:     true,
				LastUpdated: s.LastUpdated,
				Interval:    s.Interval,
			})
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
		if len(urlDisplay) > 44 {
			urlDisplay = urlDisplay[:43] + "…"
		}

		line := fmt.Sprintf(" %s  %-12s  %s  更新间隔:%s  %s",
			status, s.Name, urlDisplay, fmt.Sprintf("%dh", s.Interval/3600), MutedStyle.Render(updated))

		if i == m.subIdx {
			b.WriteString(SelectedStyle.Render(line))
		} else {
			b.WriteString(NormalStyle.Render(line))
		}
		b.WriteString("\n")
	}

	return PanelStyle.Width(panelW).Render(b.String())
}
