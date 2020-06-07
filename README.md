# Moo

Moo is a tool to auto-register Kubernetes (or k3s) clusters with a Rancher instance.

The driving force for this project was not wanting to interact with a k3s node after deployment,
but still have it register in Kubernetes.

# Usage

## Standalone

```text
NAME:
   moo - Auto-registration agent for Rancher imported clusters

USAGE:
   agent [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --kubeconfig value          kubeconfig if running outside of cluster (default: "/Users/ebauman/.config/k3d/moo/kubeconfig.yaml") [$KUBECONFIG]
   --namespace value           namespace used in registration check (default: "cattle-system") [$CATTLE_NAMESPACE]
   --deployment value          name of deployment used in registration check (default: "cattle-cluster-agent") [$CATTLE_DEPLOYMENT]
   --daemonset value           name of daemonset used in registration check (default: "cattle-node-agent") [$CATTLE_DAEMONSET]
   --rancher-url value         url of rancher instance [$RANCHER_URL]
   --rancher-access-key value  access key for rancher [$RANCHER_ACCESS_KEY]
   --rancher-secret-key value  secret key for rancher [$RANCHER_SECRET_KEY]
   --cluster-name value        name of this cluster when registering with rancher [$MOO_CLUSTER_NAME]
   --rancher-insecure          use an insecure connection to rancher (default: false) [$RANCHER_INSECURE]
   --rancher-cacerts value     path to cacerts file used when connecting to rancher [$RANCHER_CA_CERTS]
   --loglevel value            log level (trace, debug, info, warning, error, fatal, panic) (default: "info") [$LOGLEVEL]
   --use-existing-cluster      if cluster already exists in rancher, use it and import this node (default: false) [$MOO_USE_EXISTING]
   --help, -h                  show help (default: false)

```

## Kubernetes Cluster

Check out [kubernetes.yaml](package/kubernetes.yaml) for a manifest to deploy `moo-agent` into your cluster. 

If you're using k3s, this manifest can be placed in `/var/lib/rancher/k3s/server/manifests` which will auto-deploy
the `moo-agent` Job upon server installation. 


# Building

```text
cd agent/
go build
```

# Contributing

Please file issues for enhancements, upgrades, bugs, etc. 