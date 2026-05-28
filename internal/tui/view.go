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

		header := GroupHeaderStyle.Render(fmt.Sprintf(" [%s] ", groupName))
		b.WriteString(header)
		b.WriteString("\n")

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

	b.WriteString(HelpStyle.Render("up/down navigate  tab group  enter switch"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("t test  r reload  u update  q quit"))

	return b.String()
}

func (m Model) renderStatusBar() string {
	status := m.kernelMgr.Status()
	mode := "N/A"
	left := fmt.Sprintf(" mihomo: %s ", status)
	right := fmt.Sprintf(" mode: %s ", mode)
	bar := left + strings.Repeat(" ", m.width-lipgloss.Width(left)-lipgloss.Width(right)) + right
	return StatusBarStyle.Width(m.width).Render(bar)
}
