package views

import (
	"fmt"
	"math"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

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

func createErrorCmd(errMsg types.ErrorMsg) tea.Cmd {
	return func() tea.Msg {
		return errMsg
	}
}

func createCmd(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
