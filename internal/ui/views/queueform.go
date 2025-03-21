package views

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
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
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/startendtimeinput"
	"github.com/computer-technology-team/download-manager.git/internal/ui/components/textinput"
	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

var ErrEmptyQueueFormName = errors.New("queue name can not be empty")

type queueFormInputIndex int

type buttonType string

type queueFormError struct {
	error
}

const (
	name queueFormInputIndex = iota
	bandwidthLimitBPS
	directoryPicker
	maxConcurrentDownload
	retryLimit
	startEndTime
	submit

	totalInputs

	submitButton buttonType = "submit"
	cancelButton buttonType = "cancel"

	inputLocationGuide = " ↓"

	defaultRetryLimit = 3
	maxRetryLimit     = 10
)

type queueForm struct {
	queueID *int64

	name                  types.Input[string]
	bandwidthLimitBPS     types.Input[*int64]
	directoryPicker       types.Input[string]
	maxConcurrentDownload types.Input[int64]
	retryLimit            types.Input[int64]
	startEndTime          types.Input[*startendtimeinput.StartEndTime]

	submit *buttonrow.Model

	focus queueFormInputIndex
	err   error

	queueManager queues.QueueManager
	onClose      tea.Cmd
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
		v.name, v.bandwidthLimitBPS, v.directoryPicker, v.maxConcurrentDownload, v.retryLimit, v.startEndTime, v.submit,
	}
}

func (v queueForm) keyMappers() []help.KeyMap {
	return []help.KeyMap{
		v.name, v.bandwidthLimitBPS, v.directoryPicker, v.maxConcurrentDownload, v.retryLimit, v.startEndTime, v.submit,
	}
}

func (v queueForm) initCmds() []tea.Cmd {
	return []tea.Cmd{
		v.name.Init(), v.bandwidthLimitBPS.Init(), v.directoryPicker.Init(),
		v.maxConcurrentDownload.Init(), v.retryLimit.Init(), v.startEndTime.Init(),
	}
}

// inputsError returns a joined error of all input validation errors
func (v queueForm) inputsError() error {
	return errors.Join(
		v.name.Error(),
		v.bandwidthLimitBPS.Error(),
		v.directoryPicker.Error(),
		v.maxConcurrentDownload.Error(),
		v.retryLimit.Error(),
		v.startEndTime.Error(),
	)
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

	// Retry limit
	sb.WriteString("Retry Limit: ")
	if v.focus == retryLimit {
		sb.WriteString(inputLocationGuide)
	}
	sb.WriteString("\n")
	sb.WriteString(v.retryLimit.View())
	if err := v.retryLimit.Error(); err != nil {
		sb.WriteString(" ⚠️ " + err.Error())
	}
	sb.WriteString("\n")

	// Start time
	sb.WriteString("Schedule: ")
	if v.focus == startEndTime {
		sb.WriteString(inputLocationGuide)
	}
	sb.WriteString("\n")
	sb.WriteString(v.startEndTime.View())
	if err := v.startEndTime.Error(); err != nil {
		sb.WriteString(" ⚠️ " + err.Error())
	}

	sb.WriteString("\n")

	sb.WriteString(v.submit.View())

	// Form-level error message if any
	if v.err != nil {
		sb.WriteString("\nError: " + v.err.Error() + "\n\n")
	}

	return sb.String()
}

func (v queueForm) createQueueCmd(name string, bandwidthLimit *int64, directory string, maxConcurrent int64, retryLimit int64, startEndTime *startendtimeinput.StartEndTime) tea.Cmd {
	inputErrs := v.inputsError()
	return func() tea.Msg {
		if inputErrs != nil {
			return queueFormError{
				error: fmt.Errorf("some input validation have failed: %w", inputErrs),
			}
		}

		queueParam := state.CreateQueueParams{
			Name:      name,
			Directory: directory,
			MaxBandwidth: sql.NullInt64{
				Valid: bandwidthLimit != nil,
				Int64: lo.FromPtr(bandwidthLimit),
			},
			RetryLimit:    retryLimit,
			MaxConcurrent: maxConcurrent,
		}
		if startEndTime != nil {
			queueParam.StartDownload, queueParam.EndDownload = startEndTime.A, startEndTime.B
			queueParam.ScheduleMode = true
		} else {
			queueParam.ScheduleMode = false
		}

		err := v.queueManager.CreateQueue(context.Background(), queueParam)
		if err != nil {
			return queueFormError{
				error: err,
			}
		}

		return v.onClose()
	}
}

func (v queueForm) updateQueueCmd(id int64, name string, bandwidthLimit *int64, directory string, maxConcurrent int64, retryLimit int64, startEndTime *startendtimeinput.StartEndTime) tea.Cmd {
	inputErrs := v.inputsError()
	return func() tea.Msg {
		if inputErrs != nil {
			return queueFormError{
				error: fmt.Errorf("some input validation have failed: %w", inputErrs),
			}
		}

		queueParam := state.UpdateQueueParams{
			Name:      name,
			Directory: directory,
			MaxBandwidth: sql.NullInt64{
				Valid: bandwidthLimit != nil,
				Int64: lo.FromPtr(bandwidthLimit),
			},
			RetryLimit:    retryLimit,
			MaxConcurrent: maxConcurrent,
			ID:            id,
		}
		if startEndTime != nil {
			queueParam.StartDownload, queueParam.EndDownload = startEndTime.A, startEndTime.B
			queueParam.ScheduleMode = true
		} else {
			queueParam.ScheduleMode = false
		}

		err := v.queueManager.EditQueue(context.Background(), queueParam)
		if err != nil {
			return queueFormError{
				error: err,
			}
		}

		return v.onClose()
	}
}
func (v queueForm) Update(msg tea.Msg) (queueForm, tea.Cmd) {

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case queueFormError:
		v.err = msg.error
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if v.focus == submit {
				// Check which button is selected
				selectedButton := v.submit.SelectedButton().(button)
				if selectedButton.slug == string(submitButton) {

					if v.queueID == nil {
						return v, v.createQueueCmd(
							v.name.Value(),
							v.bandwidthLimitBPS.Value(),
							v.directoryPicker.Value(),
							v.maxConcurrentDownload.Value(),
							v.retryLimit.Value(),
							v.startEndTime.Value(),
						)
					} else {
						return v, v.updateQueueCmd(
							*v.queueID,
							v.name.Value(),
							v.bandwidthLimitBPS.Value(),
							v.directoryPicker.Value(),
							v.maxConcurrentDownload.Value(),
							v.retryLimit.Value(),
							v.startEndTime.Value(),
						)
					}
				} else if selectedButton.slug == string(cancelButton) {
					return v, v.onClose
				}
			} else {
				v.nextInput()
			}
		case tea.KeyUp:
			v.prevInput()
		case tea.KeyDown:
			v.nextInput()
		case tea.KeyEsc:
			return v, v.onClose
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
			case retryLimit:
				v.retryLimit, cmd = v.retryLimit.Update(msg)
				cmds = append(cmds, cmd)
			case startEndTime:
				v.startEndTime, cmd = v.startEndTime.Update(msg)
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

		var retryLimitInput types.Input[int64]
		retryLimitInput, cmd = v.retryLimit.Update(msg)
		v.retryLimit = retryLimitInput
		cmds = append(cmds, cmd)

		var startTimeInput types.Input[*startendtimeinput.StartEndTime]
		startTimeInput, cmd = v.startEndTime.Update(msg)
		v.startEndTime = startTimeInput
		cmds = append(cmds, cmd)

		var buttonModel buttonrow.Model
		buttonModel, cmd = v.submit.Update(msg)
		v.submit = &buttonModel
		cmds = append(cmds, cmd)
	}

	return v, tea.Batch(cmds...)
}

func NewQueueCreateForm(queueManager queues.QueueManager, onClose tea.Cmd) *queueForm {
	nameInput := queueFormNameInput()

	bandwidthLimitInput := newQueueFormBandwidthLimitInput()

	directoryInput := directorypicker.New()

	maxConcurrentInput := counterinput.New()

	retryLimitInput := counterinput.New(
		counterinput.WithMax(maxRetryLimit))

	startEndTimeInput := optionalinput.New(startendtimeinput.New())

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
		retryLimit:            retryLimitInput,
		startEndTime:          startEndTimeInput,
		submit:                buttonRow,
		focus:                 name,

		queueManager: queueManager,
		onClose:      onClose,
	}

	qfv.updateFocus()

	return qfv
}

func NewQueueEditForm(queue state.Queue,
	queueManager queues.QueueManager, onClose tea.Cmd) (*queueForm, error) {
	var err error

	nameInput := queueFormNameInput()
	err = nameInput.SetValue(queue.Name)
	if err != nil {
		return nil, err
	}

	bandwidthLimitInput := newQueueFormBandwidthLimitInput()
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
	err = maxConcurrentInput.SetValue(queue.MaxConcurrent)
	if err != nil {
		return nil, err
	}

	retryLimitInput := counterinput.New(
		counterinput.WithMax(maxRetryLimit))
	err = retryLimitInput.SetValue(queue.RetryLimit)
	if err != nil {
		return nil, err
	}

	startTimeInput := optionalinput.New(startendtimeinput.New())
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
		retryLimit:            retryLimitInput,
		startEndTime:          startTimeInput,
		submit:                buttonRow,
		focus:                 name,
		queueID:               &queue.ID,

		queueManager: queueManager,
		onClose:      onClose,
	}

	qfv.updateFocus()

	return qfv, nil
}

func newQueueFormBandwidthLimitInput() *optionalinput.OptionalInput[int64] {
	bandwidthLimitInput := optionalinput.New(
		counterinput.New(
			counterinput.WithStep(100),
			counterinput.WithMax(1<<63-1),
			counterinput.WithStepHandler(BPSStepHandler),
		))
	return bandwidthLimitInput
}

func queueFormNameInput() textinput.Model {
	nameInput := textinput.New()
	nameInput.Err = ErrEmptyQueueFormName
	nameInput.Placeholder = "Enter queue name"
	nameInput.Focus()
	nameInput.Validate = func(s string) error {
		if s == "" {
			return ErrEmptyQueueFormName
		}
		return nil
	}
	nameInput.Width = 40
	return nameInput
}
