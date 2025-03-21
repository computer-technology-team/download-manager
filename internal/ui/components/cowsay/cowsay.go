package cowsay

import (
	"log/slog"

	cowsay "github.com/Code-Hex/Neo-cowsay/v2"

	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

type Model struct {
	msg string
}

func (m Model) View() string {

	say, err := cowsay.Say(
		m.msg,
		cowsay.BallonWidth(40),
		cowsay.Type("default"))

	if err != nil {
		slog.Error("could not render cowsay", "error", err)

		return ""
	}

	return say
}

func New(msg string) types.Viewable {
	return Model{
		msg: msg,
	}
}
