// Package mcp exposes kfleet cluster data through the Model Context Protocol.
package mcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/mcp/hubclient"
	"github.com/1solomonwakhungu/kfleet/internal/version"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const runtimeProbeTimeout = 10 * time.Second

// Server is the kfleet MCP server.
type Server struct {
	server *mcpserver.MCPServer
}

// NewServer constructs an MCP server and registers all kfleet tools.
func NewServer(hub *hubclient.Client) *Server {
	server := mcpserver.NewMCPServer(
		"kfleet",
		version.String(),
		mcpserver.WithToolCapabilities(false),
		mcpserver.WithInstructions("Inspect and diagnose Kubernetes clusters managed by kfleet."),
	)
	registerTools(server, hub)
	return &Server{server: server}
}

// Serve runs the MCP server over standard input and output until ctx ends.
func (s *Server) Serve(ctx context.Context) error {
	if s == nil || s.server == nil {
		return errors.New("MCP server is not initialized")
	}
	return mcpserver.NewStdioServer(s.server).Listen(ctx, os.Stdin, os.Stdout)
}

// RunStdio configures and runs the MCP server using KFLEET_HUB_URL and, unless
// the target hub reports demoMode=true, KFLEET_HUB_USERNAME and
// KFLEET_HUB_PASSWORD.
func RunStdio() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	hub, err := configureHubClient(
		ctx,
		os.Getenv("KFLEET_HUB_URL"),
		os.Getenv("KFLEET_HUB_USERNAME"),
		os.Getenv("KFLEET_HUB_PASSWORD"),
	)
	if err != nil {
		return err
	}
	return NewServer(hub).Serve(ctx)
}

func configureHubClient(ctx context.Context, hubURL, username, password string) (*hubclient.Client, error) {
	hubURL = strings.TrimSpace(hubURL)
	if hubURL == "" {
		return nil, errors.New("KFLEET_HUB_URL is required")
	}
	username = strings.TrimSpace(username)
	if username != "" && password != "" {
		hub, err := hubclient.NewWithCredentials(hubURL, username, password)
		if err != nil {
			return nil, fmt.Errorf("configure hub client: %w", err)
		}
		return hub, nil
	}
	if username != "" || password != "" {
		return nil, errors.New("KFLEET_HUB_USERNAME and KFLEET_HUB_PASSWORD must both be set")
	}

	hub, err := hubclient.New(hubURL, "")
	if err != nil {
		return nil, fmt.Errorf("configure hub client: %w", err)
	}
	probeCtx, cancel := context.WithTimeout(ctx, runtimeProbeTimeout)
	defer cancel()
	info, err := hub.GetRuntimeInfo(probeCtx)
	if err != nil {
		return nil, fmt.Errorf("KFLEET_HUB_USERNAME and KFLEET_HUB_PASSWORD are required (failed to probe hub demo mode: %w)", err)
	}
	if !info.DemoMode {
		return nil, errors.New("KFLEET_HUB_USERNAME and KFLEET_HUB_PASSWORD are required")
	}
	return hub, nil
}
