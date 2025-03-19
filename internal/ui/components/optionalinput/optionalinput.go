package optionalinput

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

// Define styles using lipgloss
var (
	enabledStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	disabledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	checkboxStyle = lipgloss.NewStyle().Bold(true)
)

// It implements types.Input[*T] while accepting a types.Input[T].
type OptionalInput[T comparable] struct {
	input      types.Input[T]
	enabled    bool
	focused    bool
	toggleKeys keyMap
}

// New creates a new OptionalInput that wraps the provided input.
func New[T comparable](input types.Input[T]) *OptionalInput[T] {
	return &OptionalInput[T]{
		input:      input,
		enabled:    false,
		focused:    false,
		toggleKeys: newKeyMap(),
	}
}

func (o *OptionalInput[T]) Update(msg tea.Msg) (types.Input[*T], tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if o.focused && key.Matches(msg, o.toggleKeys.Toggle) {
			o.enabled = !o.enabled
			if o.enabled {
				return o, o.input.Focus()
			}
			o.input.Blur()
			return o, nil
		}
	}

	if o.enabled {
		var cmd tea.Cmd
		o.input, cmd = o.input.Update(msg)
		return o, cmd
	}

	return o, nil
}

func (o *OptionalInput[T]) View() string {
	checkbox := checkboxStyle.Render("[ ]")
	if o.enabled {
		checkbox = checkboxStyle.Render("[x]")
	}

	if o.enabled {
		return lipgloss.JoinHorizontal(lipgloss.Top, checkbox, enabledStyle.Render(o.input.View()))
	}

	return checkbox + disabledStyle.Render("(disabled)")
}

func (o *OptionalInput[T]) Focus() tea.Cmd {
	o.focused = true
	if o.enabled {
		return o.input.Focus()
	}
	return nil
}

func (o *OptionalInput[T]) Blur() {
	o.focused = false
	o.input.Blur()
}

func (o *OptionalInput[T]) Error() error {
	if o.enabled {
		return o.input.Error()
	}
	return nil
}

func (o *OptionalInput[T]) Value() *T {
	if o.enabled {
		val := o.input.Value()
		return &val
	}
	return nil
}

func (o *OptionalInput[T]) SetEnabled(enabled bool) {
	o.enabled = enabled
}

func (o *OptionalInput[T]) IsEnabled() bool {
	return o.enabled
}

// ShortHelp returns keybinding help.
func (o *OptionalInput[T]) ShortHelp() []key.Binding {
	if o.enabled {
		return append(o.input.ShortHelp(), o.toggleKeys.Toggle)
	}
	return []key.Binding{o.toggleKeys.Toggle}
}

func (o *OptionalInput[T]) FullHelp() [][]key.Binding {
	if o.enabled {
		inputHelp := o.input.FullHelp()
		toggleHelp := [][]key.Binding{{o.toggleKeys.Toggle}}
		return append(inputHelp, toggleHelp...)
	}
	return [][]key.Binding{{o.toggleKeys.Toggle}}
}

type keyMap struct {
	Toggle key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Toggle: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle input"),
		),
	}
}

func (m OptionalInput[T]) Init() tea.Cmd {
	return nil
}

func (m *OptionalInput[T]) SetValue(t *T) error {
	if t == nil {
		m.enabled = false
		return nil
	}

	m.enabled = true
	return m.input.SetValue(*t)
}
