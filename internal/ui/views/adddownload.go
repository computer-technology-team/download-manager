package views

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/ui/components/listinput"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/textinput"
	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

const (
	url = iota
	queueName
	fileName
)

// Error messages
var (
	ErrURLRequired        = errors.New("URL is required")
	ErrURLInvalidProtocol = errors.New("URL must start with http:// or https://")
)

func addDownloadCmd(url, queueName, fileName string) tea.Cmd {
	return func() tea.Msg {
		slog.Info("add download", "url", url, "queue_name", queueName, "file_name", fileName)
		return nil
	}
}

type addDownloadView struct {
	inputs  []types.Input[string]
	focused int
	err     error
}

func (s addDownloadView) FullHelp() [][]key.Binding {
	return [][]key.Binding{s.ShortHelp()}
}

// ShortHelp implements types.Pane.
func (s addDownloadView) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("↓", "down"), key.WithHelp("↓", "next field")),
		key.NewBinding(key.WithKeys("↑", "up"), key.WithHelp("↑", "previous field")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit/next field")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
	}
}

func NewAddDownloadPane() types.View {
	return initialModel()
}

func initialModel() addDownloadView {
	inputsUrl := textinput.New()
	inputsUrl.Placeholder = "https://example.com/"
	inputsUrl.Focus()
	inputsUrl.Width = 50
	inputsUrl.Prompt = ""

	// URL validator function
	inputsUrl.Validate = func(s string) error {
		if s == "" {
			return ErrURLRequired
		}
		if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
			return ErrURLInvalidProtocol
		}
		return nil
	}

	inputsQueueName := listinput.New("Select The Queue", "queue", "queues")

	inputsFileName := textinput.New()
	inputsFileName.Placeholder = "Leave empty to use filename from URL"
	inputsFileName.Width = 50
	inputsFileName.Prompt = ""

	inputs := make([]types.Input[string], 3)
	inputs[url] = inputsUrl
	inputs[queueName] = inputsQueueName
	inputs[fileName] = inputsFileName

	return addDownloadView{
		inputs:  inputs,
		focused: 0,
		err:     nil,
	}
}

func (m addDownloadView) Init() tea.Cmd {
	return tea.Batch(lo.Map(m.inputs, func(in types.Input[string], _ int) tea.Cmd {
		return in.Init()
	})...)
}

func (m addDownloadView) Update(msg tea.Msg) (types.View, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				if m.inputs[url].Value() == "" {
					m.err = ErrURLRequired
					return m, nil
				}

				m.err = nil
				return m, addDownloadCmd(m.inputs[url].Value(),
					m.inputs[queueName].Value(), m.inputs[fileName].Value())
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyUp:
			m.prevInput()
		case tea.KeyDown:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m addDownloadView) View() string {
	var stringBuilder strings.Builder

	stringBuilder.WriteString("Add Download\n\n")

	// URL input
	stringBuilder.WriteString("URL: ")
	if m.focused == url {
		stringBuilder.WriteString("> ")
	} else {
		stringBuilder.WriteString("  ")
	}
	stringBuilder.WriteString(m.inputs[url].View())
	if err := m.inputs[url].Error(); err != nil {
		stringBuilder.WriteString(" ⚠️ " + err.Error())
	}
	stringBuilder.WriteString("\n\n")

	// Queue name input
	stringBuilder.WriteString("Queue: ")
	if m.focused == queueName {
		stringBuilder.WriteString("> ")
	} else {
		stringBuilder.WriteString("  ")
	}
	stringBuilder.WriteString(m.inputs[queueName].View())
	if err := m.inputs[queueName].Error(); err != nil {
		stringBuilder.WriteString(" ⚠️ " + err.Error())
	}
	stringBuilder.WriteString("\n\n")

	// Filename input
	stringBuilder.WriteString("Filename: ")
	if m.focused == fileName {
		stringBuilder.WriteString("> ")
	} else {
		stringBuilder.WriteString("  ")
	}
	stringBuilder.WriteString(m.inputs[fileName].View())
	if err := m.inputs[fileName].Error(); err != nil {
		stringBuilder.WriteString(" ⚠️ " + err.Error())
	}
	stringBuilder.WriteString("\n\n")

	// Form-level error message if any
	if m.err != nil {
		stringBuilder.WriteString("Error: " + m.err.Error() + "\n\n")
	}
	return stringBuilder.String()
}

func (m *addDownloadView) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

func (m *addDownloadView) prevInput() {
	m.focused--
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
