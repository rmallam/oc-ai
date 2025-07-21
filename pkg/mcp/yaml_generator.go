package mcp

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// YAMLGenerator handles generation of YAML files for various Kubernetes resources
type YAMLGenerator struct{}

// NewYAMLGenerator creates a new YAMLGenerator instance
func NewYAMLGenerator() *YAMLGenerator {
	return &YAMLGenerator{}
}

// GenerateNamespaceYAML generates YAML for a namespace
func (y *YAMLGenerator) GenerateNamespaceYAML(name string) (string, error) {
	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
			},
		},
	}

	return y.marshalToYAML(namespace)
}

// GenerateConfigMapYAML generates YAML for a ConfigMap
func (y *YAMLGenerator) GenerateConfigMapYAML(name, namespace string, data map[string]string) (string, error) {
	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
			},
		},
		Data: data,
	}

	return y.marshalToYAML(configMap)
}

// GenerateSecretYAML generates YAML for a Secret
func (y *YAMLGenerator) GenerateSecretYAML(name, namespace string, data map[string][]byte, secretType corev1.SecretType) (string, error) {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
			},
		},
		Data: data,
		Type: secretType,
	}

	return y.marshalToYAML(secret)
}

// GenerateDeploymentYAML generates YAML for a Deployment
func (y *YAMLGenerator) GenerateDeploymentYAML(name, namespace, image string, replicas int32, env []corev1.EnvVar) (string, error) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        name,
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
							Env:   env,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}

	return y.marshalToYAML(deployment)
}

// GenerateServiceYAML generates YAML for a Service
func (y *YAMLGenerator) GenerateServiceYAML(name, namespace string, selector map[string]string, ports []corev1.ServicePort, serviceType corev1.ServiceType) (string, error) {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        name,
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports:    ports,
			Type:     serviceType,
		},
	}

	return y.marshalToYAML(service)
}

// GenerateScaleActionYAML generates YAML for a scale action (not a resource, but an action record)
func (y *YAMLGenerator) GenerateScaleActionYAML(deploymentName, namespace string, oldReplicas, newReplicas int32) (string, error) {
	scaleAction := map[string]interface{}{
		"apiVersion": "mcp.openshift.io/v1",
		"kind":       "ScaleAction",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("scale-%s-%s", deploymentName, time.Now().Format("20060102-150405")),
			"namespace": namespace,
			"labels": map[string]string{
				"action-type": "scale",
				"created-by":  "openshift-mcp",
				"created-at":  time.Now().Format("2006-01-02"),
			},
		},
		"spec": map[string]interface{}{
			"target": map[string]interface{}{
				"kind":      "Deployment",
				"name":      deploymentName,
				"namespace": namespace,
			},
			"scaleSpec": map[string]interface{}{
				"oldReplicas": oldReplicas,
				"newReplicas": newReplicas,
			},
		},
		"status": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"action":    "scale",
		},
	}

	return y.marshalToYAML(scaleAction)
}

// GenerateRestartActionYAML generates YAML for a restart action record
func (y *YAMLGenerator) GenerateRestartActionYAML(deploymentName, namespace string) (string, error) {
	restartAction := map[string]interface{}{
		"apiVersion": "mcp.openshift.io/v1",
		"kind":       "RestartAction",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("restart-%s-%s", deploymentName, time.Now().Format("20060102-150405")),
			"namespace": namespace,
			"labels": map[string]string{
				"action-type": "restart",
				"created-by":  "openshift-mcp",
				"created-at":  time.Now().Format("2006-01-02"),
			},
		},
		"spec": map[string]interface{}{
			"target": map[string]interface{}{
				"kind":      "Deployment",
				"name":      deploymentName,
				"namespace": namespace,
			},
		},
		"status": map[string]interface{}{
			"timestamp":         time.Now().Format(time.RFC3339),
			"action":            "restart",
			"restartedAt":       time.Now().Format(time.RFC3339),
			"restartAnnotation": "kubectl.kubernetes.io/restartedAt",
		},
	}

	return y.marshalToYAML(restartAction)
}

// GenerateDeleteActionYAML generates YAML for a delete action record
func (y *YAMLGenerator) GenerateDeleteActionYAML(resourceType, resourceName, namespace string) (string, error) {
	deleteAction := map[string]interface{}{
		"apiVersion": "mcp.openshift.io/v1",
		"kind":       "DeleteAction",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("delete-%s-%s-%s", strings.ToLower(resourceType), resourceName, time.Now().Format("20060102-150405")),
			"namespace": namespace,
			"labels": map[string]string{
				"action-type": "delete",
				"created-by":  "openshift-mcp",
				"created-at":  time.Now().Format("2006-01-02"),
			},
		},
		"spec": map[string]interface{}{
			"target": map[string]interface{}{
				"kind":      resourceType,
				"name":      resourceName,
				"namespace": namespace,
			},
		},
		"status": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"action":    "delete",
		},
	}

	return y.marshalToYAML(deleteAction)
}

// GenerateGenericResourceYAML generates YAML from a generic resource string
func (y *YAMLGenerator) GenerateGenericResourceYAML(yamlContent string) (string, error) {
	// Parse the YAML to ensure it's valid
	var resource interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &resource); err != nil {
		return "", fmt.Errorf("invalid YAML content: %v", err)
	}

	// Add MCP metadata if it's a map
	if resourceMap, ok := resource.(map[interface{}]interface{}); ok {
		if metadata, ok := resourceMap["metadata"].(map[interface{}]interface{}); ok {
			if labels, ok := metadata["labels"].(map[interface{}]interface{}); ok {
				labels["created-by"] = "openshift-mcp"
				labels["created-at"] = time.Now().Format("2006-01-02")
			} else {
				metadata["labels"] = map[interface{}]interface{}{
					"created-by": "openshift-mcp",
					"created-at": time.Now().Format("2006-01-02"),
				}
			}
		}
	}

	return y.marshalToYAML(resource)
}

// marshalToYAML marshals an object to YAML with proper formatting
func (y *YAMLGenerator) marshalToYAML(obj interface{}) (string, error) {
	yamlData, err := yaml.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %v", err)
	}

	// Add header comment
	header := fmt.Sprintf("# Generated by OpenShift MCP on %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return header + string(yamlData), nil
}

// ParseYAMLContent parses YAML content and returns structured data
func (y *YAMLGenerator) ParseYAMLContent(yamlContent string) (map[string]interface{}, error) {
	var data map[string]interface{}

	if err := yaml.Unmarshal([]byte(yamlContent), &data); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	return data, nil
}

// GenerateDefaultServicePortsForDeployment generates default service ports for a deployment
func (y *YAMLGenerator) GenerateDefaultServicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:       "http",
			Port:       80,
			TargetPort: intstr.FromInt(8080),
			Protocol:   corev1.ProtocolTCP,
		},
	}
}

// GenerateDefaultEnvVars generates default environment variables
func (y *YAMLGenerator) GenerateDefaultEnvVars() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "CREATED_BY",
			Value: "openshift-mcp",
		},
		{
			Name:  "CREATED_AT",
			Value: time.Now().Format("2006-01-02 15:04:05"),
		},
	}
}

// ArgoCD-specific types and functions

// ArgoCDApplication represents an ArgoCD Application
type ArgoCDApplication struct {
	APIVersion string                   `yaml:"apiVersion"`
	Kind       string                   `yaml:"kind"`
	Metadata   metav1.ObjectMeta        `yaml:"metadata"`
	Spec       ArgoCDApplicationSpec    `yaml:"spec"`
	Status     *ArgoCDApplicationStatus `yaml:"status,omitempty"`
}

// ArgoCDApplicationSpec represents the spec of an ArgoCD Application
type ArgoCDApplicationSpec struct {
	Project     string                  `yaml:"project"`
	Source      ArgoCDApplicationSource `yaml:"source"`
	Destination ArgoCDApplicationDest   `yaml:"destination"`
	SyncPolicy  *ArgoCDSyncPolicy       `yaml:"syncPolicy,omitempty"`
}

// ArgoCDApplicationSource represents the source of an ArgoCD Application
type ArgoCDApplicationSource struct {
	RepoURL        string `yaml:"repoURL"`
	Path           string `yaml:"path"`
	TargetRevision string `yaml:"targetRevision"`
}

// ArgoCDApplicationDest represents the destination of an ArgoCD Application
type ArgoCDApplicationDest struct {
	Server    string `yaml:"server"`
	Namespace string `yaml:"namespace"`
}

// ArgoCDSyncPolicy represents the sync policy of an ArgoCD Application
type ArgoCDSyncPolicy struct {
	Automated   *ArgoCDSyncPolicyAutomated `yaml:"automated,omitempty"`
	SyncOptions []string                   `yaml:"syncOptions,omitempty"`
}

// ArgoCDSyncPolicyAutomated represents automated sync policy
type ArgoCDSyncPolicyAutomated struct {
	Prune    bool `yaml:"prune"`
	SelfHeal bool `yaml:"selfHeal"`
}

// ArgoCDApplicationStatus represents the status of an ArgoCD Application
type ArgoCDApplicationStatus struct {
	Health ArgoCDHealthStatus `yaml:"health"`
	Sync   ArgoCDSyncStatus   `yaml:"sync"`
}

// ArgoCDHealthStatus represents the health status
type ArgoCDHealthStatus struct {
	Status string `yaml:"status"`
}

// ArgoCDSyncStatus represents the sync status
type ArgoCDSyncStatus struct {
	Status   string `yaml:"status"`
	Revision string `yaml:"revision"`
}

// GenerateArgoCDApplicationYAML generates ArgoCD Application YAML
func (y *YAMLGenerator) GenerateArgoCDApplicationYAML(name, namespace, repoURL, path, targetRevision, destinationServer, destinationNamespace string, automated bool) (string, error) {
	var syncPolicy *ArgoCDSyncPolicy
	if automated {
		syncPolicy = &ArgoCDSyncPolicy{
			Automated: &ArgoCDSyncPolicyAutomated{
				Prune:    true,
				SelfHeal: true,
			},
			SyncOptions: []string{
				"CreateNamespace=true",
				"PrunePropagationPolicy=foreground",
				"PruneLast=true",
			},
		}
	}

	application := &ArgoCDApplication{
		APIVersion: "argoproj.io/v1alpha1",
		Kind:       "Application",
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        name,
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
			},
			Annotations: map[string]string{
				"argocd.argoproj.io/sync-wave": "0",
			},
		},
		Spec: ArgoCDApplicationSpec{
			Project: "default",
			Source: ArgoCDApplicationSource{
				RepoURL:        repoURL,
				Path:           path,
				TargetRevision: targetRevision,
			},
			Destination: ArgoCDApplicationDest{
				Server:    destinationServer,
				Namespace: destinationNamespace,
			},
			SyncPolicy: syncPolicy,
		},
	}

	return y.marshalToYAML(application)
}

// GenerateArgoCDAppOfAppsYAML generates ArgoCD App of Apps pattern YAML
func (y *YAMLGenerator) GenerateArgoCDAppOfAppsYAML(name, namespace, repoURL, path string, applications []string) (string, error) {
	appOfApps := &ArgoCDApplication{
		APIVersion: "argoproj.io/v1alpha1",
		Kind:       "Application",
		Metadata: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        name,
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
				"app-type":   "app-of-apps",
			},
			Annotations: map[string]string{
				"argocd.argoproj.io/sync-wave": "-1",
			},
		},
		Spec: ArgoCDApplicationSpec{
			Project: "default",
			Source: ArgoCDApplicationSource{
				RepoURL:        repoURL,
				Path:           path,
				TargetRevision: "HEAD",
			},
			Destination: ArgoCDApplicationDest{
				Server:    "https://kubernetes.default.svc",
				Namespace: namespace,
			},
			SyncPolicy: &ArgoCDSyncPolicy{
				Automated: &ArgoCDSyncPolicyAutomated{
					Prune:    true,
					SelfHeal: true,
				},
				SyncOptions: []string{
					"CreateNamespace=true",
				},
			},
		},
	}

	return y.marshalToYAML(appOfApps)
}

// GenerateKustomizationYAML generates a Kustomization file for ArgoCD
func (y *YAMLGenerator) GenerateKustomizationYAML(resources []string, namespace string) (string, error) {
	kustomization := map[string]interface{}{
		"apiVersion": "kustomize.config.k8s.io/v1beta1",
		"kind":       "Kustomization",
		"metadata": map[string]interface{}{
			"name": fmt.Sprintf("kustomization-%s", time.Now().Format("20060102-150405")),
			"labels": map[string]string{
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
			},
		},
		"namespace": namespace,
		"resources": resources,
		"commonLabels": map[string]string{
			"managed-by": "openshift-mcp",
			"created-at": time.Now().Format("2006-01-02"),
		},
	}

	return y.marshalToYAML(kustomization)
}

// GenerateArgocdCompatibleNamespaceYAML generates ArgoCD-compatible namespace YAML
func (y *YAMLGenerator) GenerateArgocdCompatibleNamespaceYAML(name string) (string, error) {
	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"name":       name,
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
				"managed-by": "argocd",
			},
			Annotations: map[string]string{
				"argocd.argoproj.io/sync-wave": "-10",
				"argocd.argoproj.io/hook":      "PreSync",
			},
		},
	}

	return y.marshalToYAML(namespace)
}

// GenerateArgocdCompatibleDeploymentYAML generates ArgoCD-compatible deployment YAML
func (y *YAMLGenerator) GenerateArgocdCompatibleDeploymentYAML(name, namespace, image string, replicas int32, env []corev1.EnvVar, syncWave string) (string, error) {
	if syncWave == "" {
		syncWave = "1"
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        name,
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
				"managed-by": "argocd",
			},
			Annotations: map[string]string{
				"argocd.argoproj.io/sync-wave":      syncWave,
				"deployment.kubernetes.io/revision": "1",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":        name,
						"version":    "v1",
						"managed-by": "argocd",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
							Env:   env,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       5,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       10,
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	return y.marshalToYAML(deployment)
}

// GenerateArgocdCompatibleServiceYAML generates ArgoCD-compatible service YAML
func (y *YAMLGenerator) GenerateArgocdCompatibleServiceYAML(name, namespace string, selector map[string]string, ports []corev1.ServicePort, serviceType corev1.ServiceType, syncWave string) (string, error) {
	if syncWave == "" {
		syncWave = "2"
	}

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        name,
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
				"managed-by": "argocd",
			},
			Annotations: map[string]string{
				"argocd.argoproj.io/sync-wave": syncWave,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports:    ports,
			Type:     serviceType,
		},
	}

	return y.marshalToYAML(service)
}

// GenerateArgocdCompatibleConfigMapYAML generates ArgoCD-compatible configmap YAML
func (y *YAMLGenerator) GenerateArgocdCompatibleConfigMapYAML(name, namespace string, data map[string]string, syncWave string) (string, error) {
	if syncWave == "" {
		syncWave = "0"
	}

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        name,
				"created-by": "openshift-mcp",
				"created-at": time.Now().Format("2006-01-02"),
				"managed-by": "argocd",
			},
			Annotations: map[string]string{
				"argocd.argoproj.io/sync-wave": syncWave,
			},
		},
		Data: data,
	}

	return y.marshalToYAML(configMap)
}

// GenerateArgocdManifestBundle generates a complete ArgoCD manifest bundle
func (y *YAMLGenerator) GenerateArgocdManifestBundle(appName, namespace, image string, replicas int32, configData map[string]string, env []corev1.EnvVar) (map[string]string, error) {
	manifests := make(map[string]string)

	// Generate namespace
	namespaceYAML, err := y.GenerateArgocdCompatibleNamespaceYAML(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate namespace: %v", err)
	}
	manifests["namespace.yaml"] = namespaceYAML

	// Generate configmap if data provided
	if len(configData) > 0 {
		configMapYAML, err := y.GenerateArgocdCompatibleConfigMapYAML(appName+"-config", namespace, configData, "0")
		if err != nil {
			return nil, fmt.Errorf("failed to generate configmap: %v", err)
		}
		manifests["configmap.yaml"] = configMapYAML
	}

	// Generate deployment
	deploymentYAML, err := y.GenerateArgocdCompatibleDeploymentYAML(appName, namespace, image, replicas, env, "1")
	if err != nil {
		return nil, fmt.Errorf("failed to generate deployment: %v", err)
	}
	manifests["deployment.yaml"] = deploymentYAML

	// Generate service
	serviceYAML, err := y.GenerateArgocdCompatibleServiceYAML(
		appName,
		namespace,
		map[string]string{"app": appName},
		y.GenerateDefaultServicePorts(),
		corev1.ServiceTypeClusterIP,
		"2",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate service: %v", err)
	}
	manifests["service.yaml"] = serviceYAML

	// Generate kustomization
	resources := []string{"namespace.yaml", "deployment.yaml", "service.yaml"}
	if len(configData) > 0 {
		resources = append(resources, "configmap.yaml")
	}
	kustomizationYAML, err := y.GenerateKustomizationYAML(resources, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate kustomization: %v", err)
	}
	manifests["kustomization.yaml"] = kustomizationYAML

	return manifests, nil
}
