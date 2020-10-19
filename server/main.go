package main

import (
	"crypto/tls"
	"github.com/ebauman/moo/pkg/config"
	mooLogger "github.com/ebauman/moo/pkg/logger"
	"github.com/ebauman/moo/pkg/rancher"
	mooServer "github.com/ebauman/moo/pkg/server"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"os"
	"sync"
)

var logger *log.Logger

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
			&cli.IntFlag{
				Name: "hold-time",
				Usage: "time in seconds for agents to backoff when they are in hold status",
				Value: 300, // 5 minutes
				EnvVars: []string{"MOO_HOLD_TIME"},
			},
			&cli.IntFlag{
				Name: "pending-time",
				Usage: "time in seconds for agents to backoff when they are in pending status",
				Value: 30,
				EnvVars: []string{"MOO_PENDING_TIME"},
			},
			&cli.IntFlag{
				Name: "error-time",
				Usage: "time in seconds for agents to backoff when they are in error status",
				Value: 600, // 10 minutes
				EnvVars: []string{"MOO_ERROR_TIME"},
			},
			&cli.StringFlag{
				Name: "loglevel",
				Usage: "log level (trace, debug, info, warning, error, fatal, panic)",
				Value: "info",
				EnvVars: []string{"LOGLEVEL"},
			},
			&cli.StringFlag{
				Name: "tls-cert",
				Required: true,
				Usage: "path to certificate file to secure client communications",
				EnvVars: []string{"MOO_TLS_CERT"},
			},
			&cli.StringFlag{
				Name: "tls-key",
				Required: true,
				Usage: "path to key file to secure client communications",
				EnvVars: []string{"MOO_TLS_KEY"},
			},
		},
	}

	err := app.Run(os.Args)

	if err != nil {
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

func buildConfigFromFlags(ctx *cli.Context) *config.ServerConfig {
	cfg := &config.ServerConfig{}

	cfg.URL = ctx.String("rancher-url")
	cfg.AccessKey = ctx.String("rancher-access-key")
	cfg.SecretKey = ctx.String("rancher-secret-key")
	cfg.Insecure = ctx.Bool("rancher-insecure")
	cfg.CACerts = ctx.String("rancher-cacerts")
	cfg.HoldTime = int32(ctx.Int("hold-time"))
	cfg.PendingTime = int32(ctx.Int("pending-time"))
	cfg.ErrorTime = int32(ctx.Int("error-time"))
	cfg.TLSCert = ctx.String("tls-cert")
	cfg.TLSKey = ctx.String("tls-key")

	return cfg
}

func loadTLSCredentials(certPath string, keyPath string) (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth: tls.NoClientCert,
	}

	return credentials.NewTLS(cfg), nil
}

func run(ctx *cli.Context) error {
	logger = getLogger(ctx)
	cfg := buildConfigFromFlags(ctx)

	r, err := rancher.NewServer(&cfg.RancherConfig)
	if err != nil {
		logger.Fatalf("error building rancher server: %v", err)
	}

	tlsCreds, err := loadTLSCredentials(cfg.TLSCert, cfg.TLSKey)
	if err != nil {
		logger.Fatalf("error loading tls credentials: ", err)
	}

	rpc := grpc.NewServer(grpc.Creds(tlsCreds))

	server := mooServer.NewServer(cfg, r, logger, rpc)

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		logger.Fatalf("failed to create net listener: %v", err)
	}

	go rpc.Serve(lis) // is this right?

	var wg sync.WaitGroup

	wg.Add(1)

	go server.Run(&wg)

	wg.Wait()

	return nil
}