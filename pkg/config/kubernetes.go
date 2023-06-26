package config

import (
	"context"
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

type k8sClientContextKeyType string

const k8sClientContextKey k8sClientContextKeyType = "kubernetes client"

func KubeClientFromContext(ctx context.Context) *kubernetes.Clientset {
	value := ctx.Value(k8sClientContextKey)
	if client, ok := value.(*kubernetes.Clientset); ok {
		return client
	}

	return nil
}

func ContextWithKubeClient(ctx context.Context, client *kubernetes.Clientset) context.Context {
	return context.WithValue(ctx, k8sClientContextKey, client)
}
