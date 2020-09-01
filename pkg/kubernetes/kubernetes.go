package kubernetes

import (
	"context"
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

type KubernetesClient struct {
	clientset *kubernetes.Clientset
	dynamic dynamic.Interface
	context context.Context

	log *log.Logger
}

func NewClient(path string, log *log.Logger, context context.Context) (*KubernetesClient, error) {
	var cfg *rest.Config
	var err error

	if path == "" {
		// we are in-cluster
		log.Info("building client from in-cluster config")
		cfg, err = rest.InClusterConfig()
	} else {
		log.Infof("building client from config file %s", path)
		cfg, err = clientcmd.BuildConfigFromFlags("", path)
	}

	if err != nil {
		log.Errorf("error building kubernetes config: %v", err)
		return nil, err
	}

	cfg.QPS = 100
	cfg.Burst = 100

	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Errorf("error building kubernetes config: %v", err)
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		log.Errorf("error building kubernetes config: %v", err)
		return nil, err
	}

	return &KubernetesClient{
		k8sClient,
		dynamicClient,
		context,
		log,
	}, nil
}

func (kc *KubernetesClient) CheckForNamespace(namespace string) (bool, error) {
	obj, err := kc.clientset.CoreV1().Namespaces().Get(kc.context, namespace, v1.GetOptions{})
	if errors.IsNotFound(err) {
		kc.log.Debugf("did not find namespace %s", namespace)
		return false, nil
	}

	if err != nil {
		kc.log.Debugf("error retrieving namespace: %v", err)
		return false, err
	}

	if obj.Name == namespace {
		kc.log.Debugf("successfully found namespace %s", namespace)
		return true, nil
	}

	return false, nil
}

func (kc *KubernetesClient) CheckForDeployment(namespace string, deployment string) (bool, error) {
	obj, err := kc.clientset.AppsV1().Deployments(namespace).Get(kc.context, deployment, v1.GetOptions{})

	if errors.IsNotFound(err) {
		kc.log.Debugf("did not find deployment %s in namespace %s", deployment, namespace)
		return false, nil
	}

	if err != nil {
		kc.log.Debugf("error retrieving deployment %s in namespace %s: %v", deployment, namespace, err)
		return false, err
	}

	if obj.Name == deployment {
		kc.log.Debugf("successfully found deployment %s in namespace %s", deployment, namespace)
		return true, nil
	}

	return false, nil
}

func (kc *KubernetesClient) CheckForDaemonset(namespace string, daemonset string) (bool, error) {
	obj, err := kc.clientset.AppsV1().DaemonSets(namespace).Get(kc.context, daemonset, v1.GetOptions{})
	if errors.IsNotFound(err) {
		kc.log.Debugf("did not find daemonset %s in namespace %s", daemonset, namespace)
		return false, nil
	}

	if err != nil {
		kc.log.Debugf("error retrieving daemonset %s in namespace %s: %v", daemonset, namespace, err)
		return false, err
	}

	if obj.Name == daemonset {
		kc.log.Debugf("successfully found daemonset %s in namespace %s", daemonset, namespace)
		return true, nil
	}

	return false, nil
}

func (kc *KubernetesClient) createObject(yaml []byte) error {
	obj := &unstructured.Unstructured{}

	dec := yaml2.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode(yaml, nil, obj)

	if err != nil {
		return err
	}

	mapping, err := findGVR(gvk, kc.clientset.DiscoveryClient)

	if err != nil {
		return err
	}

	// obtain REST interface for the gvr
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// for namespaced resources
		dr = kc.dynamic.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = kc.dynamic.Resource(mapping.Resource)
	}

	// create
	_, err = dr.Create(kc.context, obj, v1.CreateOptions{})

	if errors.IsAlreadyExists(err) {
		kc.log.Debugf("object %s already exists", obj.GetName())
		return nil
	}

	return err
}

// find the corresponding GVR (available in *meta.RESTMapping) for gvk
func findGVR(gvk *schema.GroupVersionKind, dc *discovery.DiscoveryClient) (*meta.RESTMapping, error) {
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	return mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
}

func (kc *KubernetesClient) ApplyManifest(manifest []byte) error {
	var err error
	// generate chunks from the manifest
	yamls := yamlSplit(manifest)

	limiter := time.Tick(APIRateLimitMilliseconds * time.Millisecond) // TODO - is this needed? see quota increases

	for _, y := range yamls {
		<-limiter
		err = kc.createObject(y)

		if err != nil {
			kc.log.Errorf("error creating object in kubernetes: %v", err)
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