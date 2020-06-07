package config

import (
	"context"
	managementClient "github.com/rancher/types/client/management/v3"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	log "github.com/sirupsen/logrus"
)

type AgentConfig struct {
	KubeConfig string

	Kubernetes *kubernetes.Clientset
	Dynamic dynamic.Interface
	Rancher *managementClient.Client
	Context context.Context

	Log *log.Logger

	ClusterName string
	UseExisting bool

	CattleConfig
	RancherConfig
}

type CattleConfig struct {
	Namespace string
	Deployment string
	Daemonset string
}

type RancherConfig struct {
	URL string
	AccessKey string
	SecretKey string
	Insecure bool
	CACerts string
}

type ServerConfig struct {
	RancherConfig
}
