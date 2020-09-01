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

type RancherServer struct {
	client *managementClient.Client

	config.RancherConfig
}

func NewServer(config *config.RancherConfig) (*RancherServer, error) {
	server := &RancherServer{}
	server.CACerts = config.CACerts
	server.Insecure = config.Insecure
	server.SecretKey = config.SecretKey
	server.AccessKey = config.AccessKey
	server.URL = config.URL
	clientOpts := buildConfig(config)

	err := isRancherReady(config)
	if err != nil {
		return nil, err
	}

	mClient, err := managementClient.NewClient(clientOpts)
	if err != nil {
		return nil, err
	}

	server.client = mClient

	return server, nil
}

func (r *RancherServer) reconcile(clusterName string, useExisting bool) (string, error) {
	cluster, err := r.checkForCluster(clusterName)
	if err != nil {
		return "", err
	}

	if cluster != nil && !useExisting {
		return "", fmt.Errorf("cluster %s already exists in rancher and use existing is false", clusterName)
	}

	if cluster == nil {
		// need to create the cluster
		cluster, err = r.registerCluster(clusterName)
		if err != nil {
			return "", fmt.Errorf("error registering cluster with rancher: %v", err)
		}
	}

	return cluster.ID, nil
}

func (r *RancherServer) ReconcileToURL(clusterName string, useExisting bool) (string, error) {
	id, err := r.reconcile(clusterName, useExisting)
	
	if err != nil {
		return "", err
	}
	
	return r.GetManifestURLForCluster(id)
}

func (r *RancherServer) ReconcileToManifest(clusterName string, useExisting bool) ([]byte, error) {
	id, err := r.reconcile(clusterName, useExisting)
	
	if err != nil {
		return nil, err
	}

	manifest, err := r.getYAMLManifestForCluster(id)
	if err != nil {
		return nil, fmt.Errorf("error getting manifest from rancher: %v", err)
	}

	return manifest, nil
}

func (r *RancherServer) checkForCluster(clusterName string) (*managementClient.Cluster, error) {
	// check for the existence of named cluster
	filters := map[string]interface{}{}
	filters["name"] = clusterName
	clusters, err := r.client.Cluster.List(&types.ListOpts{Filters: filters})

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

func (r *RancherServer) registerCluster(clusterName string) (*managementClient.Cluster, error) {
	// create a new cluster
	cluster := &managementClient.Cluster{}

	var f = false
	cluster.Name = clusterName
	cluster.EnableClusterAlerting = false
	cluster.EnableClusterMonitoring = false
	cluster.EnableNetworkPolicy = &f // HACK
	cluster.Type = "cluster"

	created, err := r.client.Cluster.Create(cluster)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *RancherServer) GetManifestURLForCluster(clusterID string) (string, error) {
	token, err := r.client.ClusterRegistrationToken.Create(&managementClient.ClusterRegistrationToken{
		ClusterID: clusterID,
	})

	if err != nil {
		return "", fmt.Errorf("error creating clusterregistrationtoken: %v", err)
	}

	return token.ManifestURL, nil
}

func (r *RancherServer) getYAMLManifestForCluster(clusterID string) ([]byte, error) {
	token, err := r.client.ClusterRegistrationToken.Create(&managementClient.ClusterRegistrationToken{
		ClusterID: clusterID,
	})

	if err != nil {
		return nil, fmt.Errorf("error creating clusterregistrationtoken: %v", err)
	}

	manifest, err := DoGet(token.ManifestURL, "", "", r.CACerts, r.Insecure)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func (r *RancherServer) GetYAMLFromURL(url string) ([]byte, error) {
	manifest, err := DoGet(url, "", "", r.CACerts, r.Insecure)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func buildConfig(config *config.RancherConfig) *clientbase.ClientOpts {
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

func isRancherReady(config *config.RancherConfig) error {
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