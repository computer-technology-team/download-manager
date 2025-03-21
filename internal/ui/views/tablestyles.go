package views

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var (
	lightGray  = lipgloss.Color("#CCCCCC")
	headerFg   = lipgloss.Color("#FFFDF5")
	headerBg   = lipgloss.Color("#25A065")
	selectedFg = lipgloss.Color("#2D9CDB")
)

var tableStyles = table.Styles{
	Header: lipgloss.NewStyle().
		Bold(true).
		Foreground(headerFg).
		Background(headerBg).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lightGray),
	Cell: lipgloss.NewStyle().
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lightGray),
	Selected: lipgloss.NewStyle().
		Foreground(selectedFg).Bold(true),
}
