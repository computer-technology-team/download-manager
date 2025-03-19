package types

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

// Keeping the ELM structure whilst adding KeyMap
type View interface {
	help.KeyMap
	Init() tea.Cmd
	Update(tea.Msg) (View, tea.Cmd)
	View() string
}

type Viewable interface {
	View() string
}
