package views

import (
	"context"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/computer-technology-team/download-manager.git/internal/queues"
	"github.com/computer-technology-team/download-manager.git/internal/state"
	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

var downloadsColumnRatios = []float64{
	0.45,
	0.30,
	0.25,
}

var downloadsColumns = []table.Column{
	{Title: "URL", Width: 10},
	{Title: "Queue Name", Width: 10},
	{Title: "Progress", Width: 10},
}

type downloadsListView struct {
	tableModel table.Model
	width      int
	height     int

	downloads []state.ListDownloadsWithQueueNameRow

	queueManager queues.QueueManager
}

func (m *downloadsListView) updateColumnWidths() {
	availableWidth := m.width

	if availableWidth <= 0 {
		return
	}

	columns := m.tableModel.Columns()
	for i, col := range columns {
		if i < len(downloadsColumnRatios) {
			col.Width = int(float64(availableWidth) * downloadsColumnRatios[i])

			columns[i] = col
		}
	}

	m.tableModel.SetColumns(columns)
}

func (m downloadsListView) FullHelp() [][]key.Binding {
	return m.tableModel.KeyMap.FullHelp()
}

func (m downloadsListView) ShortHelp() []key.Binding {
	return m.tableModel.KeyMap.ShortHelp()
}

func (m downloadsListView) Init() tea.Cmd { return nil }

func (m downloadsListView) Update(msg tea.Msg) (types.View, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tableModel.SetWidth(msg.Width)
		m.tableModel.SetHeight(msg.Height)
		m.updateColumnWidths()
	}
	m.tableModel, cmd = m.tableModel.Update(msg)
	return m, cmd
}

func (m downloadsListView) View() string {
	return m.tableModel.View()
}

func NewDownloadsList(ctx context.Context, queueManager queues.QueueManager) (types.View, error) {
	downloads, err := queueManager.ListDownloadsWithQueueName(ctx)
	if err != nil {
		return nil, err
	}

	rows := []table.Row{
		{"url1", "queue1", "progress1"},
		{"url2", "queue2", "progress2"},
		{"url3", "queue3", "progress3"},
	}

	t := table.New(
		table.WithRows(rows),
		table.WithColumns(downloadsColumns),
		table.WithFocused(true),
		table.WithStyles(tableStyles),
	)

	return downloadsListView{
		tableModel:   t,
		queueManager: queueManager,

		downloads: downloads,
	}, nil
}
