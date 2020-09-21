package agent

import (
	"github.com/ebauman/moo/pkg/rpc"
	"github.com/urfave/cli/v2"
)

func LoadComment() *cli.Command{
	return &cli.Command{
		Name: "agent",
		Usage: "options for agents",
		Subcommands: []*cli.Command{
			{
				Name: "list",
				Usage: "list agents",
				Action: listAgents,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "status",
						Usage: "agent status (unknown, pending, held, accepted, denied, error",
					},
				},
			},
		},
	}
}

func listAgents(c *cli.Context) error {
	mooClient, _, err := rpc.SetupClients(c.String("server"), c.Bool("insecure"))
	if err != nil {
		return err
	}


}