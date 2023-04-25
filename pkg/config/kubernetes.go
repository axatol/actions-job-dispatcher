package config

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func resolveKubeConfig() (*rest.Config, error) {
	cfgCluster, errCluster := rest.InClusterConfig()
	if errCluster == nil {
		return cfgCluster, nil
	}

	precedence := []string{}
	if kubeConfig.Value() != "" {
		precedence = append(precedence, kubeConfig.Value())
	}

	if home, _ := os.UserHomeDir(); home != "" {
		precedence = append(precedence, filepath.Join(home, ".kube", "config"))
	}

	cfgLocal, errLocal := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{Precedence: precedence},
		&clientcmd.ConfigOverrides{CurrentContext: kubeContext.Value()},
	).ClientConfig()
	if errLocal == nil {
		return cfgLocal, nil
	}

	return nil, fmt.Errorf("could not resolve local kubeconfig: %s, could not resolve cluster kubeconfig: %s", errLocal, errCluster)
}

func KubeClient() (*kubernetes.Clientset, error) {
	cfg, err := resolveKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %s", err)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %s", err)
	}

	return client, nil
}
