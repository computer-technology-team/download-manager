package panes

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

// Keeping the ELM structure whilst adding KeyMap
type Pane interface {
	help.KeyMap
	Init() tea.Cmd
	Update(tea.Msg) (Pane, tea.Cmd)
	View() string
}
