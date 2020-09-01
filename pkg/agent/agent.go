package agent

import (
	"context"
	"github.com/ebauman/moo/pkg/config"
	"github.com/ebauman/moo/pkg/kubernetes"
	"github.com/ebauman/moo/pkg/rancher"
	"github.com/ebauman/moo/pkg/rpc"
	"github.com/hashicorp/go-uuid"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type RegistrationStatus string

const (
	Unregistered RegistrationStatus = "Unregistered"
	Registered   RegistrationStatus = "Registered"
)

type Agent struct {
	kubernetes *kubernetes.KubernetesClient
	rancher    *rancher.RancherServer
	context    context.Context
	config     *config.AgentConfig

	mooClient rpc.MooClient

	log *log.Logger
}

func NewServerAgent(config *config.AgentConfig, kubernetes *kubernetes.KubernetesClient, rpc rpc.MooClient, context context.Context, logger *log.Logger) *Agent {
	agent := &Agent{
		config:     config,
		kubernetes: kubernetes,
		mooClient:  rpc,
		context:    context,
		log:        logger,
	}

	return agent
}

func NewAgent(config *config.AgentConfig, kubernetes *kubernetes.KubernetesClient, rancher *rancher.RancherServer, context context.Context, logger *log.Logger) *Agent {
	agent := &Agent{
		config:     config,
		kubernetes: kubernetes,
		rancher:    rancher,
		context:    context,
		log:        logger,
	}

	return agent
}

func (a *Agent) registerCluster(agentId string) (bool, error) {
	rpcAgent := &rpc.Agent{
		ID:          agentId,
		Secret:      "", // TODO - implement
		IP:          "", // TODO - implement (may need external ip service)
		ClusterName: a.config.ClusterName,
		UseExisting: a.config.UseExisting,
	}
	resp, err := a.mooClient.RegisterAgent(a.context, rpcAgent)

	if err != nil {
		return false, err
	}

	return resp.Success, nil
}

func (a *Agent) ServerReconcile() {
	a.log.Debugf("starting server reconciliation")

	var agentId string

	agentId = a.config.ID
	if agentId == "" {
		uid, err := uuid.GenerateUUID()
		if err != nil {
			a.log.Fatalf("error generating uuid: %v", err)
		}
		agentId = uid
	}

	a.log.Infof("agent id is %s", agentId)

	for {
		rpcID := &rpc.AgentID{ID: agentId}
		status, err := a.mooClient.GetAgentStatus(a.context, rpcID)
		if err != nil {
			a.log.Errorf("error getting agent status from server: %v", err)
			time.Sleep(time.Second * 60) // TODO - figure out error backoff
			continue
		}

		if status.GetStatus() == rpc.Status_Unknown {
			// have not registered
			result, err := a.registerCluster(agentId)
			if err != nil {
				a.log.Errorf("error registering cluster with moo server : %v", err)
				time.Sleep(time.Second * 60) // TODO - figure out error backoff
				continue
			}
			if !result {
				a.log.Errorf("cluster registration unsuccessful, server returned false")
				time.Sleep(time.Second * 60) // TODO - figure out what to do if registration unsuccessful
				continue
			}
		}

		// if status is accepted, get the manifest url and proceed w/ reg
		if status.GetStatus() == rpc.Status_Accepted {
			manifestResponse, err := a.mooClient.GetManifestURL(a.context, rpcID)
			if err != nil {
				a.log.Errorf("error getting manifest from server: %v", err)
				continue // TODO - should we be continuing, or failing?
				// TODO - figure out error backoff?
			}

			manfURL := manifestResponse.GetURL()

			yaml, err := rancher.DoGet(manfURL, "", "", a.config.CACerts, a.config.Insecure)

			if err := a.kubernetes.ApplyManifest(yaml); err != nil {
				a.log.Errorf("error applying rancher import manifest: %v", err)
			}

			// should be happily registered, so exit.
			a.log.Infof("successfully registered cluster")
			os.Exit(0)
		}

		if status.GetStatus() == rpc.Status_Denied {
			a.log.Fatalf("server denied agent request, exiting")
		}

		var backoffTime int32

		switch status.GetStatus() {
		case rpc.Status_Held:
			backoffTime = status.GetHoldTime()
		case rpc.Status_Error:
			backoffTime = status.GetErrorTime()
		case rpc.Status_Pending:
			backoffTime = status.GetPendingTime()
		}

		if status.GetStatus() == rpc.Status_Error {
			a.log.Errorf("server responded with status of %s: %s", status.GetStatus(), status.GetMessage())
		} else {
			a.log.Infof("server responded with status of %s", status.GetStatus())
		}

		a.log.Infof("backing off for %d seconds", backoffTime)
		time.Sleep(time.Second * time.Duration(backoffTime))
	}
}

func (a *Agent) StandaloneReconcile() {
	a.log.Debugf("starting cluster registration")
	var regStatus RegistrationStatus

	regStatus = a.checkRegistration()

	if regStatus == Registered {
		a.log.Debugf("cluster already registered")
		return
	}

	a.log.Infof("registering cluster with rancher %s and name %s", a.config.URL, a.config.ClusterName)

	manifest, err := a.rancher.ReconcileToManifest(a.config.ClusterName, a.config.UseExisting)
	if err != nil {
		a.log.Error(err)
		return
	}

	if err := a.kubernetes.ApplyManifest(manifest); err != nil {
		a.log.Errorf("error applying rancher import manifest: %v", err)
		return
	}

	a.log.Info("cluster registered successfully")
}

func (a *Agent) checkRegistration() RegistrationStatus {
	a.log.Debugf("checking cluster registration status")
	// check if our cluster is registered or not.
	// registration is achieved when three things occur:
	// 1. creation of the cattle-system namespace
	// 2. creation of the cattle-cluster-agent deployment
	// 3. creation of the cattle-node-agent daemonset

	if ok, _ := a.kubernetes.CheckForNamespace(a.config.Namespace); !ok {
		// we do not have the namespace we need
		a.log.Debugf("namespace %s not found", a.config.Namespace)
		return Unregistered
	}

	if ok, _ := a.kubernetes.CheckForDeployment(a.config.Namespace, a.config.Deployment); !ok {
		a.log.Debugf("deployment %s in namespace %s not found", a.config.Deployment, a.config.Namespace)
		return Unregistered
	}

	if ok, _ := a.kubernetes.CheckForDaemonset(a.config.Namespace, a.config.Daemonset); !ok {
		a.log.Debugf("daemonset %s in namespace %s not found", a.config.Daemonset, a.config.Namespace)
		return Unregistered
	}

	return Registered // everything passed our checks
}
