package main

import (
	"context"
	"fmt"
	"github.com/ebauman/moo/agent/reconcile"
	"github.com/ebauman/moo/pkg/config"
	"github.com/ebauman/moo/pkg/kubernetes"
	"github.com/ebauman/moo/pkg/rancher"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

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

func buildConfigFromFlags(ctx *cli.Context) *config.Config {
	cfg := &config.Config{}

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

	cfg.Log = log.New()

	// TODO - this sucks
	switch ctx.String("loglevel") {
	case "trace":
		cfg.Log.SetLevel(log.TraceLevel)
		break
	case "debug":
		cfg.Log.SetLevel(log.DebugLevel)
		break
	case "info":
		cfg.Log.SetLevel(log.InfoLevel)
		break
	case "warning":
		cfg.Log.SetLevel(log.WarnLevel)
		break
	case "error":
		cfg.Log.SetLevel(log.ErrorLevel)
		break
	case "fatal":
		cfg.Log.SetLevel(log.FatalLevel)
		break
	case "panic":
		cfg.Log.SetLevel(log.PanicLevel)
		break
	default:
		cfg.Log.SetLevel(log.InfoLevel)
	}

	return cfg
}

func run(ctx *cli.Context) error {
	// using the ctx, build a context and related items.

	cfg := buildConfigFromFlags(ctx)

	cfg.Context = context.Background()

	k8sClient, dynClient, err := kubernetes.BuildClients(cfg.KubeConfig)
	if err != nil {
		return fmt.Errorf("error building kubernetes client: %v", err)
	}

	cfg.Kubernetes = k8sClient
	cfg.Dynamic = dynClient

	rancherClient, err := rancher.BuildClient(cfg)

	if err != nil {
		return fmt.Errorf("error building rancher client: %v", err)
	}

	cfg.Rancher = rancherClient

	// run the reconciliation

	reconcile.Reconcile(cfg)


	return nil
}
