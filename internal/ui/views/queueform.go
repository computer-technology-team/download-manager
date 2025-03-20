package views

import (
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"

	"github.com/computer-technology-team/download-manager.git/internal/queues"
	"github.com/computer-technology-team/download-manager.git/internal/state"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/buttonrow"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/counterinput"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/directorypicker"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/optionalinput"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/textinput"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/timeinput"
	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

type queueFormInputIndex int

type buttonType string

const (
	name queueFormInputIndex = iota
	bandwidthLimitBPS
	directoryPicker
	maxConcurrentDownload
	startTime
	endTime
	submit

	totalInputs

	submitButton buttonType = "submit"
	cancelButton buttonType = "cancel"

	inputLocationGuide = " ↓"
)

type queueForm struct {
	queueID *int64

	name                  types.Input[string]
	bandwidthLimitBPS     types.Input[*int64]
	directoryPicker       types.Input[string]
	maxConcurrentDownload types.Input[int64]
	startTime             types.Input[types.TimeValue]
	endTime               types.Input[types.TimeValue]

	submit *buttonrow.Model

	focus queueFormInputIndex
	err   error

	queueManager queues.QueueManager
}

type button struct {
	label string
	slug  string
	color lipgloss.TerminalColor
}

func (b button) Label() string {
	return b.label
}

func (b button) Color() lipgloss.TerminalColor {
	return b.color
}

func (v queueForm) FullHelp() [][]key.Binding {
	return [][]key.Binding{{
		key.NewBinding(key.WithKeys("↓", "down"), key.WithHelp("↓", "next field")),
		key.NewBinding(key.WithKeys("↑", "up"), key.WithHelp("↑", "previous field")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit/select")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}, v.keyMappers()[v.focus].ShortHelp()}
}

func (v queueForm) Init() tea.Cmd {
	return tea.Batch(v.initCmds()...)
}

func (v queueForm) ShortHelp() []key.Binding {
	return slices.Concat([]key.Binding{
		key.NewBinding(key.WithKeys("↓", "down"), key.WithHelp("↓", "next field")),
		key.NewBinding(key.WithKeys("↑", "up"), key.WithHelp("↑", "previous field")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit/select")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}, v.keyMappers()[v.focus].ShortHelp())
}

func (v *queueForm) updateFocus() {
	focusables := v.focusables()
	for _, focusable := range focusables {
		focusable.Blur()
	}

	focusables[v.focus].Focus()
}

func (v queueForm) focusables() []types.Focusable {
	return []types.Focusable{
		v.name, v.bandwidthLimitBPS, v.directoryPicker, v.maxConcurrentDownload, v.startTime, v.endTime, v.submit,
	}
}

func (v queueForm) keyMappers() []help.KeyMap {
	return []help.KeyMap{
		v.name, v.bandwidthLimitBPS, v.directoryPicker, v.maxConcurrentDownload, v.startTime, v.endTime, v.submit,
	}
}

func (v queueForm) initCmds() []tea.Cmd {
	return []tea.Cmd{
		v.name.Init(), v.bandwidthLimitBPS.Init(), v.directoryPicker.Init(),
		v.maxConcurrentDownload.Init(), v.startTime.Init(), v.endTime.Init(),
	}
}

func (v *queueForm) nextInput() {
	v.focus = (v.focus + 1) % totalInputs
	v.updateFocus()
}

func (v *queueForm) prevInput() {
	v.focus--
	if v.focus < 0 {
		v.focus = totalInputs - 1
	}
	v.updateFocus()
}

func (v queueForm) View() string {
	var sb strings.Builder

	sb.WriteString("Queue Form\n\n")

	// Name input
	sb.WriteString("Name: ")
	if v.focus == name {
		sb.WriteString(inputLocationGuide)
	}
	sb.WriteString("\n")
	sb.WriteString(v.name.View())
	if err := v.name.Error(); err != nil {
		sb.WriteString(" ⚠️ " + err.Error())
	}
	sb.WriteString("\n")

	// Bandwidth limit input
	sb.WriteString("Bandwidth Limit (Bytes Per Second): ")
	if v.focus == bandwidthLimitBPS {
		sb.WriteString(inputLocationGuide)
	}
	sb.WriteString("\n")
	sb.WriteString(v.bandwidthLimitBPS.View())
	if err := v.bandwidthLimitBPS.Error(); err != nil {
		sb.WriteString(" ⚠️ " + err.Error())
	} else if bwValue := v.bandwidthLimitBPS.Value(); bwValue != nil {
		sb.WriteString(" (" + FormatBytesPerSecond(*bwValue) + ")")
	}
	sb.WriteString("\n")

	// Directory picker
	sb.WriteString("Download Directory: ")
	if v.focus == directoryPicker {
		sb.WriteString(inputLocationGuide)
	}
	sb.WriteString("\n")
	sb.WriteString(v.directoryPicker.View())
	if err := v.directoryPicker.Error(); err != nil {
		sb.WriteString(" ⚠️ " + err.Error())
	}

	sb.WriteString("\n")

	// Max concurrent downloads
	sb.WriteString("Max Concurrent Downloads: ")
	if v.focus == maxConcurrentDownload {
		sb.WriteString(inputLocationGuide)
	}
	sb.WriteString("\n")
	sb.WriteString(v.maxConcurrentDownload.View())
	if err := v.maxConcurrentDownload.Error(); err != nil {
		sb.WriteString(" ⚠️ " + err.Error())
	}
	sb.WriteString("\n")
	// Start time
	sb.WriteString("Start Time: ")
	if v.focus == startTime {
		sb.WriteString(inputLocationGuide)
	}
	sb.WriteString("\n")
	sb.WriteString(v.startTime.View())
	if err := v.startTime.Error(); err != nil {
		sb.WriteString(" ⚠️ " + err.Error())
	}

	sb.WriteString("\n")

	// End time
	sb.WriteString("End Time:\n")
	if v.focus == endTime {
		sb.WriteString(inputLocationGuide)
	}
	sb.WriteString("\n")
	sb.WriteString(v.endTime.View())
	if err := v.endTime.Error(); err != nil {
		sb.WriteString(" ⚠️ " + err.Error())
	}

	sb.WriteString("\n")

	sb.WriteString(v.submit.View())

	// Form-level error message if any
	if v.err != nil {
		sb.WriteString("Error: " + v.err.Error() + "\n\n")
	}

	return sb.String()
}

func createQueueCmd(name string, bandwidthLimit *int64, directory string, maxConcurrent int64, startTime, endTime types.TimeValue) tea.Cmd {
	return func() tea.Msg {
		slog.Info("create queue",
			"name", name,
			"bandwidthLimit", bandwidthLimit,
			"directory", directory,
			"maxConcurrent", maxConcurrent,
			"startTime", startTime,
			"endTime", endTime)
		return nil
	}
}

func (v queueForm) Update(msg tea.Msg) (queueForm, tea.Cmd) {

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if v.focus == submit {
				// Check which button is selected
				selectedButton := v.submit.SelectedButton().(button)
				if selectedButton.slug == string(submitButton) {
					// Validate form before submission
					if v.name.Value() == "" {
						v.err = nil
						return v, nil
					}

					// Create the queue
					return v, createQueueCmd(
						v.name.Value(),
						v.bandwidthLimitBPS.Value(),
						v.directoryPicker.Value(),
						v.maxConcurrentDownload.Value(),
						v.startTime.Value(),
						v.endTime.Value(),
					)
				} else if selectedButton.slug == string(cancelButton) {
					// Cancel form - return to list view
					return v, tea.Quit
				}
			} else {
				v.nextInput()
			}
		case tea.KeyUp:
			v.prevInput()
		case tea.KeyDown:
			v.nextInput()
		default:
			var cmd tea.Cmd
			switch v.focus {
			case name:
				v.name, cmd = v.name.Update(msg)
				cmds = append(cmds, cmd)
			case bandwidthLimitBPS:
				v.bandwidthLimitBPS, cmd = v.bandwidthLimitBPS.Update(msg)
				cmds = append(cmds, cmd)
			case directoryPicker:
				v.directoryPicker, cmd = v.directoryPicker.Update(msg)
				cmds = append(cmds, cmd)
			case maxConcurrentDownload:
				v.maxConcurrentDownload, cmd = v.maxConcurrentDownload.Update(msg)
				cmds = append(cmds, cmd)
			case startTime:
				v.startTime, cmd = v.startTime.Update(msg)
				cmds = append(cmds, cmd)
			case endTime:
				v.endTime, cmd = v.endTime.Update(msg)
				cmds = append(cmds, cmd)
			case submit:
				var buttonModel buttonrow.Model
				buttonModel, cmd = v.submit.Update(msg)
				v.submit = &buttonModel
				cmds = append(cmds, cmd)
			}
		}
	default:
		// Update all inputs and collect their commands
		var nameInput types.Input[string]
		var cmd tea.Cmd
		nameInput, cmd = v.name.Update(msg)
		v.name = nameInput
		cmds = append(cmds, cmd)

		var bwInput types.Input[*int64]
		bwInput, cmd = v.bandwidthLimitBPS.Update(msg)
		v.bandwidthLimitBPS = bwInput
		cmds = append(cmds, cmd)

		var dirInput types.Input[string]
		dirInput, cmd = v.directoryPicker.Update(msg)
		v.directoryPicker = dirInput
		cmds = append(cmds, cmd)

		var maxConcInput types.Input[int64]
		maxConcInput, cmd = v.maxConcurrentDownload.Update(msg)
		v.maxConcurrentDownload = maxConcInput
		cmds = append(cmds, cmd)

		var startTimeInput types.Input[types.TimeValue]
		startTimeInput, cmd = v.startTime.Update(msg)
		v.startTime = startTimeInput
		cmds = append(cmds, cmd)

		var endTimeInput types.Input[types.TimeValue]
		endTimeInput, cmd = v.endTime.Update(msg)
		v.endTime = endTimeInput
		cmds = append(cmds, cmd)

		var buttonModel buttonrow.Model
		buttonModel, cmd = v.submit.Update(msg)
		v.submit = &buttonModel
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func NewQueueCreateForm(queueManager queues.QueueManager) *queueForm {
	nameInput := textinput.New()
	nameInput.Placeholder = "Enter queue name"
	nameInput.Focus()
	nameInput.Width = 40

	bandwidthLimitInput := optionalinput.New(
		counterinput.New(
			counterinput.WithStep(100),
			counterinput.WithMax(1<<63-1),
			counterinput.WithStepHandler(BPSStepHandler)))
	directoryInput := directorypicker.New()

	maxConcurrentInput := counterinput.New()

	startTimeInput := timeinput.New()
	endTimeInput := timeinput.New()

	buttonRow, err := buttonrow.New([]buttonrow.Button{
		button{label: "Submit", slug: string(submitButton), color: lipgloss.Color("#00FF00")},
		button{label: "Cancel", slug: string(cancelButton), color: lipgloss.Color("#FF0000")},
	})
	if err != nil {
		slog.Error("could not create button row", "error", err)
		panic(err)
	}

	qfv := &queueForm{
		name:                  nameInput,
		bandwidthLimitBPS:     bandwidthLimitInput,
		directoryPicker:       directoryInput,
		maxConcurrentDownload: maxConcurrentInput,
		startTime:             startTimeInput,
		endTime:               endTimeInput,
		submit:                buttonRow,
		focus:                 name,

		queueManager: queueManager,
	}

	qfv.updateFocus()

	return qfv
}

func NewQueueEditForm(queue state.Queue, queueManager queues.QueueManager) (*queueForm, error) {
	var err error

	nameInput := textinput.New()
	nameInput.Placeholder = "Enter queue name"
	nameInput.Width = 40
	err = nameInput.SetValue(queue.Name)
	if err != nil {
		return nil, err
	}

	bandwidthLimitInput := optionalinput.New(
		counterinput.New(
			counterinput.WithStep(100),
			counterinput.WithMax(1<<63-1),
			counterinput.WithStepHandler(BPSStepHandler),
		))
	err = bandwidthLimitInput.SetValue(
		lo.Ternary(queue.MaxBandwidth.Valid, &queue.MaxBandwidth.Int64, nil),
	)
	if err != nil {
		return nil, err
	}

	directoryInput := directorypicker.New()
	err = directoryInput.SetValue(queue.Directory)
	if err != nil {
		return nil, err
	}

	maxConcurrentInput := counterinput.New()
	//do this

	startTimeInput := timeinput.New()
	//do this

	endTimeInput := timeinput.New()
	//do this

	buttonRow, err := buttonrow.New([]buttonrow.Button{
		button{label: "Submit", slug: string(submitButton), color: lipgloss.Color("#00FF00")},
		button{label: "Cancel", slug: string(cancelButton), color: lipgloss.Color("#FF0000")},
	})
	if err != nil {
		slog.Error("could not create queue edit form button row", "error", err)
		return nil, err
	}

	qfv := &queueForm{
		name:                  nameInput,
		bandwidthLimitBPS:     bandwidthLimitInput,
		directoryPicker:       directoryInput,
		maxConcurrentDownload: maxConcurrentInput,
		startTime:             startTimeInput,
		endTime:               endTimeInput,
		submit:                buttonRow,
		focus:                 name,
		queueID:               &queue.ID,

		queueManager: queueManager,
	}

	qfv.updateFocus()

	return qfv, nil
}

func BPSStepHandler(_, value int64) int64 {
	switch {
	case value <= 200:
		return 100
	case value < 100000:
		return int64(math.Pow10(int(math.Log10(float64(value))+1))) / 2
	case value < 500000:
		return 100000
	case value < 8000000:
		return 500000
	default:
		return 1000000
	}
}

func FormatBytesPerSecond(bps int64) string {
	const (
		KB float64 = 1024
		MB float64 = KB * 1024
		GB float64 = MB * 1024
	)

	bytesPerSec := float64(bps)

	switch {
	case bytesPerSec >= GB:
		return fmt.Sprintf("%.2f GB/s", bytesPerSec/GB)
	case bytesPerSec >= MB:
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/MB)
	case bytesPerSec >= KB:
		return fmt.Sprintf("%.2f KB/s", bytesPerSec/KB)
	default:
		return fmt.Sprintf("%d B/s", bps)
	}
}
