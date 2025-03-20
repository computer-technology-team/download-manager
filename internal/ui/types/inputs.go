package types

import (

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

type Focusable interface {
	Focus() tea.Cmd
	Blur()
}

type Input[T comparable] interface {
	Init() tea.Cmd
	Update(tea.Msg) (Input[T], tea.Cmd)
	View() string
	Focus() tea.Cmd
	Blur()
	Error() error
	Value() T
	SetValue(T) error
	help.KeyMap
}

