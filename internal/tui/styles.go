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

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color(colorPrimary))

	GroupHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#a78bfa"))

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

// Panel
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

	SearchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorText)).
			Background(lipgloss.Color(colorBgLight)).
			Padding(0, 1)
)

// Log level styles
var (
	LogInfoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan))
	LogWarnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorWarning))
	LogErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDanger))
	LogDebugStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
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
	return lipgloss.NewStyle().Foreground(LatencyColor(delay)).Render(bar)
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
