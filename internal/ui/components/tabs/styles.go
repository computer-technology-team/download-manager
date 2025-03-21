package tabs

import "github.com/charmbracelet/lipgloss"

var (

	inactiveTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╯",
		BottomRight: "╰",
	}

	inactiveTabStyle = lipgloss.NewStyle().
				Border(inactiveTabBorder, true).
				BorderForeground(lipgloss.Color("#666666")).
				AlignHorizontal(lipgloss.Center).
				Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Border(activeTabBorder, true).
			BorderForeground(lipgloss.Color("#89B4FA")).
			Foreground(lipgloss.Color("#89B4FA")).
			AlignHorizontal(lipgloss.Center).
			Bold(true).
			Padding(0, 1)

	windowStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("#89B4FA")).
			Border(lipgloss.RoundedBorder()).
			UnsetBorderTop()
)
