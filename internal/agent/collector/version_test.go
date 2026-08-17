package collector

import (
	"context"
	"testing"

	"github.com/1solomonwakhungu/kfleet/internal/version"
	"k8s.io/client-go/kubernetes/fake"
)

// TestCollectReportsStampedVersion proves collected snapshots carry the build
// version rather than a hard-coded constant.
func TestCollectReportsStampedVersion(t *testing.T) {
	original := version.Version
	t.Cleanup(func() { version.Version = original })
	version.Version = "v9.9.9"

	c := &Collector{clientset: fake.NewSimpleClientset()}
	state, err := c.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if state.AgentVersion != "v9.9.9" {
		t.Fatalf("AgentVersion = %q, want the stamped build version", state.AgentVersion)
	}
}
