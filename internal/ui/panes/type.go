package panes

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

type Pane interface {
	tea.Model
	help.KeyMap
}
