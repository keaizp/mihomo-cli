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

	footer := m.renderFooter()

	mainHeight := m.height - lipgloss.Height(header) - lipgloss.Height(tabBar) - lipgloss.Height(footer) - 2
	if mainHeight < 5 {
		mainHeight = 5
	}
	body = lipgloss.NewStyle().Height(mainHeight).Render(body)

	content := lipgloss.JoinVertical(lipgloss.Left, header, tabBar, body, footer)
	return AppStyle.Render(content)
}

// ─── Header ───────────────────────────────────────────────────

func (m Model) renderHeader() string {
	appName := BoldStyle.Foreground(lipgloss.Color(colorPrimary)).Render("⚡ mihomo-cli")

	statusText := StatusStoppedStyle.Render("● 已停止")
	if m.kernelMgr != nil {
		switch m.kernelMgr.Status() {
		case "running":
			statusText = StatusRunningStyle.Render("● 运行中")
		case "starting":
			statusText = StatusStartingStyle.Render("● 启动中")
		}
	}

	modeText := "规则"
	if m.cfgMgr != nil {
		switch m.cfgMgr.Config().Mode {
		case "global":
			modeText = "全局"
		case "direct":
			modeText = "直连"
		case "script":
			modeText = "脚本"
		}
	}
	modeBadge := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorPrimary2)).
		Background(lipgloss.Color(colorBgLight)).
		Padding(0, 1).
		Render(modeText)

	traffic := ""
	if m.trafficUp > 0 || m.trafficDown > 0 {
		up := lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess)).Render(fmt.Sprintf("↑%s", FormatRate(m.trafficUp)))
		down := lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan)).Render(fmt.Sprintf("↓%s", FormatRate(m.trafficDown)))
		traffic = up + " " + down
	}

	right := fmt.Sprintf("%s  %s  %s", statusText, modeBadge, traffic)

	if m.notification != "" {
		right += "  " + MutedStyle.Render(m.notification)
	}

	rightW := lipgloss.Width(right)
	leftW := lipgloss.Width(appName)
	space := m.width - leftW - rightW - 4
	if space < 1 {
		space = 1
	}

	return appName + strings.Repeat(" ", space) + right
}

// ─── Tab Bar ──────────────────────────────────────────────────

func (m Model) renderTabBar() string {
	tabs := []string{"代理", "连接", "日志", "规则", "订阅"}
	tabCount := len(tabs)
	usableWidth := m.width - 4
	tabWidth := usableWidth / tabCount

	var rendered []string
	for i, t := range tabs {
		label := fmt.Sprintf("%d %s", i+1, t)
		var tab string
		if i == m.tabIdx {
			tab = TabActiveStyle.Width(tabWidth).Align(lipgloss.Center).Render(label)
		} else {
			tab = TabInactiveStyle.Width(tabWidth).Align(lipgloss.Center).Render(label)
		}
		if i < tabCount-1 {
			tab += TabSeparator.Render("│")
		}
		rendered = append(rendered, tab)
	}

	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	divider := DividerStyle.Render(strings.Repeat("─", usableWidth))

	return lipgloss.JoinVertical(lipgloss.Left, tabRow, divider)
}

// ─── Footer ───────────────────────────────────────────────────

func (m Model) renderFooter() string {
	if m.searchMode {
		return FooterStyle.Width(m.width - 2).Render(
			HelpKeyStyle.Render("输入") + " 搜索节点  " +
				HelpKeyStyle.Render("Esc") + " 取消  " +
				HelpKeyStyle.Render("Enter") + " 确定")
	}

	var keys []string
	switch m.tabIdx {
	case 0: // 代理
		keys = append(keys, "↑↓ 导航", "Enter 切换", "Tab 换组", "t 测速", "a 全组测速", "/ 搜索", "空格 折叠")
	case 1: // 连接
		keys = append(keys, "↑↓ 导航", "d 关闭连接", "X 关闭全部")
	case 2: // 日志
		keys = append(keys, "↑↓ 滚动", "l 切换级别", "s 自动滚动")
	case 3: // 规则
		keys = append(keys, "↑↓ 滚动")
	case 4: // 订阅
		keys = append(keys, "↑↓ 导航", "a 添加", "d 删除", "u 更新")
	}
	keys = append(keys, "1-5 视图", "m 模式", "r 重载", "q 退出")

	var parts []string
	for _, k := range keys {
		parts = append(parts, HelpKeyStyle.Render(k))
	}
	keyBar := strings.Join(parts, "  ")

	// Separator line above footer
	divider := DividerStyle.Render(strings.Repeat("─", m.width-4))

	return lipgloss.JoinVertical(lipgloss.Left, divider, FooterStyle.Width(m.width-2).Render(keyBar))
}

// ─── Column Layout Constants ──────────────────────────────────
//
//	┌ marker ┐  ┌── name (L) ──┐  ┌ delay (R) ┐  ┌── bar ──┐  ┌ type (L) ┐
//	  ●         香港 01                   12ms  ████······      ss
const (
	colMarker = 4  // "  ● " — marker centered
	colName   = 26 // left-aligned
	colDelay  = 6  // right-aligned, e.g. "  12ms"
	colBar    = 10 // latency bar
	colType   = 8  // left-aligned type badge
	colGap    = 2  // gap between columns
)

var proxyHeaderLine = fmt.Sprintf(
	"%-*s%-*s%-*s%-*s%-*s%-*s%-*s%-*s",
	colMarker, "",
	colName, "节点",
	colGap, "",
	colDelay, "延迟",
	colGap, "",
	colBar, "速度柱",
	colGap, "",
	colType, "类型",
)

// ─── Proxy List Item ──────────────────────────────────────────

func (m Model) renderProxyItem(node string, isNow bool, isSelected bool, maxDelay int, nodeType string) string {
	// Marker column (centered)
	marker := "  ○ "
	if isNow {
		marker = "  " + lipgloss.NewStyle().Foreground(lipgloss.Color(colorSuccess)).Bold(true).Render("●") + " "
	}

	// Name column (left-aligned, truncated)
	name := fmt.Sprintf("%-*s", colName, Truncate(node, colName))

	// Latency column (right-aligned)
	latStr := fmt.Sprintf("%*s", colDelay, "···")
	if d, ok := m.delayResults[node]; ok && d > 0 {
		latStr = lipgloss.NewStyle().Foreground(LatencyColor(d)).Bold(true).
			Render(fmt.Sprintf("%*dms", colDelay-2, d))
	}

	// Bar column
	bar := fmt.Sprintf("%-*s", colBar, LatencyBar(m.delayResults[node], maxDelay, colBar-2))

	// Type badge column (left-aligned, fixed width)
	badge := strings.Repeat(" ", colType)
	if nodeType != "" {
		displayType := Truncate(nodeType, colType-2)
		styled := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTextDim)).
			Background(lipgloss.Color(colorBgLight)).
			Padding(0, 1).
			Render(displayType)
		w := lipgloss.Width(styled)
		if w < colType {
			styled += strings.Repeat(" ", colType-w)
		}
		badge = styled
	}

	// Build column line
	line := marker +
		name +
		strings.Repeat(" ", colGap) +
		latStr +
		strings.Repeat(" ", colGap) +
		bar +
		strings.Repeat(" ", colGap) +
		badge

	if isSelected {
		return SelectedStyle.Padding(0, 1).Render(line)
	}
	return NormalStyle.Padding(0, 1).Render(line)
}

// ─── Tab 1: Proxies ───────────────────────────────────────────

func (m Model) renderProxiesTab() string {
	panelW := m.width - 4
	var b strings.Builder

	if m.searchMode {
		b.WriteString(SearchStyle.Width(panelW - 6).Render(fmt.Sprintf(" 🔍 %s▎", m.searchQuery)))
		b.WriteString("\n\n")
	}

	if m.proxiesErr != nil {
		return PanelStyle.Width(panelW - 2).Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorDanger)).Render(fmt.Sprintf("✗ 错误: %v", m.proxiesErr)))
	}

	if len(m.groups) == 0 {
		return PanelStyle.Width(panelW - 2).Render(MutedStyle.Render("暂无代理组"))
	}

	for gi, groupName := range m.groups {
		group, ok := m.proxies.Proxies[groupName]
		if !ok {
			continue
		}

		isActiveGroup := gi == m.groupIdx
		collapsed := m.collapsed[groupName]

		// Group header
		arrow := "▾"
		if collapsed {
			arrow = "▸"
		}

		groupHdr := fmt.Sprintf(" %s %s  [%d 节点]  当前: %s",
			arrow, AccentStyle.Render(groupName), len(group.All), BoldStyle.Render(group.Now))

		if isActiveGroup {
			groupHdr = lipgloss.NewStyle().
				Background(lipgloss.Color(colorBgLight)).
				Padding(0, 1).
				Render(groupHdr)
		} else {
			groupHdr = MutedStyle.Render(groupHdr)
		}
		b.WriteString(groupHdr)
		b.WriteString("\n")

		if collapsed || !isActiveGroup {
			if gi < len(m.groups)-1 {
				b.WriteString("\n")
			}
			continue
		}

		// Collect nodes and find max delay
		nodes := group.All
		if m.searchQuery != "" {
			filtered := make([]string, 0)
			q := strings.ToLower(m.searchQuery)
			for _, node := range nodes {
				if strings.Contains(strings.ToLower(node), q) {
					filtered = append(filtered, node)
				}
			}
			nodes = filtered
		}

		maxDelay := 0
		for _, node := range nodes {
			if d, ok := m.delayResults[node]; ok && d > maxDelay {
				maxDelay = d
			}
		}
		if maxDelay < 50 {
			maxDelay = 500
		}

		// Column header
		b.WriteString(MutedStyle.Render(proxyHeaderLine))
		b.WriteString("\n")
		b.WriteString(DividerStyle.Render(strings.Repeat("─", panelW-6)))
		b.WriteString("\n")

		// Render each node
		visibleCount := 0
		for ni, node := range nodes {
			// Determine node type from proxy info
			nodeType := ""
			if p, ok := m.proxies.Proxies[node]; ok {
				nodeType = p.Type
			}

			isNow := node == group.Now
			isSelected := ni == m.nodeIdx

			line := m.renderProxyItem(node, isNow, isSelected, maxDelay, nodeType)
			b.WriteString(line)
			b.WriteString("\n")
			visibleCount++
		}

		if m.speedtesting {
			b.WriteString(MutedStyle.Render("  ⏳ 正在测速..."))
			b.WriteString("\n")
		}

		// Scroll hint if many nodes
		if visibleCount > 15 && m.nodeIdx > 10 {
			b.WriteString(ScrollHintStyle.Width(panelW - 4).Render(fmt.Sprintf("↑ %d/%d ↑", m.nodeIdx+1, visibleCount)))
			b.WriteString("\n")
		}

		if gi < len(m.groups)-1 {
			b.WriteString("\n")
		}
	}

	return PanelStyle.Width(panelW - 2).Render(b.String())
}

// ─── Tab 2: Connections ───────────────────────────────────────

func (m Model) renderConnectionsTab() string {
	panelW := m.width - 4
	var b strings.Builder

	if m.connectionsErr != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorDanger)).Render(fmt.Sprintf("✗ 错误: %v", m.connectionsErr)))
		return PanelStyle.Width(panelW - 2).Render(b.String())
	}

	// Column header
	header := fmt.Sprintf("  %-6s  %-30s  %-22s  %8s  %8s",
		"协议", "目标主机", "代理链", "上行", "下行")
	b.WriteString(ListHeaderStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(DividerStyle.Render("  " + strings.Repeat("─", panelW-8)))
	b.WriteString("\n")

	if len(m.connections) == 0 {
		b.WriteString(MutedStyle.Render("\n  暂无活跃连接"))
	} else {
		for i, conn := range m.connections {
			host := Truncate(conn.Metadata.Host, 28)
			chain := Truncate(strings.Join(conn.Chains, " → "), 20)
			up := FormatBytes(conn.Upload)
			down := FormatBytes(conn.Download)

			line := fmt.Sprintf("  %-6s  %-30s  %-22s  %8s  %8s",
				conn.Metadata.Network, host, chain, up, down)

			if i == m.connIdx {
				b.WriteString(SelectedStyle.Padding(0, 1).Render(line))
			} else {
				b.WriteString(NormalStyle.Padding(0, 1).Render(line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(fmt.Sprintf("  共 %d 条活跃连接", len(m.connections))))

	return PanelStyle.Width(panelW - 2).Render(b.String())
}

// ─── Tab 3: Logs ──────────────────────────────────────────────

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

	// Filter bar
	filterBar := fmt.Sprintf("  级别: %s  │  自动滚动: %s  │  共 %d 条",
		BoldStyle.Render(levelLabel), BoldStyle.Render(autoLabel), len(m.logs))
	b.WriteString(MutedStyle.Render(filterBar))
	b.WriteString("\n")
	b.WriteString(DividerStyle.Render("  " + strings.Repeat("─", panelW-8)))
	b.WriteString("\n")

	if m.logsErr != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorDanger)).Render(fmt.Sprintf("✗ 错误: %v", m.logsErr)))
		return PanelStyle.Width(panelW - 2).Render(b.String())
	}

	if len(m.logs) == 0 {
		b.WriteString(MutedStyle.Render("\n  暂无日志"))
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
			b.WriteString(style.Render(fmt.Sprintf("  %s", entry.Payload)))
			b.WriteString("\n")
		}
	}

	return PanelStyle.Width(panelW - 2).Render(b.String())
}

// ─── Tab 4: Rules ─────────────────────────────────────────────

func (m Model) renderRulesTab() string {
	panelW := m.width - 4
	var b strings.Builder

	if m.rulesErr != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(colorDanger)).Render(fmt.Sprintf("✗ 错误: %v", m.rulesErr)))
		return PanelStyle.Width(panelW - 2).Render(b.String())
	}

	// Column header
	header := fmt.Sprintf("  %-16s  %-36s  %s", "类型", "匹配内容", "代理目标")
	b.WriteString(ListHeaderStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(DividerStyle.Render("  " + strings.Repeat("─", panelW-8)))
	b.WriteString("\n")

	if len(m.rules) == 0 {
		b.WriteString(MutedStyle.Render("\n  暂无规则"))
	} else {
		for i, rule := range m.rules {
			payload := Truncate(rule.Payload, 34)
			proxy := Truncate(rule.Proxy, 22)

			line := fmt.Sprintf("  %-16s  %-36s  %s", rule.Type, payload, proxy)
			if i == m.ruleIdx {
				b.WriteString(SelectedStyle.Padding(0, 1).Render(line))
			} else {
				b.WriteString(NormalStyle.Padding(0, 1).Render(line))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(fmt.Sprintf("  共 %d 条规则", len(m.rules))))

	return PanelStyle.Width(panelW - 2).Render(b.String())
}

// ─── Tab 5: Subscriptions ─────────────────────────────────────

func (m Model) renderSubsTab() string {
	panelW := m.width - 4
	var b strings.Builder

	if m.cfgMgr == nil {
		return PanelStyle.Width(panelW - 2).Render(MutedStyle.Render("配置未加载"))
	}

	subs := m.cfgMgr.Config().Subscriptions
	if len(subs) == 0 {
		empty := MutedStyle.Render("暂无订阅") + "\n\n" +
			HelpKeyStyle.Render("a") + " 添加订阅  " +
			HelpKeyStyle.Render("u") + " 更新全部"
		return PanelStyle.Width(panelW - 2).Render(empty)
	}

	// Header
	b.WriteString(ListHeaderStyle.Render(fmt.Sprintf("  状态  %-16s  %-40s  %s", "名称", "URL", "更新间隔")))
	b.WriteString("\n")
	b.WriteString(DividerStyle.Render("  " + strings.Repeat("─", panelW-8)))
	b.WriteString("\n")

	for i, s := range subs {
		status := StatusRunningStyle.Render("●")
		urlDisplay := Truncate(s.URL, 38)
		interval := fmt.Sprintf("%dh", s.Interval/3600)
		updated := "从未"
		if s.LastUpdated > 0 {
			updated = FormatDuration(time.Now().Unix() - s.LastUpdated) + "前"
		}

		line := fmt.Sprintf("  %s   %-16s  %-40s  %-8s  %s",
			status, s.Name, urlDisplay, interval, MutedStyle.Render(updated))

		if i == m.subIdx {
			b.WriteString(SelectedStyle.Padding(0, 1).Render(line))
		} else {
			b.WriteString(NormalStyle.Padding(0, 1).Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(MutedStyle.Render(fmt.Sprintf("  共 %d 个订阅", len(subs))))

	return PanelStyle.Width(panelW - 2).Render(b.String())
}
