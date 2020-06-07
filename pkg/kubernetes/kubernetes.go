package kubernetes

import (
	"github.com/ebauman/moo/pkg/config"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	yaml2 "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
	"time"
)

const (
	APIRateLimitMilliseconds = 3000 // how many ms to wait between k8s api calls, helps avoid server-side throttling
)

func BuildClients(path string) (*kubernetes.Clientset, dynamic.Interface, error) {
	var cfg *rest.Config
	var err error

	if path == "" {
		// we are in-cluster
		log.Info("building client from in-cluster cfg")
		cfg, err = rest.InClusterConfig()
	} else {
		log.Info("building client from cfg file")
		cfg, err = clientcmd.BuildConfigFromFlags("", path)
	}

	if err != nil {
		log.Errorf("error building kubernetes cfg: %v", err)
		return nil, nil, err
	}

	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Errorf("error building kubernetes client: %v", err)
		return nil, nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		log.Errorf("error building dynamic client: %v", err)
		return nil, nil, err
	}

	return k8sClient, dynamicClient, nil
}

func CheckForNamespace(config *config.AgentConfig) (bool, error) {
	obj, err := config.Kubernetes.CoreV1().Namespaces().Get(config.Context, config.Namespace, v1.GetOptions{})
	if errors.IsNotFound(err) {
		config.Log.Debugf("did not find namespace %s", config.Namespace)
		return false, nil
	}

	if err != nil {
		config.Log.Debugf("error retrieving namespace: %v", err)
		return false, err
	}

	if obj.Name == config.Namespace {
		config.Log.Debugf("successfully found namespace %s", config.Namespace)
		return true, nil
	}

	return false, nil
}

func CheckForDeployment(config *config.AgentConfig) (bool, error) {
	obj, err := config.Kubernetes.AppsV1().Deployments(config.Namespace).Get(config.Context, config.Deployment, v1.GetOptions{})

	if errors.IsNotFound(err) {
		config.Log.Debugf("did not find deployment %s in namespace %s", config.Deployment, config.Namespace)
		return false, nil
	}

	if err != nil {
		config.Log.Debugf("error retrieving deployment %s in namespace %s: %v", config.Deployment, config.Namespace, err)
		return false, err
	}

	if obj.Name == config.Deployment {
		config.Log.Debugf("successfully found deployment %s in namespace %s", config.Deployment, config.Namespace)
		return true, nil
	}

	return false, nil
}

func CheckForDaemonset(config *config.AgentConfig) (bool, error) {
	obj, err := config.Kubernetes.AppsV1().DaemonSets(config.Namespace).Get(config.Context, config.Daemonset, v1.GetOptions{})
	if errors.IsNotFound(err) {
		config.Log.Debugf("did not find daemonset %s in namespace %s", config.Daemonset, config.Namespace)
		return false, nil
	}

	if err != nil {
		config.Log.Debugf("error retrieving daemonset %s in namespace %s: %v", config.Daemonset, config.Namespace, err)
		return false, err
	}

	if obj.Name == config.Daemonset {
		config.Log.Debugf("successfully found daemonset %s in namespace %s", config.Daemonset, config.Namespace)
		return true, nil
	}

	return false, nil
}

func createObject(config *config.AgentConfig, yaml []byte) error {
	obj := &unstructured.Unstructured{}

	dec := yaml2.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode(yaml, nil, obj)

	if err != nil {
		return err
	}

	mapping, err := findGVR(gvk, config.Kubernetes.DiscoveryClient)

	if err != nil {
		return err
	}

	// obtain REST interface for the gvr
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// for namespaced resources
		dr = config.Dynamic.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = config.Dynamic.Resource(mapping.Resource)
	}

	// create
	_, err = dr.Create(config.Context, obj, v1.CreateOptions{})

	if errors.IsAlreadyExists(err) {
		config.Log.Debugf("object %s already exists", obj.GetName())
		return nil
	}

	return err
}

// find the corresponding GVR (available in *meta.RESTMapping) for gvk
func findGVR(gvk *schema.GroupVersionKind, dc *discovery.DiscoveryClient) (*meta.RESTMapping, error) {
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	return mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
}

func ApplyManifest(config *config.AgentConfig, manifest []byte) error {
	var err error
	// generate chunks from the manifest
	yamls := yamlSplit(manifest)

	limiter := time.Tick(APIRateLimitMilliseconds * time.Millisecond)

	for _, y := range yamls {
		<-limiter
		err = createObject(config, y)

		if err != nil {
			config.Log.Errorf("error creating object in kubernetes: %v", err)
		}
	}

	return err
}


func yamlSplit(manifest []byte) [][]byte {
	// need to split up the yaml into []byte chunks on ---
	// this is so we can parse out the individual objects before they are applied

	rawYaml := string(manifest)

	yamls := strings.Split(rawYaml, "---")

	output := make([][]byte, 0)

	for _, y := range yamls {
		if !strings.Contains(y, "apiVersion:") || !strings.Contains(y, "kind:"){
			continue // would be an invalid yaml, prune out some garbage
		}
		output = append(output, []byte(y))
	}

	return output
}