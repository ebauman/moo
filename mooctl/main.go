package main

import (
	"github.com/ebauman/moo/mooctl/cmd/agent"
	"github.com/ebauman/moo/mooctl/cmd/rule"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App {
		Name: "mooctl",
		Usage: "manage moo servers",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "server",
				Usage: "moo server hostname",
				EnvVars: []string{"MOO_SERVER"},
			},
			&cli.BoolFlag{
				Name: "insecure",
				Usage: "insecure connection to moo server",
				EnvVars: []string{"MOO_SERVER_INSECURE"},
			},
			&cli.StringFlag{
				Name: "cacerts",
				Usage: "path to file containing ca certificate(s) (PEM format)",
				EnvVars: []string{"MOO_SERVER_CACERTS"},
			},
		},
		Commands: []*cli.Command{
			agent.LoadCommand(),
			rule.LoadCommand(),
		},
	}

	app.Before = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}





