package generalerror

import (
	"fmt"
	"log/slog"

	cowsay "github.com/Code-Hex/Neo-cowsay/v2"

	"github.com/computer-technology-team/download-manager.git/internal/ui/types"
)

type generalErrorModel struct {
	err types.ErrorMsg
}

func (g generalErrorModel) View() string {

	say, err := cowsay.Say(
		g.err.Error(),
		cowsay.BallonWidth(40),
		cowsay.Type("www"), cowsay.Eyes("XX"))

	if err != nil {
		slog.Error("could not render error cowsay", "error", err)

		return fmt.Sprintf("Error: %s", g.err.Error())
	}

	return say
}

func New(err types.ErrorMsg) types.Viewable {
	return generalErrorModel{
		err: err,
	}
}
