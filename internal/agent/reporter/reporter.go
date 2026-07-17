// Package reporter sends collected Kubernetes state to the kfleet hub.
package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/agent/collector"
	"github.com/1solomonwakhungu/kfleet/internal/agent/config"
)

const requestTimeout = 10 * time.Second

// Reporter posts cluster state snapshots to the hub.
type Reporter struct {
	hubURL      string
	token       string
	clusterName string
	client      *http.Client
}

// New constructs a reporter from agent configuration.
func New(cfg *config.Config) *Reporter {
	return &Reporter{
		hubURL:      strings.TrimRight(cfg.HubURL, "/"),
		token:       cfg.HubToken,
		clusterName: cfg.ClusterName,
		client:      &http.Client{Timeout: requestTimeout},
	}
}

// Report posts state to the cluster status endpoint.
func (r *Reporter) Report(ctx context.Context, state *collector.ClusterState) error {
	body, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode cluster state: %w", err)
	}
	endpoint := r.hubURL + "/api/v1/clusters/" + url.PathEscape(r.clusterName) + "/status"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create status request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.token)

	response, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("report cluster state: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, response.Body)
		return fmt.Errorf("hub returned status %s", response.Status)
	}
	_, _ = io.Copy(io.Discard, response.Body)
	return nil
}
