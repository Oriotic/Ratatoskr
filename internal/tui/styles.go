package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent = lipgloss.Color("#7DD3FC")
	colorGood   = lipgloss.Color("#86EFAC")
	colorBad    = lipgloss.Color("#FCA5A5")
	colorDim    = lipgloss.Color("#6B7280")
	colorText   = lipgloss.Color("#E5E7EB")
	colorWarn   = lipgloss.Color("#FDE68A")

	titleStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	ruleStyle = lipgloss.NewStyle().Foreground(colorDim)

	cursorStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	selectedStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	dimStyle = lipgloss.NewStyle().Foreground(colorDim)

	goodStyle = lipgloss.NewStyle().Foreground(colorGood).Bold(true)

	badStyle = lipgloss.NewStyle().Foreground(colorBad).Bold(true)

	warnStyle = lipgloss.NewStyle().Foreground(colorWarn)

	textStyle = lipgloss.NewStyle().Foreground(colorText)

	helpStyle = lipgloss.NewStyle().Foreground(colorDim).Italic(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDim).
			Padding(1, 2)
)

func rule(width int) string {
	if width <= 0 {
		width = 41
	}
	s := ""
	for i := 0; i < width; i++ {
		s += "─"
	}
	return ruleStyle.Render(s)
}
