package ui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/downloads"
	"github.com/computer-technology-team/download-manager.git/internal/events"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/cowsay"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/generalerror"
	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

type keymap struct {
	helpToggle          key.Binding
	quit                key.Binding
	confirmError        key.Binding
	toggleNotifications key.Binding
	dismissNotification key.Binding
}

func defaultKeyMap() *keymap {
	return &keymap{
		helpToggle: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggles full help"),
		),
		quit: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc/ctrl+c", "quit"),
		),
		confirmError: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm error"),
		),
		toggleNotifications: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "toggle notifications"),
		),
		dismissNotification: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "dismiss notification"),
		),
	}
}

func listenForEvent() tea.Msg {
	return <-events.GetUIEventChannel()
}

type downloadManagerModel struct {
	tabsModel types.View
	helpModel help.Model

	keymap *keymap

	generalErrors     []types.Viewable
	notifications     []types.Viewable
	showNotifications bool
	height            int
	width             int
}

func (d downloadManagerModel) FullHelp() [][]key.Binding {
	return append([][]key.Binding{{d.keymap.helpToggle, d.keymap.quit, d.keymap.confirmError, d.keymap.toggleNotifications}},
		d.tabsModel.FullHelp()...)
}

func (d downloadManagerModel) ShortHelp() []key.Binding {
	return append([]key.Binding{d.keymap.helpToggle, d.keymap.quit, d.keymap.confirmError, d.keymap.toggleNotifications}, d.tabsModel.ShortHelp()...)
}

func (d downloadManagerModel) Init() tea.Cmd {
	return tea.Batch(d.tabsModel.Init(), listenForEvent)
}

func (d downloadManagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch v := msg.(type) {
	case events.Event:
		var cmds []tea.Cmd

		cmds = append(cmds, listenForEvent)

		switch v.EventType {
		case events.DownloadCompleted:
			status := v.Payload.(downloads.DownloadStatus)
			cmds = append(cmds, createCmd(types.NotifMsg{
				Msg: fmt.Sprintf("Download has finished: %s", status.URL),
			}))
		case events.DownloadFailed:
			event := v.Payload.(events.DownloadFailedEvent)
			cmds = append(cmds, createCmd(types.ErrorMsg{
				Err: fmt.Errorf("download from %s failed: %w", event.URL,
					event.Error),
			}))
		}

		d.tabsModel, cmd = d.tabsModel.Update(msg)
		cmds = append(cmds, cmd)

		return d, tea.Batch(cmds...)
	case tea.WindowSizeMsg:
		d.tabsModel, cmd = d.tabsModel.Update(tea.WindowSizeMsg{
			Width:  v.Width,
			Height: v.Height - 6,
		})
		d.height, d.width = v.Height, v.Width
		return d, cmd
	case types.ErrorMsg:
		d.generalErrors = append(d.generalErrors, generalerror.New(v))
		return d, nil
	case types.NotifMsg:
		d.notifications = append(d.notifications, cowsay.New(v.Msg))

		if len(d.notifications) > 0 && !d.showNotifications {
			d.showNotifications = true
		}
		return d, nil
	case tea.KeyMsg:
		if len(d.generalErrors) > 0 && key.Matches(v, d.keymap.confirmError) {
			d.generalErrors = lo.Slice(d.generalErrors, 0, len(d.generalErrors)-1)
			return d, nil
		} else if key.Matches(v, d.keymap.toggleNotifications) {

			d.showNotifications = !d.showNotifications
			return d, nil
		} else if d.showNotifications && len(d.notifications) > 0 && key.Matches(v, d.keymap.dismissNotification) {

			d.notifications = lo.Slice(d.notifications, 0, len(d.notifications)-1)

			if len(d.notifications) == 0 {
				d.showNotifications = false
			}
			return d, nil
		} else if key.Matches(v, d.keymap.quit) {
			return d, tea.Quit
		} else if key.Matches(v, d.keymap.helpToggle) {
			d.helpModel.ShowAll = !d.helpModel.ShowAll
			return d, nil
		}
		if len(d.generalErrors) > 0 || d.showNotifications {
			return d, nil
		}
	}

	d.tabsModel, cmd = d.tabsModel.Update(msg)
	return d, cmd
}

func (d downloadManagerModel) View() string {
	helpView := d.helpModel.View(d)

	var mainView string
	if len(d.generalErrors) > 0 {

		latestError := d.generalErrors[len(d.generalErrors)-1].View()

		errorCount := len(d.generalErrors)

		counterStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#FF0000")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

		counterDisplay := counterStyle.Render(strconv.Itoa(errorCount))

		errorHeader := counterDisplay + " Errors"

		headerStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(d.width) 

		styledHeader := headerStyle.Render(errorHeader)

		cowsayStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF0000")).
			Padding(1, 2)

		styledCowsay := cowsayStyle.Render(latestError)

		dismissStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(d.width)

		styledDismiss := dismissStyle.Render(fmt.Sprintf("Press Enter to dismiss. (1 of %d)", errorCount))

		containerStyle := lipgloss.NewStyle().
			Width(d.width).
			Align(lipgloss.Center)

		errorSection := containerStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				styledHeader,
				"", 
				styledCowsay,
				styledDismiss,
			),
		)

		mainView = errorSection
	} else if d.showNotifications && len(d.notifications) > 0 {

		latestNotification := d.notifications[len(d.notifications)-1].View()

		notificationCount := len(d.notifications)

		counterStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#4B8BBE")). 
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

		counterDisplay := counterStyle.Render(strconv.Itoa(notificationCount))

		notificationHeader := counterDisplay + " Notifications"

		headerStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(100) 

		styledHeader := headerStyle.Render(notificationHeader)

		cowsayStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4B8BBE")). 
			Padding(1, 2)

		styledCowsay := cowsayStyle.Render(latestNotification)

		dismissStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(d.width)

		styledDismiss := dismissStyle.Render(fmt.Sprintf("Press Enter to dismiss. (1 of %d) | Press 'n' to hide notifications", notificationCount))

		containerStyle := lipgloss.NewStyle().
			Width(d.width).
			Align(lipgloss.Center)

		notificationSection := containerStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				styledHeader,
				"", 
				styledCowsay,
				styledDismiss,
			),
		)

		mainView = notificationSection
	} else {
		mainView = d.tabsModel.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, mainView, helpView)
}

func newDownloadManagerViewModel(tabsModel types.View) downloadManagerModel {
	helpModel := help.New()
	helpModel.ShowAll = true

	return downloadManagerModel{
		tabsModel:         tabsModel,
		keymap:            defaultKeyMap(),
		helpModel:         helpModel,
		showNotifications: false,
	}
}

func createCmd(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
