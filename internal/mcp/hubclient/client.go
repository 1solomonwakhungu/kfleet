// Package hubclient provides an HTTP client for the kfleet hub API.
package hubclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
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

func clusterPath(id, suffix string) string {
	return "/api/v1/clusters/" + url.PathEscape(strings.TrimSpace(id)) + suffix
}

func (c *Client) get(ctx context.Context, path string, query url.Values, target any) error {
	if c.username != "" {
		if err := c.ensureAuthenticated(ctx); err != nil {
			return err
		}
	}

	response, err := c.doGet(ctx, path, query)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusUnauthorized && c.username != "" {
		_ = response.Body.Close()
		c.authMu.Lock()
		c.loggedIn = false
		c.authMu.Unlock()
		if err := c.ensureAuthenticated(ctx); err != nil {
			return err
		}
		response, err = c.doGet(ctx, path, query)
		if err != nil {
			return err
		}
	}
	defer func() { _ = response.Body.Close() }()
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

func (c *Client) doGet(ctx context.Context, path string, query url.Values) (*http.Response, error) {
	endpoint := c.hubURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create hub request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
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
