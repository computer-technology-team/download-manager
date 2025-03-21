package startendtimeinput

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/state"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/timeinput"
	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

type focusedInput int

const (
	startTimeFocused focusedInput = iota
	endTimeFocused
)

type StartEndTime lo.Tuple2[state.TimeValue, state.TimeValue]

type KeyMap struct {
	Next key.Binding
	Prev key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Next: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "next time input"),
		),
		Prev: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "previous time input"),
		),
	}
}

type Model struct {
	startTime *timeinput.Model
	endTime   *timeinput.Model
	focused   focusedInput
	keyMap    KeyMap
}

func (m *Model) Blur() {
	switch m.focused {
	case startTimeFocused:
		m.startTime.Blur()
	case endTimeFocused:
		m.endTime.Blur()
	}

}

func (m *Model) Error() error {
	return errors.Join(m.startTimeError(), m.endTimeError())
}

func (m Model) startTimeError() error {
	if m.startTime.Error() != nil {
		return fmt.Errorf("start time: %w", m.startTime.Error())
	}
	return nil
}

func (m Model) endTimeError() error {
	if m.endTime.Error() != nil {
		return fmt.Errorf("end time: %w", m.endTime.Error())
	}
	return nil
}

func (m *Model) Focus() tea.Cmd {
	m.focused = startTimeFocused

	return m.startTime.Focus()
}

func (m *Model) FullHelp() [][]key.Binding {
	var bindings [][]key.Binding

	bindings = append(bindings, []key.Binding{m.keyMap.Next, m.keyMap.Prev})

	switch m.focused {
	case startTimeFocused:
		bindings = append(bindings, m.startTime.FullHelp()...)
	case endTimeFocused:
		bindings = append(bindings, m.endTime.FullHelp()...)
	}

	return bindings
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.startTime.Init(),
		m.endTime.Init(),
	)
}

func (m *Model) SetValue(value StartEndTime) error {
	startErr := m.startTime.SetValue(value.A)
	endErr := m.endTime.SetValue(value.B)

	if startErr != nil || endErr != nil {
		return errors.Join(startErr, endErr)
	}

	return nil
}

func (m *Model) ShortHelp() []key.Binding {
	bindings := []key.Binding{m.keyMap.Next, m.keyMap.Prev}

	switch m.focused {
	case startTimeFocused:
		bindings = append(bindings, m.startTime.ShortHelp()...)
	case endTimeFocused:
		bindings = append(bindings, m.endTime.ShortHelp()...)
	}

	return bindings
}

func (m *Model) Update(msg tea.Msg) (types.Input[StartEndTime], tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch {
		case key.Matches(msg, m.keyMap.Next):
			if m.focused == startTimeFocused {
				m.startTime.Blur()
				m.focused = endTimeFocused

				return m, m.endTime.Focus()
			}
		case key.Matches(msg, m.keyMap.Prev):
			if m.focused == endTimeFocused {
				m.endTime.Blur()
				m.focused = startTimeFocused

				return m, m.startTime.Focus()
			}
		}

		switch m.focused {
		case startTimeFocused:
			updatedInput, cmd := m.startTime.Update(msg)
			m.startTime = updatedInput.(*timeinput.Model)

			cmds = append(cmds, cmd)
		case endTimeFocused:
			updatedInput, cmd := m.endTime.Update(msg)
			m.endTime = updatedInput.(*timeinput.Model)

			cmds = append(cmds, cmd)
		}
	case cursor.BlinkMsg:

		switch m.focused {
		case startTimeFocused:
			updatedInput, cmd := m.startTime.Update(msg)
			m.startTime = updatedInput.(*timeinput.Model)

			cmds = append(cmds, cmd)
		case endTimeFocused:
			updatedInput, cmd := m.endTime.Update(msg)
			m.endTime = updatedInput.(*timeinput.Model)

			cmds = append(cmds, cmd)
		}
	default:

		startInput, startCmd := m.startTime.Update(msg)
		m.startTime = startInput.(*timeinput.Model)

		endInput, endCmd := m.endTime.Update(msg)
		m.endTime = endInput.(*timeinput.Model)

		cmds = append(cmds, startCmd, endCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) Value() StartEndTime {
	return StartEndTime{
		A: m.startTime.Value(),
		B: m.endTime.Value(),
	}
}

func (m *Model) View() string {
	startLabel := lipgloss.NewStyle().
		Bold(true).
		Render("Start Time: ")

	endLabel := lipgloss.NewStyle().
		Bold(true).
		Render("End Time: ")

	startTimeView := startLabel + m.startTime.View()
	endTimeView := endLabel + m.endTime.View()

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		startTimeView,
		"   ", 
		endTimeView,
	)
}

func New() types.Input[StartEndTime] {
	startTimeModel := timeinput.New()
	endTimeModel := timeinput.New()

	endTimeModel.Blur()

	model := &Model{
		startTime: startTimeModel,
		endTime:   endTimeModel,
		focused:   startTimeFocused,
		keyMap:    DefaultKeyMap(),
	}

	return model
}
