package main

import (
	"github.com/ebauman/moo/pkg/agent"
	"github.com/ebauman/moo/pkg/config"
	mooServer "github.com/ebauman/moo/pkg/server"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

func main() {
	app := &cli.App{
		Name: "moo-server",
		Usage: "Auto-registration server for Rancher imported clusters",
		Action: run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "rancher-url",
				Usage:  "url of rancher instance",
				EnvVars: []string{"RANCHER_URL"},
				Required: true,
			},
			&cli.StringFlag{
				Name: "rancher-access-key",
				Usage: "access key for rancher",
				EnvVars: []string{"RANCHER_ACCESS_KEY"},
				Required: true,
			},
			&cli.StringFlag{
				Name: "rancher-secret-key",
				Usage: "secret key for rancher",
				EnvVars: []string{"RANCHER_SECRET_KEY"},
				Required: true,
			},
			&cli.BoolFlag{
				Name: "rancher-insecure",
				Usage: "use an insecure connection to rancher",
				Value: false,
				EnvVars: []string{"RANCHER_INSECURE"},
			},
			&cli.StringFlag{
				Name: "rancher-cacerts",
				Usage: "path to cacerts file used when connecting to rancher",
				EnvVars: []string{"RANCHER_CA_CERTS"},
			},
		},
	}

	err := app.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func buildConfigFromFlags(ctx *cli.Context) *config.ServerConfig {
	cfg := &config.ServerConfig{}

	cfg.URL = ctx.String("rancher-url")
	cfg.AccessKey = ctx.String("rancher-access-key")
	cfg.SecretKey = ctx.String("rancher-secret-key")
	cfg.Insecure = ctx.Bool("rancher-insecure")
	cfg.CACerts = ctx.String("rancher-cacerts")

	return cfg
}

func registerAgent(a *agent.Agent) error {
	return nil
	// TODO - implement
}

func run(ctx *cli.Context) error {
	// build a server
	server := mooServer.New()
	for {
		// get all pending agents
		pending := server.ListAgentsByStatus(agent.StatusPending)

		// register all pending agents
		for _, v := range pending {
			// register
			err := registerAgent(v)
			if err != nil {
				v.Status = agent.StatusError
				v.StatusMessage = err.Error()
				server.UpdateAgent(v)
				continue
			}

			v.Status = agent.StatusAccepted
			v.StatusMessage = "agent accepted"
			server.UpdateAgent(v)
		}

		// sleep
		time.Sleep(time.Second * 30) // TODO - make this configurable
	}
}