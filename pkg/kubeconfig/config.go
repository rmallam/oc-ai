package kubeconfig

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// Service account token path in OpenShift/Kubernetes pods
	serviceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	serviceAccountCAPath    = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	serviceAccountNamespace = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

// GetKubeConfig returns a Kubernetes config, preferring in-cluster configuration
// when running inside a pod, falling back to kubeconfig file
func GetKubeConfig(kubeconfigPath string) (*rest.Config, error) {
	// Check if we're running inside a Kubernetes pod by looking for service account token
	if isInCluster() {
		// Use in-cluster configuration (service account token)
		config, err := rest.InClusterConfig()
		if err == nil {
			return config, nil
		}
		// If in-cluster config fails but we detected we're in a pod, log but continue
		// This allows for development scenarios where files exist but config doesn't work
	}

	// Fall back to kubeconfig file
	if kubeconfigPath == "" {
		// Use default kubeconfig path
		if home, err := os.UserHomeDir(); err == nil {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	// Load kubeconfig from file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// IsInCluster detects if we're running inside a Kubernetes pod (exported version)
func IsInCluster() bool {
	return isInCluster()
}

// isInCluster detects if we're running inside a Kubernetes pod
func isInCluster() bool {
	// Check for service account files that are automatically mounted in pods
	_, tokenExists := os.Stat(serviceAccountTokenPath)
	_, caExists := os.Stat(serviceAccountCAPath)

	return tokenExists == nil && caExists == nil
}

// GetCurrentNamespace returns the current namespace from service account or kubeconfig
func GetCurrentNamespace(kubeconfigPath string) (string, error) {
	// If in cluster, read namespace from service account
	if isInCluster() {
		if data, err := os.ReadFile(serviceAccountNamespace); err == nil {
			return string(data), nil
		}
	}

	// Fall back to kubeconfig
	if kubeconfigPath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			kubeconfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	config, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return "default", nil // Return default if can't determine
	}

	context := config.CurrentContext
	if context == "" {
		return "default", nil
	}

	contextConfig, exists := config.Contexts[context]
	if !exists || contextConfig.Namespace == "" {
		return "default", nil
	}

	return contextConfig.Namespace, nil
}
