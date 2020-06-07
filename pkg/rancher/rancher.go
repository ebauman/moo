package rancher

import (
	"fmt"
	"github.com/ebauman/moo/pkg/config"
	"github.com/google/martian/log"
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/types"
	managementClient "github.com/rancher/types/client/management/v3"
	"github.com/terraform-providers/terraform-provider-rancher2/rancher2"
	"time"
)

const (
	rancher2ReadyAnswer = "pong"
	rancher2RetriesWait = 5
)

func BuildClient(config *config.AgentConfig) (*managementClient.Client, error) {
	clientOpts := buildConfig(config)

	err := isRancherReady(config)
	if err != nil {
		return nil, err
	}

	mClient, err := managementClient.NewClient(clientOpts)
	if err != nil {
		return nil, err
	}

	return mClient, nil
}

func CheckForCluster(config *config.AgentConfig) (*managementClient.Cluster, error) {
	// check for the existence of named cluster
	filters := map[string]interface{}{}
	filters["name"] = config.ClusterName
	clusters, err := config.Rancher.Cluster.List(&types.ListOpts{Filters: filters})

	if err != nil {
		return nil, err
	}

	if len(clusters.Data) > 0 {
		// there is at least one cluster in existence with this name
		return &clusters.Data[0], nil
		// only return the first which shouldn't be an issue because Rancher
		// enforces unique naming
	}

	return nil, nil // we did not find a cluster with this name
}

func RegisterCluster(config *config.AgentConfig) (*managementClient.Cluster, error) {
	// create a new cluster
	cluster := &managementClient.Cluster{}

	var f = false
	cluster.Name = config.ClusterName
	cluster.EnableClusterAlerting = false
	cluster.EnableClusterMonitoring = false
	cluster.EnableNetworkPolicy = &f // HACK
	cluster.Type = "cluster"

	created, err := config.Rancher.Cluster.Create(cluster)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func GetYAMLManifestForCluster(cluster *managementClient.Cluster, config *config.AgentConfig) ([]byte, error) {
	token, err := config.Rancher.ClusterRegistrationToken.Create(&managementClient.ClusterRegistrationToken{
		ClusterID: cluster.ID,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating clusterregistrationtoken: %v", err)
	}

	manifest, err := DoGet(token.ManifestURL, "", "", config.CACerts, config.Insecure)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func buildConfig(config *config.AgentConfig) *clientbase.ClientOpts {
	log.Debugf("building clientopts")
	log.Debugf("url: %s", config.URL)
	log.Debugf("tokenkey: %s", createTokenKey(config.AccessKey, config.SecretKey))
	log.Debugf("cacerts: %s", config.CACerts)
	log.Debugf("insecure: %v", config.Insecure)
	return &clientbase.ClientOpts{
		URL:      NormalizeURL(config.URL),
		TokenKey: createTokenKey(config.AccessKey, config.SecretKey),
		CACerts:  config.CACerts,
		Insecure: config.Insecure,
	}
}

func createTokenKey(access string, secret string) string {
	return fmt.Sprintf("%s:%s", access, secret)
}

func isRancherReady(config *config.AgentConfig) error {
	var err error
	var resp []byte
	url := rancher2.RootURL(config.URL) + "/ping"
	for i := 0; i <= 5; i++ {
		resp, err = DoGet(url, "", "", config.CACerts, config.Insecure)
		if err == nil && rancher2ReadyAnswer == string(resp) {
			return nil
		}
		time.Sleep(rancher2RetriesWait * time.Second)
	}
	return fmt.Errorf("rancher is not ready: %v", err)
}