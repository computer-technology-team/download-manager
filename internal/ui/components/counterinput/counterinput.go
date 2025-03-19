package counterinput

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

var ErrInvalidValue = errors.New("value must be between min and max")

// Option is a function that configures a Model
type Option func(*Model)

type StepHandler func(step int64, value int64) int64

// WithStep sets the step value for the counter
func WithStep(step int64) Option {
	return func(m *Model) {
		if step > 0 {
			m.step = step
		}
	}
}

func WithMin(mn int64) Option {
	return func(m *Model) {
		m.min = min(mn, m.max)
	}
}

func WithMax(mx int64) Option {
	return func(m *Model) {
		m.max = max(mx, m.min)
	}
}

// WithStyle sets a custom style for the counter
func WithStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.style = style
	}
}

func WithStepHandler(stepHandler StepHandler) Option {
	return func(m *Model) {
		m.stepHandler = stepHandler
	}
}

type keyMap struct {
	Increment key.Binding
	Decrement key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Increment: key.NewBinding(
			key.WithKeys("+"),
			key.WithHelp("+", "increment"),
		),
		Decrement: key.NewBinding(
			key.WithKeys("-", "_"),
			key.WithHelp("-/_", "decrement"),
		),
	}
}

type Model struct {
	value       int64
	min         int64
	max         int64
	step        int64
	stepHandler StepHandler
	focused     bool
	err         error
	style       lipgloss.Style
	keyMap      keyMap
}

func New(opts ...Option) *Model {

	defaultStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")). // Brighter blue
		Bold(true)

	m := &Model{
		value: 0,
		min:   0,
		max:   100,
		step:  1,
		stepHandler: func(step, _ int64) int64 {
			return step
		},
		focused: false,
		style:   defaultStyle,
		keyMap:  defaultKeyMap(),
	}

	// Apply all options
	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Value returns the current counter value
func (m Model) Value() int64 {
	return m.value
}

func (m *Model) Error() error {
	return m.err
}

func (m *Model) Focus() tea.Cmd {
	m.focused = true
	return nil
}

func (m *Model) Blur() {
	m.focused = false
}

func (m *Model) Focused() bool {
	return m.focused
}

func (m *Model) increment() {
	m.value = min(m.value+m.getStep(), m.max)
}

func (m *Model) decrement() {
	m.value = max(m.value-m.getStep(), m.min)
}

func (m Model) getStep() int64 {
	return m.stepHandler(m.step, m.value)
}

func (m *Model) Update(msg tea.Msg) (types.Input[int64], tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Increment):
			m.increment()
		case key.Matches(msg, m.keyMap.Decrement):
			m.decrement()
		}
	}

	return m, cmd
}

func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{m.keyMap.Increment, m.keyMap.Decrement}
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{{m.keyMap.Increment, m.keyMap.Decrement}}
}

func (m Model) View() string {
	style := m.style

	if m.focused {
		style = style.
			BorderForeground(lipgloss.Color("213")). // Vibrant magenta
			Foreground(lipgloss.Color("213")).
			Background(lipgloss.Color("236")) // Dark background for contrast
	} else {
		style = style.
			BorderForeground(lipgloss.Color("246")). // Lighter gray for better visibility
			Foreground(lipgloss.Color("252"))        // Even lighter gray for the text
	}

	return style.Render(fmt.Sprintf("%d", m.value))
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetValue(val int64) error {
	if val < m.min || val > m.max {
		return ErrInvalidValue
	}

	m.value = val
	return nil
}
