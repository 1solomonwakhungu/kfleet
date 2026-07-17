// Command hub runs the kfleet hub server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/1solomonwakhungu/kfleet/internal/config"
	kfleetmcp "github.com/1solomonwakhungu/kfleet/internal/mcp"
	"github.com/1solomonwakhungu/kfleet/internal/server"
	"github.com/1solomonwakhungu/kfleet/internal/store"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		if err := kfleetmcp.RunStdio(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "kfleet MCP server: %v\n", err)
			os.Exit(1)
		}
		return
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel(cfg.LogLevel),
	}))
	slog.SetDefault(logger)
	st, err := store.Open(cfg.DBPath)
	if err != nil {
		logger.Error("failed to open cluster store", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := st.Close(); err != nil {
			logger.Error("failed to close cluster store", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := server.New(cfg, logger, st)
	logger.Info("starting hub server", "address", cfg.ListenAddr)
	if err := srv.Start(ctx); err != nil {
		logger.Error("hub server stopped with an error", "error", err)
		os.Exit(1)
	}
	logger.Info("hub server stopped")
}

func logLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
