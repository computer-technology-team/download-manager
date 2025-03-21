package views

import (
	"context"
	"errors"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/events"
	"github.com/computer-technology-team/download-manager.git/internal/queues"
	"github.com/computer-technology-team/download-manager.git/internal/state"
	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

var errNoDownloadAvailable = errors.New("no download is available")

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

type downloadsListKeyMap struct {
	Resume key.Binding
	Pause  key.Binding
	Retry  key.Binding
}

type downloadsListView struct {
	tableModel table.Model
	width      int
	height     int

	keymap downloadsListKeyMap

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

func (m downloadsListView) handleUpdate(msg events.Event) (types.View, tea.Cmd) {
	switch msg.EventType {
	case events.DownloadCreated:

	case events.DownloadDeleted:
		downloadID := msg.Payload.(int64)
		m.downloads = lo.Filter(m.downloads, func(download state.ListDownloadsWithQueueNameRow, _ int) bool {
			return download.ID != downloadID
		})

		m.setTableRows()

		return m, nil
	case events.DownloadProgressed:

	case events.DownloadFailed:

	}

	return m, nil
}

func (m downloadsListView) Update(msg tea.Msg) (types.View, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case events.Event:
		return m.handleUpdate(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tableModel.SetWidth(msg.Width)
		m.tableModel.SetHeight(msg.Height)
		m.updateColumnWidths()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Pause):
			return m, m.pause()
		case key.Matches(msg, m.keymap.Resume):
			return m, m.resume()
		case key.Matches(msg, m.keymap.Retry):
			return m, m.retry()

		}
	}
	m.tableModel, cmd = m.tableModel.Update(msg)
	return m, cmd
}

func (m *downloadsListView) pause() tea.Cmd {
	_, err := m.getUnderCursorDownload()
	if err != nil {
		return createErrorCmd(types.ErrorMsg{Err: err})
	}

	return nil
}

func (m *downloadsListView) resume() tea.Cmd {
	_, err := m.getUnderCursorDownload()
	if err != nil {
		return createErrorCmd(types.ErrorMsg{Err: err})
	}

	return nil
}

func (m *downloadsListView) retry() tea.Cmd {
	_, err := m.getUnderCursorDownload()
	if err != nil {
		return createErrorCmd(types.ErrorMsg{Err: err})
	}

	return nil
}

func (m downloadsListView) View() string {
	return m.tableModel.View()
}

func (m downloadsListView) getUnderCursorDownload() (*state.ListDownloadsWithQueueNameRow, error) {
	if len(m.downloads) == 0 {
		return nil, errNoDownloadAvailable
	}
	return &m.downloads[m.tableModel.Cursor()], nil
}

func (m *downloadsListView) setTableRows() {
	m.tableModel.SetRows(lo.Map(m.downloads, func(download state.ListDownloadsWithQueueNameRow, _ int) table.Row {
		return downloadToDownloadTableRow(download)
	}))
}

func NewDownloadsList(ctx context.Context, queueManager queues.QueueManager) (types.View, error) {
	downloads, err := queueManager.ListDownloadsWithQueueName(ctx)
	if err != nil {
		return nil, err
	}

	rows := lo.Map(downloads, func(d state.ListDownloadsWithQueueNameRow, _ int) table.Row {
		return downloadToDownloadTableRow(d)
	})

	t := table.New(
		table.WithRows(rows),
		table.WithColumns(downloadsColumns),
		table.WithFocused(true),
		table.WithStyles(tableStyles),
	)

	return downloadsListView{
		tableModel:   t,
		queueManager: queueManager,
		keymap:       defaultDownloadsListKeyMap(),

		downloads: downloads,
	}, nil
}

func downloadToDownloadTableRow(download state.ListDownloadsWithQueueNameRow) table.Row {
	return table.Row{download.Url, download.QueueName, download.State}
}

func defaultDownloadsListKeyMap() downloadsListKeyMap {
	return downloadsListKeyMap{
		Resume: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "resume download")),
		Pause:  key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pause download")),
		Retry:  key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "retry download")),
	}
}
