package main

import (
	"context"
	"fmt"
	agent "github.com/ebauman/moo/pkg/agent"
	"github.com/ebauman/moo/pkg/config"
	"github.com/ebauman/moo/pkg/kubernetes"
	mooLogger "github.com/ebauman/moo/pkg/logger"
	"github.com/ebauman/moo/pkg/rancher"
	"github.com/ebauman/moo/pkg/rpc"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"os"
)

var logger *log.Logger

func main() {
	app := &cli.App{
		Name: "moo",
		Usage: "Auto-registration agent for Rancher imported clusters",
		Action: run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "kubeconfig",
				Usage: "kubeconfig if running outside of cluster",
				EnvVars: []string{"KUBECONFIG"},
			},
			&cli.StringFlag{
				Name: "namespace",
				Usage: "namespace used in registration check",
				Value: "cattle-system",
				EnvVars: []string{"CATTLE_NAMESPACE"},
			},
			&cli.StringFlag{
				Name: "deployment",
				Usage: "name of deployment used in registration check",
				Value: "cattle-cluster-agent",
				EnvVars: []string{"CATTLE_DEPLOYMENT"},
			},
			&cli.StringFlag{
				Name: "daemonset",
				Usage: "name of daemonset used in registration check",
				Value:  "cattle-node-agent",
				EnvVars: []string{"CATTLE_DAEMONSET"},
			},
			&cli.StringFlag{
				Name: "rancher-url",
				Usage:  "url of rancher instance",
				EnvVars: []string{"RANCHER_URL"},
			},
			&cli.StringFlag{
				Name: "rancher-access-key",
				Usage: "access key for rancher",
				EnvVars: []string{"RANCHER_ACCESS_KEY"},
			},
			&cli.StringFlag{
				Name: "rancher-secret-key",
				Usage: "secret key for rancher",
				EnvVars: []string{"RANCHER_SECRET_KEY"},
			},
			&cli.StringFlag{
				Name: "cluster-name",
				Usage: "name of this cluster when registering with rancher",
				EnvVars: []string{"MOO_CLUSTER_NAME"},
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
			&cli.StringFlag{
				Name: "moo-server",
				Usage: "hostname for moo server. specifying this will enable server mode",
				Value: "",
				EnvVars: []string{"MOO_SERVER"},
			},
			&cli.StringFlag {
				Name: "loglevel",
				Usage: "log level (trace, debug, info, warning, error, fatal, panic)",
				Value: "info",
				EnvVars: []string{"LOGLEVEL"},
			},
			&cli.BoolFlag{
				Name: "use-existing-cluster",
				Usage: "if cluster already exists in rancher, use it and import this node",
				Value: false,
				EnvVars: []string{"MOO_USE_EXISTING"},
			},
		},
	}

	err := app.Run(os.Args)
	if err !=  nil {
		log.Fatal(err)
	}
}

func getLogger(ctx *cli.Context) *log.Logger {
	if logger != nil {
		return logger
	}
	
	logger = mooLogger.BuildLogger(ctx.String("loglevel"))

	return logger
}



func buildConfigFromFlags(ctx *cli.Context) *config.AgentConfig {
	cfg := &config.AgentConfig{}

	cfg.KubeConfig = ctx.String("kubeconfig")

	cfg.Namespace = ctx.String("namespace")
	cfg.Deployment = ctx.String("deployment")
	cfg.Daemonset = ctx.String("daemonset")

	cfg.URL = ctx.String("rancher-url")
	cfg.AccessKey = ctx.String("rancher-access-key")
	cfg.SecretKey = ctx.String("rancher-secret-key")
	cfg.ClusterName = ctx.String("cluster-name")
	cfg.CACerts = ctx.String("rancher-cacerts")
	cfg.Insecure = ctx.Bool("rancher-insecure")
	cfg.UseExisting = ctx.Bool("use-existing-cluster")
	cfg.ServerHostname = ctx.String("moo-server")

	return cfg
}

func run(ctx *cli.Context) error {
	// using the ctx, build a context and related items.

	cfg := buildConfigFromFlags(ctx)
	appContext := context.Background()
	logger := getLogger(ctx)

	if ((cfg.URL == "" || cfg.AccessKey == "" || cfg.SecretKey == "") && cfg.ServerHostname == "") || cfg.ServerHostname == "" {
		logger.Fatalf("--moo-server or all of {--rancher-url, --rancher-access-key, --rancher-secret-key} required")
	}
	
	k8sClient, err := kubernetes.NewClient(cfg.KubeConfig, logger, appContext)
	if err != nil {
		return fmt.Errorf("error building kubernetes client: %v", err)
	}

	var ag *agent.Agent
	if cfg.ServerHostname != "" {
		// running in server mode
		conn, err := grpc.Dial(cfg.ServerHostname, grpc.WithInsecure()) // TODO - figure out authentication
		if err != nil {
			logger.Fatalf("error dialing moo server: %v", err)
		}
		defer conn.Close()

		mooClient := rpc.NewMooClient(conn)

		ag = agent.NewServerAgent(cfg, k8sClient, mooClient, appContext, logger)

		ag.ServerReconcile()
	} else {
		rancherClient, err := rancher.NewServer(&cfg.RancherConfig)

		if err != nil {
			return fmt.Errorf("error building rancher client: %v", err)
		}

		ag = agent.NewAgent(cfg, k8sClient, rancherClient, appContext, logger)

		ag.StandaloneReconcile()
	}

	return nil
}
