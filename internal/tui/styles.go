package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/rivo/uniseg"
)

// Color palette
const (
	colorPrimary  = "#7c3aed"
	colorPrimary2 = "#a78bfa"
	colorSuccess  = "#22c55e"
	colorWarning  = "#eab308"
	colorDanger   = "#ef4444"
	colorMuted    = "#71717a"
	colorBg       = "#18181b"
	colorBgLight  = "#27272a"
	colorBorder   = "#3f3f46"
	colorText     = "#e4e4e7"
	colorTextDim  = "#a1a1aa"
	colorCyan     = "#06b6d4"
)

// ─── App Shell ────────────────────────────────────────────────

var (
	AppStyle = lipgloss.NewStyle().
			Margin(1, 1)

	// Thin line used as horizontal divider
	DividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorBorder))
)

// ─── Text Styles ──────────────────────────────────────────────

var (
	BoldStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorText))

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorText))

	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTextDim))

	AccentStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorPrimary))

	SelectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(lipgloss.Color(colorPrimary))

	GroupHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorPrimary2))

	TypeBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorCyan)).
			Background(lipgloss.Color(colorBgLight)).
			Padding(1, 1)
)

// ─── Status Indicators ────────────────────────────────────────

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

// ─── Tab Bar ──────────────────────────────────────────────────

var (
	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Padding(0, 2)

	TabInactiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Padding(0, 2)

	TabSeparator = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorBorder))
)

// ─── Panel ────────────────────────────────────────────────────

var (
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Padding(1, 2).
			Margin(0, 0, 1, 0)

	PanelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorPrimary)).
			Padding(0, 1)
)

// ─── Footer ───────────────────────────────────────────────────

var (
	FooterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTextDim)).
			Padding(0, 1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPrimary)).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTextDim))

	SearchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorText)).
			Background(lipgloss.Color(colorBgLight)).
			Padding(1, 1)
)

// ─── Log Level Colors ─────────────────────────────────────────

var (
	LogInfoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorCyan))
	LogWarnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(colorWarning))
	LogErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorDanger))
	LogDebugStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted))
)

// ─── List Styles ──────────────────────────────────────────────

var (
	ListHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTextDim)).
			Bold(true)

	ScrollHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Align(lipgloss.Center)
)

// ─── Helpers ──────────────────────────────────────────────────

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
		return MutedStyle.Render(strings.Repeat("·", width))
	}
	ratio := float64(delay) / float64(maxDelay)
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	if filled < 1 && delay > 0 {
		filled = 1
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("·", width-filled)
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

// displayWidth returns the display cell width of a string.
func displayWidth(s string) int {
	return uniseg.StringWidth(s)
}

// Truncate clips a string to maxWidth display cells with ellipsis.
func Truncate(s string, maxWidth int) string {
	if displayWidth(s) <= maxWidth {
		return s
	}
	tail := "…"
	limit := maxWidth - displayWidth(tail)
	if limit < 0 {
		limit = 0
	}
	gr := uniseg.NewGraphemes(s)
	cut := 0
	w := 0
	for gr.Next() {
		cw := gr.Width()
		if w+cw > limit {
			break
		}
		w += cw
		cut += len(gr.Str())
	}
	if cut == 0 {
		return tail
	}
	return s[:cut] + tail
}

// PadRight pads s on the right with spaces to reach width display cells.
func PadRight(s string, width int) string {
	w := displayWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// PadLeft pads s on the left with spaces to reach width display cells.
func PadLeft(s string, width int) string {
	w := displayWidth(s)
	if w >= width {
		return s
	}
	return strings.Repeat(" ", width-w) + s
}
