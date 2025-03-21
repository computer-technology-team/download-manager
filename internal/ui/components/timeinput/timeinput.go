package timeinput

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/computer-technology-team/download-manager.git/internal/state"
	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

type section int

const (
	hourSection section = iota
	minuteSection
	secondSection
	totalSections
)

type keyMap struct {
	Next     key.Binding
	Previous key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Next: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "next field"),
		),
		Previous: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "previous field"),
		),
	}
}

type Model struct {
	hourInput   *textinput.Model
	minuteInput *textinput.Model
	secondInput *textinput.Model
	focused     section
	err         error
	style       lipgloss.Style
	separator   string
	keyMap      keyMap
}

func New() *Model {
	hourInput := textinput.New()
	hourInput.Placeholder = "HH"
	hourInput.CharLimit = 2
	hourInput.Width = 2
	hourInput.Validate = func(s string) error {
		if s != "" {
			hour, err := strconv.Atoi(s)
			if err != nil || hour < 0 || hour > 23 {
				return errors.New("hour must be between 0 and 23")
			}
		}
		return nil
	}
	hourInput.Focus()

	minuteInput := textinput.New()
	minuteInput.Placeholder = "MM"
	minuteInput.CharLimit = 2
	minuteInput.Width = 2
	minuteInput.Validate = func(s string) error {
		if s != "" {
			minute, err := strconv.Atoi(s)
			if err != nil || minute < 0 || minute > 59 {
				return errors.New("minute must be between 0 and 59")
			}
		}
		return nil
	}

	secondInput := textinput.New()
	secondInput.Placeholder = "SS"
	secondInput.CharLimit = 2
	secondInput.Width = 2
	secondInput.Validate = func(s string) error {
		if s != "" {
			second, err := strconv.Atoi(s)
			if err != nil || second < 0 || second > 59 {
				return errors.New("second must be between 0 and 59")
			}
		}
		return nil
	}

	return &Model{
		hourInput:   &hourInput,
		minuteInput: &minuteInput,
		secondInput: &secondInput,
		focused:     hourSection,
		separator:   ":",
		style:       lipgloss.NewStyle(),
		keyMap:      defaultKeyMap(),
	}
}

func (m *Model) Value() state.TimeValue {
	hour, _ := strconv.Atoi(m.hourInput.Value())
	minute, _ := strconv.Atoi(m.minuteInput.Value())
	second, _ := strconv.Atoi(m.secondInput.Value())
	return state.TimeValue{Hour: hour, Minute: minute, Second: second}
}

func (m *Model) Focus() tea.Cmd {
	m.focused = hourSection
	m.minuteInput.Blur()
	m.secondInput.Blur()
	return m.hourInput.Focus()
}

func (m *Model) Blur() {
	m.hourInput.Blur()
	m.minuteInput.Blur()
	m.secondInput.Blur()
}

func (m *Model) Focused() bool {
	return m.hourInput.Focused() || m.minuteInput.Focused() || m.secondInput.Focused()
}

func (m *Model) validate() {
	m.err = errors.Join(m.hourInput.Err, m.minuteInput.Err, m.secondInput.Err)
}

func (m *Model) Update(msg tea.Msg) (types.Input[state.TimeValue], tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyMatched := true
		switch {
		case key.Matches(msg, m.keyMap.Next):
			m.Blur()
			m.focused = (m.focused + 1) % totalSections
			cmds = append(cmds, m.getFocusedInput().Focus())
		case key.Matches(msg, m.keyMap.Previous):
			m.Blur()
			m.focused = (m.focused - 1 + totalSections) % totalSections
			cmds = append(cmds, m.getFocusedInput().Focus())
		default:
			keyMatched = false
		}
		if keyMatched {
			return m, tea.Batch(cmds...)
		}

	}

	updatedFocusInput, cmd := m.getFocusedInput().Update(msg)
	cmds = append(cmds, cmd)

	m.setFocusedInput(updatedFocusInput)
	m.validate()

	return m, tea.Batch(cmds...)
}

func (m Model) getFocusedInput() *textinput.Model {
	switch m.focused {
	case hourSection:
		return m.hourInput
	case minuteSection:
		return m.minuteInput
	case secondSection:
		return m.secondInput
	default:
		return m.hourInput
	}
}

func (m *Model) setFocusedInput(inputModel textinput.Model) {
	switch m.focused {
	case hourSection:
		m.hourInput = &inputModel
	case minuteSection:
		m.minuteInput = &inputModel
	case secondSection:
		m.secondInput = &inputModel
	}
}

func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{m.keyMap.Next, m.keyMap.Previous}
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{{m.keyMap.Next, m.keyMap.Previous}}
}

func (m Model) View() string {
	hourView := m.hourInput.View()
	minuteView := m.minuteInput.View()
	secondView := m.secondInput.View()

	timeView := fmt.Sprintf("%s%s%s%s%s", hourView, m.separator, minuteView, m.separator, secondView)

	return m.style.Render(timeView)
}

func (m Model) Error() error {
	return m.err
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetValue(val state.TimeValue) error {
	if err := val.Validate(); err != nil {
		return err
	}

	m.hourInput.SetValue(strconv.Itoa(val.Hour))
	m.minuteInput.SetValue(strconv.Itoa(val.Minute))
	m.secondInput.SetValue(strconv.Itoa(val.Second))

	return nil
}
