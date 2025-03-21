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
			key.WithKeys("n"),
			key.WithHelp("n", "toggle notifications"),
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

// FullHelp implements help.KeyMap.
func (d downloadManagerModel) FullHelp() [][]key.Binding {
	return append([][]key.Binding{{d.keymap.helpToggle, d.keymap.quit, d.keymap.confirmError, d.keymap.toggleNotifications}},
		d.tabsModel.FullHelp()...)
}

// ShortHelp implements help.KeyMap.
func (d downloadManagerModel) ShortHelp() []key.Binding {
	return append([]key.Binding{d.keymap.helpToggle, d.keymap.quit, d.keymap.confirmError, d.keymap.toggleNotifications}, d.tabsModel.ShortHelp()...)
}

func (d downloadManagerModel) Init() tea.Cmd {
	return tea.Batch(d.tabsModel.Init(), listenForEvent)
}

// Update implements tea.Model.
func (d downloadManagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch v := msg.(type) {
	case events.Event:
		var cmds []tea.Cmd

		cmds = append(cmds, listenForEvent)

		if v.EventType == events.DownloadCompleted {
			status := v.Payload.(downloads.DownloadStatus)
			cmds = append(cmds, createCmd(types.NotifMsg{
				Msg: fmt.Sprintf("Download has finished: %s", status.URL),
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
		// Show notifications automatically when a new one arrives
		if len(d.notifications) > 0 && !d.showNotifications {
			d.showNotifications = true
		}
		return d, nil
	case tea.KeyMsg:
		if len(d.generalErrors) > 0 && key.Matches(v, d.keymap.confirmError) {
			d.generalErrors = lo.Slice(d.generalErrors, 0, len(d.generalErrors)-1)
			return d, nil
		} else if key.Matches(v, d.keymap.toggleNotifications) {
			// Toggle notifications view
			d.showNotifications = !d.showNotifications
			return d, nil
		} else if d.showNotifications && len(d.notifications) > 0 && key.Matches(v, d.keymap.dismissNotification) {
			// Dismiss the latest notification
			d.notifications = lo.Slice(d.notifications, 0, len(d.notifications)-1)
			// If no more notifications, hide the notifications view
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
		// Get the latest error message (which will be in cowsay format)
		latestError := d.generalErrors[len(d.generalErrors)-1].View()

		// Get error count
		errorCount := len(d.generalErrors)

		// Style for the error counter
		counterStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#FF0000")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

		// Create the counter display
		counterDisplay := counterStyle.Render(strconv.Itoa(errorCount))

		// Header with counter
		errorHeader := counterDisplay + " Errors"

		// Style and center the header only
		headerStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(d.width) // Use the full width of the terminal

		styledHeader := headerStyle.Render(errorHeader)

		// Add a border around the cowsay error to help center it without distorting
		cowsayStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF0000")).
			Padding(1, 2)

		// Apply the style to the cowsay error
		styledCowsay := cowsayStyle.Render(latestError)

		// Add dismiss instruction below the bordered cowsay
		dismissStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(d.width)

		styledDismiss := dismissStyle.Render(fmt.Sprintf("Press Enter to dismiss. (1 of %d)", errorCount))

		// Create a container style that centers the content as a whole

		containerStyle := lipgloss.NewStyle().
			Width(d.width).
			Align(lipgloss.Center)

		// Apply the container style to the entire error section
		errorSection := containerStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				styledHeader,
				"", // Empty line for spacing
				styledCowsay,
				styledDismiss,
			),
		)

		mainView = errorSection
	} else if d.showNotifications && len(d.notifications) > 0 {
		// Get the latest notification message (which will be in cowsay format)
		latestNotification := d.notifications[len(d.notifications)-1].View()

		// Get notification count
		notificationCount := len(d.notifications)

		// Style for the notification counter
		counterStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#4B8BBE")). // Using a blue color for notifications
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

		// Create the counter display
		counterDisplay := counterStyle.Render(strconv.Itoa(notificationCount))

		// Header with counter
		notificationHeader := counterDisplay + " Notifications"

		// Style and center the header only
		headerStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(100) // Using a large value to ensure it takes full width

		styledHeader := headerStyle.Render(notificationHeader)

		// Add a border around the cowsay notification to help center it without distorting
		cowsayStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#4B8BBE")). // Blue border for notifications
			Padding(1, 2)

		// Apply the style to the cowsay notification
		styledCowsay := cowsayStyle.Render(latestNotification)

		// Add dismiss instruction below the bordered cowsay
		dismissStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(d.width)

		styledDismiss := dismissStyle.Render(fmt.Sprintf("Press Enter to dismiss. (1 of %d) | Press 'n' to hide notifications", notificationCount))

		// Create a container style that centers the content as a whole
		containerStyle := lipgloss.NewStyle().
			Width(d.width).
			Align(lipgloss.Center)

		// Apply the container style to the entire notification section
		notificationSection := containerStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				styledHeader,
				"", // Empty line for spacing
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
