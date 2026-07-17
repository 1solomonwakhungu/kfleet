// Package registration communicates with the hub registration API.
package registration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/collector"
)

const requestTimeout = 10 * time.Second

// Client communicates with the hub registration API.
type Client struct {
	hubURL     string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewClient constructs a registration client.
func NewClient(hubURL string, logger *slog.Logger) *Client {
	return &Client{
		hubURL:     strings.TrimRight(hubURL, "/"),
		httpClient: &http.Client{Timeout: requestTimeout},
		logger:     logger,
	}
}

// Register registers a named cluster with the hub and returns its agent token.
func (c *Client) Register(ctx context.Context, clusterName string) (token string, err error) {
	payload := struct {
		ClusterName string `json:"clusterName"`
	}{ClusterName: clusterName}

	var response struct {
		Token string `json:"token"`
	}
	if err := c.postJSON(ctx, "/api/v1/agents/register", "", payload, &response); err != nil {
		return "", err
	}
	if response.Token == "" {
		return "", fmt.Errorf("hub returned an empty registration token")
	}
	c.logger.Debug("cluster registration completed", "cluster", clusterName)
	return response.Token, nil
}

// Heartbeat sends a cluster snapshot to the hub.
func (c *Client) Heartbeat(ctx context.Context, token string, snap collector.ClusterSnapshot) error {
	return c.postJSON(ctx, "/api/v1/agents/heartbeat", token, snap, nil)
}

func (c *Client) postJSON(ctx context.Context, path, token string, payload, response any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.hubURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("hub returned status %s", resp.Status)
	}
	if response == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
