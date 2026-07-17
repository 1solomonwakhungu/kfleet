// Package cluster provides helpers for reading state from Kubernetes clusters.
package cluster

import (
	"context"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

// NewClientset creates a Kubernetes client from a kubeconfig file or, when the
// path is empty, from the pod's in-cluster service account configuration.
func NewClientset(kubeconfig string) (*kubernetes.Clientset, error) {
	var (
		config *rest.Config
		err    error
	)
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// ServerVersion returns the Kubernetes API server's Git version.
func ServerVersion(ctx context.Context, cs *kubernetes.Clientset) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	version, err := cs.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return version.GitVersion, nil
}

// ListNodes returns a normalized view of all nodes in the cluster.
func ListNodes(ctx context.Context, cs *kubernetes.Clientset) ([]types.Node, error) {
	nodes, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]types.Node, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		ready := nodeReady(node.Status.Conditions)
		status := "NotReady"
		if ready {
			status = "Ready"
		}
		result = append(result, types.Node{
			Name:           node.Name,
			Status:         status,
			Roles:          nodeRoles(node.Labels),
			Version:        node.Status.NodeInfo.KubeletVersion,
			CPUCapacity:    node.Status.Capacity.Cpu().String(),
			MemoryCapacity: node.Status.Capacity.Memory().String(),
			Ready:          ready,
		})
	}
	return result, nil
}

// ListPods returns a normalized view of pods in namespace. An empty namespace
// lists pods across all namespaces.
func ListPods(ctx context.Context, cs *kubernetes.Clientset, namespace string) ([]types.Pod, error) {
	pods, err := cs.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]types.Pod, 0, len(pods.Items))
	for _, pod := range pods.Items {
		var restartCount int32
		for _, container := range pod.Status.ContainerStatuses {
			restartCount += container.RestartCount
		}

		item := types.Pod{
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			Phase:        string(pod.Status.Phase),
			NodeName:     pod.Spec.NodeName,
			RestartCount: restartCount,
			Ready:        podReady(pod.Status.Conditions),
		}
		if pod.Status.StartTime != nil {
			item.StartTime = pod.Status.StartTime.Time
		}
		result = append(result, item)
	}
	return result, nil
}

func nodeReady(conditions []corev1.NodeCondition) bool {
	for _, condition := range conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func podReady(conditions []corev1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func nodeRoles(labels map[string]string) []string {
	const rolePrefix = "node-role.kubernetes.io/"

	roles := make([]string, 0)
	for key, value := range labels {
		if strings.HasPrefix(key, rolePrefix) {
			role := strings.TrimPrefix(key, rolePrefix)
			if role != "" {
				roles = append(roles, role)
			}
		}
		if key == "kubernetes.io/role" && value != "" {
			roles = append(roles, value)
		}
	}
	sort.Strings(roles)
	return roles
}
