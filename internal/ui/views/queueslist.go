package views

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/events"
	"github.com/computer-technology-team/download-manager.git/internal/queues"
	"github.com/computer-technology-team/download-manager.git/internal/state"
	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

type queueListViewMode string

type backToTable struct{}

const (
	tableMode      queueListViewMode = "table"
	createFormMode queueListViewMode = "create-form"
	editFormMode   queueListViewMode = "edit-form"
)

var queuesColumnRatios = []float64{
	0.15,
	0.15,
	0.3,
	0.1,
	0.3,
}

var queuesColumns = []table.Column{
	{Title: "Name", Width: 10},
	{Title: "Bandwidth Limit (Bytes Per Second)", Width: 10},
	{Title: "Download Directory", Width: 10},
	{Title: "Maxiumum Concurrent Download", Width: 10},
	{Title: "Start - End Time", Width: 10},
}

// KeyMap defines the keybindings for the queues list view
type queueListKeyMap struct {
	EditQueue key.Binding
	NewQueue  key.Binding
	Back      key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultQueueListKeyMap() queueListKeyMap {
	return queueListKeyMap{
		NewQueue: key.NewBinding(
			key.WithKeys("n", "+"),
			key.WithHelp("n/+", "new queue"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back to list"),
		),
	}
}

func backToTableCmd() tea.Msg {
	return backToTable{}
}

type queuesListView struct {
	mode queueListViewMode

	tableModel      table.Model
	queueCreateForm *queueForm
	queueEditForm   *queueForm
	keyMap          queueListKeyMap

	queues []state.Queue

	width  int
	height int

	queueManager queues.QueueManager
}

func (m *queuesListView) updateColumnWidths() {
	availableWidth := m.width

	if availableWidth <= 0 {
		return
	}

	columns := m.tableModel.Columns()
	for i, col := range columns {
		if i < len(queuesColumnRatios) {
			col.Width = int(float64(availableWidth) * queuesColumnRatios[i])

			columns[i] = col
		}
	}

	m.tableModel.SetColumns(columns)
}

func (m queuesListView) FullHelp() [][]key.Binding {
	switch m.mode {
	case tableMode:
		tableHelp := m.tableModel.KeyMap.FullHelp()
		return append(tableHelp, []key.Binding{m.keyMap.NewQueue, m.keyMap.EditQueue})
	case editFormMode:
		return m.queueEditForm.FullHelp()
	case createFormMode:
		return m.queueCreateForm.FullHelp()
	default:
		return nil
	}
}

func (m queuesListView) ShortHelp() []key.Binding {
	switch m.mode {
	case tableMode:
		return append(m.tableModel.KeyMap.ShortHelp(), m.keyMap.NewQueue, m.keyMap.EditQueue)
	case editFormMode:
		return m.queueCreateForm.ShortHelp()
	case createFormMode:
		return m.queueCreateForm.ShortHelp()
	default:
		return nil
	}

}

func (m queuesListView) Init() tea.Cmd { return nil }

func (m *queuesListView) switchToCreateFormMode() tea.Cmd {
	m.mode = createFormMode
	if m.queueCreateForm == nil {
		m.queueCreateForm = NewQueueCreateForm(m.queueManager, backToTableCmd)
		return m.queueCreateForm.Init()
	}

	return nil
}

func (m *queuesListView) switchToEditFormMode() tea.Cmd {
	m.mode = editFormMode
	if m.queueEditForm == nil {
		var err error
		m.queueEditForm, err = NewQueueEditForm(state.Queue{}, m.queueManager, backToTableCmd)
		if err != nil {
			slog.Error("could not render edit form", "error", err)
			m.mode = tableMode
		} else {
			return m.queueCreateForm.Init()
		}
	}

	return nil
}

func (m *queuesListView) switchToTableMode() {
	m.mode = tableMode
}

func (m *queuesListView) handleEvent(msg events.Event) {
	switch msg.EventType {
	case events.QueueCreated:
		m.queues = append(m.queues, msg.Payload.(state.Queue))
	case events.QueueDeleted:
		deletedQueueID := msg.Payload.(int64)
		m.queues = lo.Filter(m.queues, func(queue state.Queue, _ int) bool {
			return deletedQueueID != queue.ID
		})

	case events.QueueEdited:
		queue := msg.Payload.(state.Queue)
		_, queueIdx, found := lo.FindIndexOf(m.queues, func(queueI state.Queue) bool {
			return queueI.ID == queue.ID
		})
		if !found {
			slog.Warn("queue was not found but update requested", "queue_id", queue.ID)
		}

		m.queues[queueIdx] = queue
	default:
		return
	}

	m.tableModel.SetRows(lo.Map(m.queues, func(queue state.Queue, _ int) table.Row {
		return queueToQueueTableRow(queue)
	}))
}

func (m queuesListView) Update(msg tea.Msg) (types.View, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case events.Event:
		m.handleEvent(msg)
	case backToTable:
		m.switchToTableMode()
		return m, nil
	case tea.KeyMsg:
		switch m.mode {
		case tableMode:
			// Handle table mode key presses
			switch {
			case key.Matches(msg, m.keyMap.NewQueue):
				return m, m.switchToCreateFormMode()
			case key.Matches(msg, m.keyMap.EditQueue):
				return m, m.switchToEditFormMode()
			}
		case createFormMode:
			if key.Matches(msg, m.keyMap.Back) {
				m.switchToTableMode()
				return m, nil
			}

			formView, cmd := m.queueCreateForm.Update(msg)
			m.queueCreateForm = &formView

			return m, cmd
		case editFormMode:
			if key.Matches(msg, m.keyMap.Back) {
				m.switchToTableMode()
				return m, nil
			}

			formView, cmd := m.queueEditForm.Update(msg)
			m.queueCreateForm = &formView
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tableModel.SetWidth(msg.Width)
		m.tableModel.SetHeight(msg.Height)
		m.updateColumnWidths()
	}

	if m.mode == tableMode {
		m.tableModel, cmd = m.tableModel.Update(msg)
	} else {
		var formView queueForm
		formView, cmd = m.queueCreateForm.Update(msg)
		m.queueCreateForm = &formView
	}

	return m, cmd
}

func (m queuesListView) View() string {
	if m.mode == tableMode {
		return m.tableModel.View()
	}

	return m.queueCreateForm.View()
}

func NewQueuesList(ctx context.Context, queueManager queues.QueueManager) (types.View, error) {
	queues, err := queueManager.ListQueue(ctx)
	if err != nil {
		return nil, err
	}

	rows := lo.Map(queues, func(queue state.Queue, _ int) table.Row {
		return queueToQueueTableRow(queue)
	})

	t := table.New(
		table.WithRows(rows),
		table.WithColumns(queuesColumns),
		table.WithFocused(true),
		table.WithStyles(tableStyles),
	)

	return queuesListView{
		tableModel: t,
		mode:       tableMode,
		keyMap:     DefaultQueueListKeyMap(),

		queues: queues,

		queueManager: queueManager,
	}, nil
}

func queueToQueueTableRow(queue state.Queue) table.Row {
	var bandwidthLimit, startEndTime string

	if queue.MaxBandwidth.Valid {
		bandwidthLimit = FormatBytesPerSecond(queue.MaxBandwidth.Int64)
	} else {
		bandwidthLimit = "No Limit"
	}

	if queue.ScheduleMode {
		startEndTime = fmt.Sprintf("%s - %s", queue.StartDownload, queue.EndDownload)
	} else {
		startEndTime = "No Schedule"
	}

	return table.Row{queue.Name, bandwidthLimit, queue.Directory, "", startEndTime}
}
