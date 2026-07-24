// Package registrar manages agent registration and lifecycle calls to the hub.
package registrar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/agent/config"
)

const (
	requestTimeout = 10 * time.Second
	agentVersion   = "0.1.0"
)

// Registrar communicates with the hub's agent lifecycle API.
type Registrar struct {
	hubURL            string
	registrationToken string
	token             string
	clusterName       string
	agentVersion      string
	labels            map[string]string
	client            *http.Client
}

// RegisterRequest is sent to register or re-register an agent.
type RegisterRequest struct {
	ClusterName  string            `json:"clusterName"`
	Labels       map[string]string `json:"labels"`
	AgentVersion string            `json:"agentVersion"`
	K8sVersion   string            `json:"k8sVersion"`
}

// RegisterResponse describes the hub's registration decision.
type RegisterResponse struct {
	ClusterID string `json:"clusterId"`
	Approved  bool   `json:"approved"`
}

// New constructs a registrar from agent configuration.
func New(cfg *config.Config, labels map[string]string) *Registrar {
	return &Registrar{
		hubURL:            strings.TrimRight(cfg.HubURL, "/"),
		registrationToken: cfg.HubToken,
		token:             cfg.HubToken,
		clusterName:       cfg.ClusterName,
		agentVersion:      agentVersion,
		labels:            labels,
		client:            &http.Client{Timeout: requestTimeout},
	}
}

// Register registers the agent. HTTP 200 means approved and HTTP 202 means pending.
func (r *Registrar) Register(ctx context.Context, k8sVersion string) (*RegisterResponse, error) {
	payload := RegisterRequest{
		ClusterName:  r.clusterName,
		Labels:       r.labels,
		AgentVersion: r.agentVersion,
		K8sVersion:   k8sVersion,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode registration request: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, r.hubURL+"/api/v1/agents/register", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create registration request: %w", err)
	}
	r.setHeaders(request, true, r.registrationToken)
	response, err := r.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("register agent: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusUnauthorized {
		_, _ = io.Copy(io.Discard, response.Body)
		return nil, errors.New("hub rejected agent token")
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusAccepted && response.StatusCode != http.StatusCreated {
		_, _ = io.Copy(io.Discard, response.Body)
		return nil, fmt.Errorf("hub returned registration status %s", response.Status)
	}

	var result struct {
		RegisterResponse
		Token string `json:"token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode registration response: %w", err)
	}
	if result.ClusterID == "" {
		return nil, errors.New("hub returned an empty cluster ID")
	}
	if result.Token != "" {
		r.token = result.Token
	}
	result.Approved = response.StatusCode == http.StatusOK
	return &result.RegisterResponse, nil
}

// Heartbeat tells the hub that the agent process remains reachable.
func (r *Registrar) Heartbeat(ctx context.Context) error {
	return r.postLifecycle(ctx, "heartbeat")
}

// Deregister marks the agent unreachable during graceful shutdown.
func (r *Registrar) Deregister(ctx context.Context) error {
	return r.postLifecycle(ctx, "deregister")
}

// Token returns the current agent token, including one rotated during registration.
func (r *Registrar) Token() string {
	return r.token
}

func (r *Registrar) postLifecycle(ctx context.Context, action string) error {
	endpoint := r.hubURL + "/api/v1/agents/" + url.PathEscape(r.clusterName) + "/" + action
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create agent %s request: %w", action, err)
	}
	r.setHeaders(request, false, r.token)
	response, err := r.client.Do(request)
	if err != nil {
		return fmt.Errorf("send agent %s: %w", action, err)
	}
	defer func() { _ = response.Body.Close() }()
	_, _ = io.Copy(io.Discard, response.Body)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("hub returned %s status %s", action, response.Status)
	}
	return nil
}

func (r *Registrar) setHeaders(request *http.Request, jsonBody bool, token string) {
	if jsonBody {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Authorization", "Bearer "+token)
}
