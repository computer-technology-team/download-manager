package types

import (
	"errors"

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

type TimeValue struct {
	Hour   int
	Minute int
	Second int
}

func (t TimeValue) Validate() error {
	if t.Hour < 0 || t.Hour > 23 {
		return errors.New("hour must be between 0 and 23")
	}
	if t.Minute < 0 || t.Minute > 59 {
		return errors.New("minute must be between 0 and 59")
	}
	if t.Second < 0 || t.Second > 59 {
		return errors.New("second must be between 0 and 59")
	}
	return nil
}
