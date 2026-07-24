// Package collector gathers Kubernetes cluster state for the kfleet agent.
package collector

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/agent/config"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	agentVersion = "0.1.0"
	maxEvents    = 100
)

// Collector gathers Kubernetes state using a client-go clientset.
type Collector struct {
	clientset kubernetes.Interface
}

// New builds a collector using an explicit kubeconfig or in-cluster credentials.
func New(cfg *config.Config) (*Collector, error) {
	var (
		restConfig *rest.Config
		err        error
	)
	if cfg.Kubeconfig != "" {
		restConfig, err = clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	} else {
		restConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("load Kubernetes client configuration: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes client: %w", err)
	}
	return &Collector{clientset: clientset}, nil
}

// Collect lists cluster-wide resources and returns a point-in-time snapshot.
func (c *Collector) Collect(ctx context.Context) (*ClusterState, error) {
	collectedAt := time.Now().UTC()
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	pods, err := c.clientset.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}
	services, err := c.clientset.CoreV1().Services(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	deployments, err := c.clientset.AppsV1().Deployments(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list deployments: %w", err)
	}
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}
	events, err := c.clientset.CoreV1().Events(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}

	state := &ClusterState{
		Nodes:        make([]NodeInfo, 0, len(nodes.Items)),
		Pods:         make([]PodInfo, 0, len(pods.Items)),
		Services:     make([]ServiceInfo, 0, len(services.Items)),
		Deployments:  make([]DeploymentInfo, 0, len(deployments.Items)),
		Namespaces:   make([]NamespaceInfo, 0, len(namespaces.Items)),
		Events:       make([]EventInfo, 0, min(len(events.Items), maxEvents)),
		NodeCount:    len(nodes.Items),
		PodCount:     len(pods.Items),
		CollectedAt:  collectedAt,
		AgentVersion: agentVersion,
	}
	for _, node := range nodes.Items {
		info := nodeInfo(node)
		state.Nodes = append(state.Nodes, info)
		if state.K8sVersion == "" && info.K8sVersion != "" {
			state.K8sVersion = info.K8sVersion
		}
	}
	for _, pod := range pods.Items {
		state.Pods = append(state.Pods, podInfo(pod))
	}
	for _, service := range services.Items {
		state.Services = append(state.Services, serviceInfo(service, collectedAt))
	}
	for _, deployment := range deployments.Items {
		state.Deployments = append(state.Deployments, deploymentInfo(deployment, collectedAt))
	}
	for _, namespace := range namespaces.Items {
		state.Namespaces = append(state.Namespaces, NamespaceInfo{
			Name:   namespace.Name,
			Labels: cloneMap(namespace.Labels),
		})
	}
	sort.Slice(events.Items, func(i, j int) bool {
		return eventTimestamp(events.Items[i]).After(eventTimestamp(events.Items[j]))
	})
	for index, event := range events.Items {
		if index == maxEvents {
			break
		}
		state.Events = append(state.Events, eventInfo(event))
	}
	return state, nil
}

func nodeInfo(node corev1.Node) NodeInfo {
	status := "Unknown"
	for _, condition := range node.Status.Conditions {
		if condition.Type != corev1.NodeReady {
			continue
		}
		switch condition.Status {
		case corev1.ConditionTrue:
			status = "Ready"
		case corev1.ConditionFalse:
			status = "NotReady"
		}
		break
	}
	roles := make([]string, 0)
	for key, value := range node.Labels {
		if strings.HasPrefix(key, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(key, "node-role.kubernetes.io/")
			if role != "" {
				roles = append(roles, role)
			}
		} else if key == "kubernetes.io/role" && value != "" {
			roles = append(roles, value)
		}
	}
	sort.Strings(roles)
	return NodeInfo{
		Name:           node.Name,
		Status:         status,
		K8sVersion:     node.Status.NodeInfo.KubeletVersion,
		Roles:          roles,
		CPUCapacity:    node.Status.Capacity.Cpu().String(),
		MemoryCapacity: node.Status.Capacity.Memory().String(),
	}
}

func podInfo(pod corev1.Pod) PodInfo {
	var restarts int32
	ready := len(pod.Status.ContainerStatuses) > 0
	for _, status := range pod.Status.ContainerStatuses {
		restarts += status.RestartCount
		ready = ready && status.Ready
	}
	for _, status := range pod.Status.InitContainerStatuses {
		restarts += status.RestartCount
	}
	info := PodInfo{
		Namespace:   pod.Namespace,
		Name:        pod.Name,
		Phase:       string(pod.Status.Phase),
		Restarts:    restarts,
		Node:        pod.Spec.NodeName,
		Ready:       ready,
		StartTime:   timestampValue(pod.Status.StartTime),
		HostNetwork: pod.Spec.HostNetwork,
		HostPID:     pod.Spec.HostPID,
		HostIPC:     pod.Spec.HostIPC,
	}
	containers := make([]corev1.Container, 0, len(pod.Spec.InitContainers)+len(pod.Spec.Containers)+len(pod.Spec.EphemeralContainers))
	containers = append(containers, pod.Spec.InitContainers...)
	containers = append(containers, pod.Spec.Containers...)
	for _, container := range pod.Spec.EphemeralContainers {
		containers = append(containers, corev1.Container{
			Name:            container.Name,
			SecurityContext: container.SecurityContext,
		})
	}
	if len(containers) == 0 {
		return info
	}

	info.SecurityContextKnown = true
	info.RunAsNonRoot = true
	info.ReadOnlyRootFilesystem = true
	info.CapabilitiesDroppedAll = true
	for _, container := range containers {
		security := container.SecurityContext
		if security != nil && security.Privileged != nil && *security.Privileged {
			info.Privileged = true
		}
		if !effectiveRunAsNonRoot(pod.Spec.SecurityContext, security) {
			info.RunAsNonRoot = false
		}
		if security == nil || security.ReadOnlyRootFilesystem == nil || !*security.ReadOnlyRootFilesystem {
			info.ReadOnlyRootFilesystem = false
		}
		if security == nil || security.AllowPrivilegeEscalation == nil || *security.AllowPrivilegeEscalation {
			info.AllowsPrivilegeEscalation = true
		}
		if !dropsAllCapabilities(security) {
			info.CapabilitiesDroppedAll = false
		}
	}
	return info
}

func serviceInfo(service corev1.Service, collectedAt time.Time) ServiceInfo {
	ports := make([]string, 0, len(service.Spec.Ports))
	for _, port := range service.Spec.Ports {
		value := strconv.Itoa(int(port.Port)) + "/" + string(port.Protocol)
		if port.Name != "" {
			value = port.Name + ":" + value
		}
		ports = append(ports, value)
	}
	return ServiceInfo{
		Namespace:   service.Namespace,
		Name:        service.Name,
		Type:        string(service.Spec.Type),
		ClusterIP:   service.Spec.ClusterIP,
		ExternalIPs: append([]string(nil), service.Spec.ExternalIPs...),
		Ports:       ports,
		Age:         formatAge(service.CreationTimestamp.Time, collectedAt),
	}
}

func deploymentInfo(deployment appsv1.Deployment, collectedAt time.Time) DeploymentInfo {
	var desired int32
	if deployment.Spec.Replicas != nil {
		desired = *deployment.Spec.Replicas
	}
	images := make([]string, 0, len(deployment.Spec.Template.Spec.InitContainers)+len(deployment.Spec.Template.Spec.Containers))
	for _, container := range deployment.Spec.Template.Spec.InitContainers {
		images = append(images, container.Image)
	}
	for _, container := range deployment.Spec.Template.Spec.Containers {
		images = append(images, container.Image)
	}
	sort.Strings(images)
	config, _ := json.Marshal(deployment.Spec)
	hash := sha256.Sum256(config)
	return DeploymentInfo{
		Namespace:         deployment.Namespace,
		Name:              deployment.Name,
		DesiredReplicas:   desired,
		ReadyReplicas:     deployment.Status.ReadyReplicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		UpdatedReplicas:   deployment.Status.UpdatedReplicas,
		Age:               formatAge(deployment.CreationTimestamp.Time, collectedAt),
		ConfigHash:        fmt.Sprintf("%x", hash[:12]),
		Images:            images,
	}
}

func effectiveRunAsNonRoot(pod *corev1.PodSecurityContext, container *corev1.SecurityContext) bool {
	if container != nil && container.RunAsNonRoot != nil {
		return *container.RunAsNonRoot
	}
	return pod != nil && pod.RunAsNonRoot != nil && *pod.RunAsNonRoot
}

func dropsAllCapabilities(security *corev1.SecurityContext) bool {
	if security == nil || security.Capabilities == nil {
		return false
	}
	for _, capability := range security.Capabilities.Drop {
		if strings.EqualFold(string(capability), "ALL") {
			return true
		}
	}
	return false
}

func cloneMap(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func timestampValue(value *metav1.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.Time
}

func formatAge(createdAt, collectedAt time.Time) string {
	if createdAt.IsZero() {
		return ""
	}
	age := collectedAt.Sub(createdAt)
	if age < 0 {
		age = 0
	}
	switch {
	case age < time.Minute:
		return strconv.Itoa(int(age/time.Second)) + "s"
	case age < time.Hour:
		return strconv.Itoa(int(age/time.Minute)) + "m"
	case age < 24*time.Hour:
		return strconv.Itoa(int(age/time.Hour)) + "h"
	default:
		return strconv.Itoa(int(age/(24*time.Hour))) + "d"
	}
}

func eventInfo(event corev1.Event) EventInfo {
	return EventInfo{
		Namespace:      event.Namespace,
		Name:           event.Name,
		Type:           event.Type,
		Reason:         event.Reason,
		Message:        event.Message,
		InvolvedObject: event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name,
		Count:          event.Count,
		LastTimestamp:  eventTimestamp(event),
	}
}

func eventTimestamp(event corev1.Event) time.Time {
	if !event.EventTime.IsZero() {
		return event.EventTime.Time
	}
	if !event.LastTimestamp.IsZero() {
		return event.LastTimestamp.Time
	}
	return event.CreationTimestamp.Time
}
