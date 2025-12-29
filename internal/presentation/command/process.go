package command

import (
	"context"

	"github.com/drybin/TrackMyCoin/internal/app/cli/usecase"
	"github.com/urfave/cli/v2"
)

func NewProcessCommand(service usecase.IProcess) *cli.Command {
	return &cli.Command{
		Name:  "process",
		Usage: "process command",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			return service.Process(context.Background())
		},
	}
}

