// Package hubclient provides an HTTP client for the kfleet hub API.
package hubclient

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

const requestTimeout = 15 * time.Second

// Client calls the kfleet hub REST API.
type Client struct {
	hubURL     string
	token      string
	username   string
	password   string
	httpClient *http.Client
	authMu     sync.Mutex
	loggedIn   bool
}

// NewWithCredentials constructs a hub client that establishes a user session
// with the supplied credentials before its first API request. A dedicated
// read_only account is recommended for MCP integrations.
func NewWithCredentials(hubURL, username, password string) (*Client, error) {
	client, err := New(hubURL, "")
	if err != nil {
		return nil, err
	}
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, errors.New("hub username and password are required")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}
	client.username = username
	client.password = password
	client.httpClient.Jar = jar
	return client, nil
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

// ListNamespaces returns the namespace names in a cluster's latest snapshot.
func (c *Client) ListNamespaces(ctx context.Context, id string) ([]string, error) {
	var namespaces []string
	if err := c.get(ctx, clusterPath(id, "/namespaces"), nil, &namespaces); err != nil {
		return nil, err
	}
	return namespaces, nil
}

// PodLogsQuery selects the log output GetPodLogs fetches.
type PodLogsQuery struct {
	ClusterID string
	Namespace string
	Pod       string
	// Container is optional for single-container pods.
	Container string
	// TailLines bounds the number of lines returned; the hub applies its
	// own default and cap when it is zero or too large.
	TailLines int64
	// Previous reads the previous terminated container instance.
	Previous bool
	// SinceSeconds limits output to lines newer than this many seconds.
	SinceSeconds int64
}

// sseMaxLineBytes bounds one SSE line. Agents truncate log lines at 16 KiB,
// so this leaves ample headroom for the field prefix.
const sseMaxLineBytes = 1 << 20

// GetPodLogs fetches a bounded, non-following snapshot of pod logs. The hub
// streams logs as server-sent events; this reads the stream until the hub's
// end event and returns the lines in order.
func (c *Client) GetPodLogs(ctx context.Context, q PodLogsQuery) ([]string, error) {
	namespace := strings.TrimSpace(q.Namespace)
	pod := strings.TrimSpace(q.Pod)
	if namespace == "" || pod == "" {
		return nil, errors.New("namespace and pod are required")
	}
	query := url.Values{}
	query.Set("follow", "false")
	if container := strings.TrimSpace(q.Container); container != "" {
		query.Set("container", container)
	}
	if q.TailLines > 0 {
		query.Set("tailLines", strconv.FormatInt(q.TailLines, 10))
	}
	if q.Previous {
		query.Set("previous", "true")
	}
	if q.SinceSeconds > 0 {
		query.Set("sinceSeconds", strconv.FormatInt(q.SinceSeconds, 10))
	}
	path := clusterPath(q.ClusterID, "/pods/"+url.PathEscape(namespace)+"/"+url.PathEscape(pod)+"/logs")

	response, err := c.open(ctx, path, query, "text/event-stream")
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	return readLogEvents(response.Body)
}

// readLogEvents consumes the hub's pod log SSE stream. Unnamed events carry
// one log line each, "end" closes a completed stream, and "stream-error"
// carries a JSON {"message": ...} describing why the stream failed.
func readLogEvents(body io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64<<10), sseMaxLineBytes)

	lines := make([]string, 0)
	event := ""
	var data []string
	hasData := false
	for scanner.Scan() {
		field := scanner.Text()
		if field == "" {
			switch event {
			case "", "message":
				if hasData {
					lines = append(lines, strings.Join(data, "\n"))
				}
			case "end":
				return lines, nil
			case "stream-error":
				var payload struct {
					Message string `json:"message"`
				}
				message := strings.Join(data, "\n")
				if json.Unmarshal([]byte(message), &payload) == nil && payload.Message != "" {
					message = payload.Message
				}
				return nil, fmt.Errorf("pod log stream failed: %s", message)
			}
			event, data, hasData = "", nil, false
			continue
		}
		if strings.HasPrefix(field, ":") {
			continue
		}
		name, value, _ := strings.Cut(field, ":")
		value = strings.TrimPrefix(value, " ")
		switch name {
		case "event":
			event = value
		case "data":
			data = append(data, value)
			hasData = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read pod log stream: %w", err)
	}
	return nil, errors.New("pod log stream ended before the hub finished it")
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
	response, err := c.open(ctx, path, query, "application/json")
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("decode hub response: %w", err)
	}
	return nil
}

// open issues an authenticated GET, re-establishing an expired session once,
// and returns the response only when it carries a 2xx status. The caller owns
// the response body.
func (c *Client) open(ctx context.Context, path string, query url.Values, accept string) (*http.Response, error) {
	if c.username != "" {
		if err := c.ensureAuthenticated(ctx); err != nil {
			return nil, err
		}
	}

	response, err := c.doGet(ctx, path, query, accept)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == http.StatusUnauthorized && c.username != "" {
		_ = response.Body.Close()
		c.authMu.Lock()
		c.loggedIn = false
		c.authMu.Unlock()
		if err := c.ensureAuthenticated(ctx); err != nil {
			return nil, err
		}
		response, err = c.doGet(ctx, path, query, accept)
		if err != nil {
			return nil, err
		}
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		defer func() { _ = response.Body.Close() }()
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		var apiError api.ErrorResponse
		if json.Unmarshal(body, &apiError) == nil && apiError.Error != "" {
			return nil, fmt.Errorf("hub API returned %s: %s", response.Status, apiError.Error)
		}
		return nil, fmt.Errorf("hub API returned %s", response.Status)
	}
	return response, nil
}

func (c *Client) doGet(ctx context.Context, path string, query url.Values, accept string) (*http.Response, error) {
	endpoint := c.hubURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create hub request: %w", err)
	}
	req.Header.Set("Accept", accept)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call hub API: %w", err)
	}
	return response, nil
}

func (c *Client) ensureAuthenticated(ctx context.Context) error {
	c.authMu.Lock()
	defer c.authMu.Unlock()
	if c.loggedIn {
		return nil
	}

	payload, err := json.Marshal(api.LoginRequest{Username: c.username, Password: c.password})
	if err != nil {
		return fmt.Errorf("encode hub login: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.hubURL+"/api/v1/auth/login", strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("create hub login request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("authenticate with hub: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		var apiError api.ErrorResponse
		if json.Unmarshal(body, &apiError) == nil && apiError.Error != "" {
			return fmt.Errorf("hub login returned %s: %s", response.Status, apiError.Error)
		}
		return fmt.Errorf("hub login returned %s", response.Status)
	}
	c.loggedIn = true
	return nil
}
