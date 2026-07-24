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

	"github.com/1solomonwakhungu/kfleet/internal/mcp/hubclient"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const serverVersion = "0.1.0"

// Server is the kfleet MCP server.
type Server struct {
	server *mcpserver.MCPServer
}

// NewServer constructs an MCP server and registers all kfleet tools.
func NewServer(hub *hubclient.Client) *Server {
	server := mcpserver.NewMCPServer(
		"kfleet",
		serverVersion,
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

// RunStdio configures and runs the MCP server using KFLEET_HUB_URL,
// KFLEET_HUB_USERNAME, and KFLEET_HUB_PASSWORD.
func RunStdio() error {
	hubURL := strings.TrimSpace(os.Getenv("KFLEET_HUB_URL"))
	if hubURL == "" {
		return errors.New("KFLEET_HUB_URL is required")
	}
	username := strings.TrimSpace(os.Getenv("KFLEET_HUB_USERNAME"))
	password := os.Getenv("KFLEET_HUB_PASSWORD")
	if username == "" || password == "" {
		return errors.New("KFLEET_HUB_USERNAME and KFLEET_HUB_PASSWORD are required")
	}
	hub, err := hubclient.NewWithCredentials(hubURL, username, password)
	if err != nil {
		return fmt.Errorf("configure hub client: %w", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return NewServer(hub).Serve(ctx)
}
