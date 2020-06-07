package reconcile

import (
	"github.com/ebauman/moo/pkg/config"
	"github.com/ebauman/moo/pkg/kubernetes"
	"github.com/ebauman/moo/pkg/rancher"
)

type RegistrationStatus string

const (
	Unregistered RegistrationStatus = "Unregistered"
	Registered   RegistrationStatus = "Registered"
)

func Reconcile(config *config.Config) {
	config.Log.Debugf("starting cluster registration")
	var registrationStatus RegistrationStatus

	// start by checking registration status
	registrationStatus = checkRegistration(config)

	if registrationStatus == Registered {
		config.Log.Debugf("cluster already registered")
		return // sleep for 5 minutes, then reconcile again
	}

	config.Log.Infof("registering cluster with rancher %s and name %s", config.URL, config.ClusterName)

	// unregistered at this point, which is where the work needs to be done.

	// once we have a rancher client, check for existence of this cluster
	cluster, err := rancher.CheckForCluster(config)
	if err != nil {
		config.Log.Errorf("error checking for cluster: %v", err)
		return
	}

	if cluster != nil && !config.UseExisting {
		config.Log.Errorf("cluster %s already exists in rancher and --use-existing-cluster not passed", config.ClusterName)
		return
	}

	if cluster == nil {
		// need to create cluster
		cluster, err = rancher.RegisterCluster(config)
		if err != nil {
			config.Log.Errorf("error registering cluster with rancher: %v", err)
			return
		}
	}

	// at this point we should have a valid cluster either by creation or config.UseExisting = true

	// grab the manifests
	manifest, err := rancher.GetYAMLManifestForCluster(cluster, config)
	if err != nil {
		config.Log.Errorf("error getting manifest from rancher: %v", err)
		return
	}

	// with this manifest, apply to local cluster
	if err := kubernetes.ApplyManifest(config, manifest); err != nil {
		config.Log.Errorf("error applying rancher import manifest: %v", err)
		return
	}

	config.Log.Info("cluster registered successfully")
}

func checkRegistration(config *config.Config) RegistrationStatus {
	config.Log.Debugf("checking cluster registration status")
	// check if our cluster is registered or not.
	// registration is achieved when three things occur:
	// 1. creation of the cattle-system namespace
	// 2. creation of the cattle-cluster-agent deployment
	// 3. creation of the cattle-node-agent daemonset

	if ok, _ := kubernetes.CheckForNamespace(config); !ok {
		// we do not have the namespace we need
		config.Log.Debugf("namespace %s not found", config.Namespace)
		return Unregistered
	}

	if ok, _ := kubernetes.CheckForDeployment(config); !ok {
		config.Log.Debugf("deployment %s in namespace %s not found", config.Deployment, config.Namespace)
		return Unregistered
	}

	if ok, _ := kubernetes.CheckForDaemonset(config); !ok {
		config.Log.Debugf("daemonset %s in namespace %s not found", config.Daemonset, config.Namespace)
		return Unregistered
	}

	return Registered // everything passed our checks
}
