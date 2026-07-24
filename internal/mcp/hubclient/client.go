// Package hubclient provides an HTTP client for the kfleet hub API.
package hubclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const requestTimeout = 15 * time.Second

// Client calls the kfleet hub REST API.
type Client struct {
	hubURL     string
	token      string
	httpClient *http.Client
}

// New constructs a hub API client.
func New(hubURL, token string) (*Client, error) {
	hubURL = strings.TrimRight(strings.TrimSpace(hubURL), "/")
	if hubURL == "" {
		return nil, errors.New("hub URL is required")
	}
	parsed, err := url.Parse(hubURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid hub URL %q", hubURL)
	}
	return &Client{
		hubURL:     hubURL,
		token:      strings.TrimSpace(token),
		httpClient: &http.Client{Timeout: requestTimeout},
	}, nil
}

// ListClusters returns all clusters registered with the hub.
func (c *Client) ListClusters(ctx context.Context) ([]types.Cluster, error) {
	var response api.ListClustersResponse
	if err := c.get(ctx, "/api/v1/clusters", nil, &response); err != nil {
		return nil, err
	}
	return response.Clusters, nil
}

// GetClusterStatus returns the latest full status snapshot for a cluster.
func (c *Client) GetClusterStatus(ctx context.Context, id string) (api.ClusterStatusResponse, error) {
	var response api.ClusterStatusResponse
	err := c.get(ctx, clusterPath(id, "/status"), nil, &response)
	return response, err
}

// GetPods returns pods in a cluster, optionally limited to a namespace.
func (c *Client) GetPods(ctx context.Context, id, namespace string) ([]types.Pod, error) {
	query := url.Values{}
	if namespace = strings.TrimSpace(namespace); namespace != "" {
		query.Set("namespace", namespace)
	}
	var pods []types.Pod
	if err := c.get(ctx, clusterPath(id, "/pods"), query, &pods); err != nil {
		return nil, err
	}
	return pods, nil
}

// GetEvents returns events in a cluster, optionally limited to a namespace.
func (c *Client) GetEvents(ctx context.Context, id, namespace string) ([]types.Event, error) {
	query := url.Values{}
	if namespace = strings.TrimSpace(namespace); namespace != "" {
		query.Set("namespace", namespace)
	}
	var events []types.Event
	if err := c.get(ctx, clusterPath(id, "/events"), query, &events); err != nil {
		return nil, err
	}
	return events, nil
}

// TimelineQuery narrows a GetTimeline call. An empty ClusterID queries the
// whole fleet rather than a single cluster.
type TimelineQuery struct {
	ClusterID string
	Since     *time.Time
	Until     *time.Time
	Before    int64
	Limit     int
}

// GetTimeline returns a page of the durable operational event timeline,
// optionally scoped to one cluster and a time range.
func (c *Client) GetTimeline(ctx context.Context, q TimelineQuery) (api.ListTimelineEventsResponse, error) {
	query := url.Values{}
	if q.Since != nil {
		query.Set("since", q.Since.UTC().Format(time.RFC3339))
	}
	if q.Until != nil {
		query.Set("until", q.Until.UTC().Format(time.RFC3339))
	}
	if q.Before > 0 {
		query.Set("before", strconv.FormatInt(q.Before, 10))
	}
	if q.Limit > 0 {
		query.Set("limit", strconv.Itoa(q.Limit))
	}

	path := "/api/v1/timeline"
	if clusterID := strings.TrimSpace(q.ClusterID); clusterID != "" {
		path = clusterPath(clusterID, "/timeline")
	}

	var response api.ListTimelineEventsResponse
	if err := c.get(ctx, path, query, &response); err != nil {
		return api.ListTimelineEventsResponse{}, err
	}
	return response, nil
}

func clusterPath(id, suffix string) string {
	return "/api/v1/clusters/" + url.PathEscape(strings.TrimSpace(id)) + suffix
}

func (c *Client) get(ctx context.Context, path string, query url.Values, target any) error {
	endpoint := c.hubURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create hub request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call hub API: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		var apiError api.ErrorResponse
		if json.Unmarshal(body, &apiError) == nil && apiError.Error != "" {
			return fmt.Errorf("hub API returned %s: %s", response.Status, apiError.Error)
		}
		return fmt.Errorf("hub API returned %s", response.Status)
	}
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("decode hub response: %w", err)
	}
	return nil
}
