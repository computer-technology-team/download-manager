package views

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	neturl "net/url"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/events"
	"github.com/computer-technology-team/download-manager.git/internal/queues"
	"github.com/computer-technology-team/download-manager.git/internal/state"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/listinput"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/textinput"
	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

const (
	url = iota
	queueName
	fileName
)

var (
	ErrURLRequired        = errors.New("URL is required")
	ErrURLInvalidProtocol = errors.New("URL must start with http:
	ErrURLParseFailed     = errors.New("failed to parse the URL")
	ErrURLHostEmpty       = errors.New("URL host can not be empty")
)

type addDownloadFormError struct {
	error
}

type addDownloadFormClear struct{}

func (s addDownloadView) addDownloadCmd(url, fileName string, queueIDStr string) tea.Cmd {
	return func() tea.Msg {
		slog.Info("add download", "url", url, "queue_name", queueName, "file_name", fileName)

		queueID, err := strconv.ParseInt(queueIDStr, 10, 64)
		if err != nil {
			slog.Error("could not parse queue id in add download", "queue_id_str", queueIDStr)
			return types.ErrorMsg{
				Err: fmt.Errorf("could not parse queue id to add download for url: %s\n error: %w",
					url, err),
			}
		}

		err = s.queueManager.CreateDownload(context.Background(), url, fileName, queueID)
		if err != nil {
			return addDownloadFormError{error: err}
		}

		var msgs []tea.Cmd
		msgs = append(msgs, createCmd(types.NotifMsg{
			Msg: fmt.Sprintf("Download from %s added successfully", url),
		}), createCmd(addDownloadFormClear{}))

		return tea.BatchMsg(msgs)
	}
}

type addDownloadView struct {
	inputs  []types.Input[string]
	focused int
	err     error

	queues []state.Queue

	queueManager queues.QueueManager
}

func (s addDownloadView) FullHelp() [][]key.Binding {
	return [][]key.Binding{s.ShortHelp()}
}

func (s addDownloadView) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("↓", "down"), key.WithHelp("↓", "next field")),
		key.NewBinding(key.WithKeys("↑", "up"), key.WithHelp("↑", "previous field")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit/next field")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit")),
	}
}

func NewAddDownloadPane(ctx context.Context, queueManager queues.QueueManager) (types.View, error) {
	return initialModel(ctx, queueManager)
}

func initialModel(ctx context.Context, queueManager queues.QueueManager) (types.View, error) {
	queues, err := queueManager.ListQueue(ctx)
	if err != nil {
		return nil, err
	}

	inputsUrl := textinput.New()
	inputsUrl.Placeholder = "https:
	inputsUrl.Focus()
	inputsUrl.Width = 50
	inputsUrl.Prompt = ""

	inputsUrl.Validate = func(s string) error {
		if s == "" {
			return ErrURLRequired
		}

		parsedUrl, err := neturl.Parse(s)
		if err != nil {
			return errors.Join(ErrURLParseFailed, err)
		}

		if parsedUrl.Scheme != "https" && parsedUrl.Scheme != "http" {
			return ErrURLInvalidProtocol
		}

		if parsedUrl.Host == "" {
			return ErrURLHostEmpty
		}

		return nil
	}

	queueList := lo.Map(queues, func(q state.Queue, _ int) list.Item {
		return queueToAddDownloadQueueItem(q)
	})

	inputsQueueName := listinput.New("Select The Queue", "queue", "queues", queueList)

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

		queues: queues,

		queueManager: queueManager,
	}, nil
}

func (m addDownloadView) Init() tea.Cmd {
	return tea.Batch(lo.Map(m.inputs, func(in types.Input[string], _ int) tea.Cmd {
		return in.Init()
	})...)
}

func (m *addDownloadView) handleEvent(msg events.Event) (types.View, tea.Cmd) {
	var cmd tea.Cmd
	listInputM := m.inputs[queueName].(*listinput.Model)
	switch msg.EventType {
	case events.QueueCreated:
		cmd = listInputM.InsertItem(len(listInputM.Items()),
			queueToAddDownloadQueueItem(msg.Payload.(state.Queue)))
	case events.QueueEdited:
		queue := msg.Payload.(state.Queue)
		idx, found := listInputM.FindItemIdx(strconv.Itoa(int(queue.ID)))
		if found {
			cmd = listInputM.SetItem(idx, queueToAddDownloadQueueItem(queue))
		} else {
			slog.Warn("queue edited but was not in add download list", "queue_id", queue.ID)
			cmd = listInputM.InsertItem(len(listInputM.Items()),
				queueToAddDownloadQueueItem(msg.Payload.(state.Queue)))
		}
	case events.QueueDeleted:
		queueID := msg.Payload.(int64)
		idx, found := listInputM.FindItemIdx(strconv.Itoa(int(queueID)))
		if found {
			listInputM.RemoveItem(idx)
		} else {
			slog.Warn("queue was deleted and not found in add download queue list", "queue_id", queueID)
		}

	}
	return m, cmd
}

func (m addDownloadView) Update(msg tea.Msg) (types.View, tea.Cmd) {
	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case events.Event:
		return m.handleEvent(msg)
	case addDownloadFormError:
		m.err = msg.error
		return m, nil
	case addDownloadFormClear:
		m.err = nil

		urlInput := m.inputs[url]
		err := urlInput.SetValue("")
		if err != nil {
			slog.Error("could not reset url in add download form",
				"error", err)
			return m, createErrorCmd(types.ErrorMsg{
				Err: fmt.Errorf("could not reset form"),
			})
		}

		fileNameInput := m.inputs[fileName]
		err = fileNameInput.SetValue("")
		if err != nil {
			slog.Error("could not reset file name in add download form",
				"error", err)
			return m, createErrorCmd(types.ErrorMsg{
				Err: fmt.Errorf("could not reset form"),
			})
		}

		m.focused = url
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				if m.inputs[url].Value() == "" {
					m.err = ErrURLRequired
					return m, nil
				}

				m.err = nil
				return m, m.addDownloadCmd(m.inputs[url].Value(),
					m.inputs[fileName].Value(), m.inputs[queueName].Value())
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

func queueToAddDownloadQueueItem(queue state.Queue) list.Item {
	return listinput.NewItem(strconv.Itoa(int(queue.ID)), queue.Name, queue.Directory)
}
