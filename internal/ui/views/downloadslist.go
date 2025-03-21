package views

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/downloads"
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
	Delete key.Binding
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
	return append([][]key.Binding{{m.keymap.Pause, m.keymap.Resume, m.keymap.Retry, m.keymap.Delete}}, m.tableModel.KeyMap.FullHelp()...)
}

func (m downloadsListView) ShortHelp() []key.Binding {
	return append(m.tableModel.KeyMap.ShortHelp(), m.keymap.Pause, m.keymap.Resume, m.keymap.Retry, m.keymap.Delete)
}

func (m downloadsListView) Init() tea.Cmd { return nil }

func (m downloadsListView) handleUpdate(msg events.Event) (types.View, tea.Cmd) {
	switch msg.EventType {
	case events.DownloadCreated:
		download := msg.Payload.(state.ListDownloadsWithQueueNameRow)
		m.downloads = append(m.downloads, download)

		m.setTableRows()

		return m, nil
	case events.DownloadDeleted:
		downloadID := msg.Payload.(int64)
		m.downloads = lo.Filter(m.downloads, func(download state.ListDownloadsWithQueueNameRow, _ int) bool {
			return download.ID != downloadID
		})

		m.setTableRows()

		return m, nil
	case events.DownloadProgressed:
		status := msg.Payload.(downloads.DownloadStatus)

		for i, download := range m.downloads {
			if download.ID == status.ID {
				m.downloads[i].State = fmt.Sprintf("%f - %s", status.ProgressPercentage,
					FormatBytesPerSecond(int64(status.Speed)))
			}
		}

		m.setTableRows()

		return m, nil

	case events.DownloadStateChanged:
		stateChange := msg.Payload.(state.SetDownloadStateParams)
		for i, download := range m.downloads {
			if download.ID == stateChange.ID {
				m.downloads[i].State = stateChange.State
			}
		}

		m.setTableRows()

		return m, nil
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
		case key.Matches(msg, m.keymap.Delete):
			return m, m.delete()

		}
	}
	m.tableModel, cmd = m.tableModel.Update(msg)
	return m, cmd
}

func (m *downloadsListView) delete() tea.Cmd {
	download, err := m.getUnderCursorDownload()
	if err != nil {
		return createErrorCmd(types.ErrorMsg{Err: err})
	}

	return func() tea.Msg {
		err := m.queueManager.DeleteDownload(context.Background(), download.ID)
		if err != nil {
			return types.ErrorMsg{
				Err: fmt.Errorf("could not resume download: %w", err),
			}
		}

		return nil
	}
}

func (m *downloadsListView) pause() tea.Cmd {
	download, err := m.getUnderCursorDownload()
	if err != nil {
		return createErrorCmd(types.ErrorMsg{Err: err})
	}

	return func() tea.Msg {
		err := m.queueManager.PauseDownload(context.Background(), download.ID)
		if err != nil {
			return types.ErrorMsg{
				Err: fmt.Errorf("could not resume download: %w", err),
			}
		}

		return nil
	}
}

func (m *downloadsListView) resume() tea.Cmd {
	download, err := m.getUnderCursorDownload()
	if err != nil {
		return createErrorCmd(types.ErrorMsg{Err: err})
	}

	return func() tea.Msg {
		err := m.queueManager.ResumeDownload(context.Background(), download.ID)
		if err != nil {
			return types.ErrorMsg{
				Err: fmt.Errorf("could not resume download: %w", err),
			}
		}

		return nil
	}
}

func (m *downloadsListView) retry() tea.Cmd {
	download, err := m.getUnderCursorDownload()
	if err != nil {
		return createErrorCmd(types.ErrorMsg{Err: err})
	}

	return func() tea.Msg {
		err := m.queueManager.RetryDownload(context.Background(), download.ID)
		if err != nil {
			return types.ErrorMsg{
				Err: fmt.Errorf("could not resume download: %w", err),
			}
		}

		return nil
	}
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

	t := table.New(
		table.WithColumns(downloadsColumns),
		table.WithFocused(true),
		table.WithStyles(tableStyles),
	)

	dv := downloadsListView{
		tableModel:   t,
		queueManager: queueManager,
		keymap:       defaultDownloadsListKeyMap(),

		downloads: downloads,
	}

	dv.setTableRows()

	return dv, nil
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
