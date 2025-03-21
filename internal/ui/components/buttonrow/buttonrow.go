package buttonrow

import (
	"errors"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var ErrEmptyButtons = errors.New("buttons can not be empty")

type KeyMap struct {
	Left  key.Binding
	Right key.Binding
}

var DefaultKeyMap = KeyMap{
	Left: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "previous button"),
	),
	Right: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "next button"),
	),
}

type Button interface {
	Label() string
	Color() lipgloss.TerminalColor
}

type Model struct {
	focused      bool
	buttons      []Button
	selectedIdx  int
	buttonStyle  lipgloss.Style
	focusedStyle lipgloss.Style
	keyMap       KeyMap
}

func (m *Model) Focus() tea.Cmd {
	m.focused = true
	return nil
}

func (m *Model) Blur() {
	m.focused = false
}

func (m *Model) SelectedButton() Button {
	return m.buttons[m.selectedIdx]
}

func (m *Model) Next() {
	m.selectedIdx = (m.selectedIdx + 1) % len(m.buttons)
}

func (m *Model) Prev() {
	m.selectedIdx--
	if m.selectedIdx < 0 {
		m.selectedIdx = len(m.buttons) - 1
	}
}

func (m Model) View() string {

	if len(m.buttons) == 0 {
		return ""
	}

	var buttonViews []string
	for i, button := range m.buttons {
		style := m.buttonStyle
		style = style.BorderForeground(button.Color())
		if m.focused && i == m.selectedIdx {
			style = m.focusedStyle
			style = style.Foreground(button.Color()).BorderForeground(button.Color()).BorderStyle(lipgloss.DoubleBorder()).Bold(true)
		}
		buttonViews = append(buttonViews, style.Render(button.Label()))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, buttonViews...)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Left):
			m.Prev()
		case key.Matches(msg, m.keyMap.Right):
			m.Next()
		}
	}

	return m, nil
}

func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{m.keyMap.Left, m.keyMap.Right}
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{m.keyMap.Left, m.keyMap.Right},
	}
}

type Option func(*Model)

func WithButtonStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.buttonStyle = style
	}
}

func WithFocusedStyle(style lipgloss.Style) Option {
	return func(m *Model) {
		m.focusedStyle = style
	}
}

func WithKeyMap(keyMap KeyMap) Option {
	return func(m *Model) {
		m.keyMap = keyMap
	}
}

func New(buttons []Button, opts ...Option) (*Model, error) {
	if len(buttons) == 0 {
		return nil, ErrEmptyButtons
	}

	defaultButtonStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#888888"))

	defaultFocusedStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		Foreground(lipgloss.Color("#FFFFFF"))

	m := &Model{
		focused:      false,
		buttons:      buttons,
		selectedIdx:  0,
		buttonStyle:  defaultButtonStyle,
		focusedStyle: defaultFocusedStyle,
		keyMap:       DefaultKeyMap,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m, nil
}
