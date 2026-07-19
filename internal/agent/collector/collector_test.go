package collector

import (
	"context"
	"fmt"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCollect(t *testing.T) {
	replicas := int32(3)
	baseTime := time.Now().UTC().Add(-time.Hour)
	objects := []runtime.Object{
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node-1", Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""}},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
				NodeInfo:   corev1.NodeSystemInfo{KubeletVersion: "v1.32.0"},
				Capacity: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("4"),
					corev1.ResourceMemory: resource.MustParse("16Gi"),
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"},
			Spec:       corev1.PodSpec{NodeName: "node-1"},
			Status: corev1.PodStatus{
				Phase:             corev1.PodRunning,
				StartTime:         &metav1.Time{Time: baseTime},
				ContainerStatuses: []corev1.ContainerStatus{{RestartCount: 2, Ready: true}},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default", CreationTimestamp: metav1.NewTime(baseTime)},
			Spec: corev1.ServiceSpec{
				Type:      corev1.ServiceTypeClusterIP,
				ClusterIP: "10.0.0.1",
				Ports:     []corev1.ServicePort{{Name: "http", Port: 80, TargetPort: intstr.FromInt32(8080)}},
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default", CreationTimestamp: metav1.NewTime(baseTime)},
			Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
			Status:     appsv1.DeploymentStatus{ReadyReplicas: 2, UpdatedReplicas: 2, AvailableReplicas: 2},
		},
	}
	for index := range 105 {
		objects = append(objects, &corev1.Event{
			ObjectMeta:    metav1.ObjectMeta{Name: fmt.Sprintf("event-%03d", index), Namespace: "default"},
			Reason:        "Scheduled",
			LastTimestamp: metav1.NewTime(baseTime.Add(time.Duration(index) * time.Minute)),
		})
	}

	c := &Collector{clientset: fake.NewSimpleClientset(objects...)}
	state, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if state.NodeCount != 1 || state.PodCount != 1 || state.K8sVersion != "v1.32.0" {
		t.Fatalf("Collect() counts/version = (%d, %d, %q)", state.NodeCount, state.PodCount, state.K8sVersion)
	}
	if state.Nodes[0].Status != "Ready" || len(state.Nodes[0].Roles) != 1 || state.Nodes[0].Roles[0] != "control-plane" {
		t.Fatalf("node info = %#v", state.Nodes[0])
	}
	if state.Pods[0].Restarts != 2 || state.Pods[0].Node != "node-1" || !state.Pods[0].Ready || !state.Pods[0].StartTime.Equal(baseTime) {
		t.Fatalf("pod info = %#v", state.Pods[0])
	}
	if len(state.Services) != 1 || len(state.Deployments) != 1 {
		t.Fatalf("resource counts = services %d, deployments %d", len(state.Services), len(state.Deployments))
	}
	if state.Services[0].Age == "" || state.Deployments[0].Age == "" || state.Deployments[0].UpdatedReplicas != 2 {
		t.Fatalf("resource metadata = service %#v, deployment %#v", state.Services[0], state.Deployments[0])
	}
	if len(state.Events) != maxEvents || state.Events[0].Name != "event-104" {
		t.Fatalf("events = %d, first = %q; want latest 100", len(state.Events), state.Events[0].Name)
	}
	if state.CollectedAt.IsZero() || state.AgentVersion == "" {
		t.Fatalf("collection metadata = (%v, %q)", state.CollectedAt, state.AgentVersion)
	}
}
