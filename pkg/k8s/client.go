package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/axatol/actions-job-dispatcher/pkg/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	client    *kubernetes.Clientset
	namespace string
}

var instance *Client

func GetClient() (*Client, error) {
	if instance != nil {
		return instance, nil
	}

	cfg, err := resolveKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %s", err)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %s", err)
	}

	instance = &Client{client, config.Namespace}
	return instance, nil
}

func resolveKubeConfig() (*rest.Config, error) {
	cfgCluster, errCluster := rest.InClusterConfig()
	if errCluster == nil {
		return cfgCluster, nil
	}

	precedence := []string{}
	if config.KubeConfig != "" {
		precedence = append(precedence, config.KubeConfig)
	}

	if home, _ := os.UserHomeDir(); home != "" {
		precedence = append(precedence, filepath.Join(home, ".kube", "config"))
	}

	cfgLocal, errLocal := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{Precedence: precedence},
		&clientcmd.ConfigOverrides{CurrentContext: config.KubeContext},
	).ClientConfig()
	if errLocal == nil {
		return cfgLocal, nil
	}

	return nil, fmt.Errorf("could not resolve local kubeconfig: %s, could not resolve cluster kubeconfig: %s", errLocal, errCluster)
}
