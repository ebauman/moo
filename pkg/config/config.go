package config

type AgentConfig struct {
	KubeConfig string

	ClusterName string
	UseExisting bool
	ID          string

	ServerHostname string

	CACerts string

	CattleConfig
	RancherConfig
}

type CattleConfig struct {
	Namespace  string
	Deployment string
	Daemonset  string
}

type RancherConfig struct {
	URL       string
	AccessKey string
	SecretKey string
	Insecure  bool
	CACerts   string
}

type ServerConfig struct {
	RancherConfig
	TLSCert     string
	TLSKey      string
	HoldTime    int32
	PendingTime int32
	ErrorTime   int32
}
