// Command webhook-receiver is a loopback-only signed webhook receiver for
// local kfleet alert development and failure-injection testing.
package main

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1solomonwakhungu/kfleet/internal/alerts"
)

const maxWebhookBodyBytes = 1 << 20

type receivedWebhook struct {
	ReceivedAt time.Time       `json:"receivedAt"`
	DeliveryID string          `json:"deliveryId"`
	Event      string          `json:"event"`
	Body       json.RawMessage `json:"body"`
}

type receiver struct {
	mu            sync.Mutex
	secret        string
	failRemaining int
	failStatus    int
	received      []receivedWebhook
}

func main() {
	listenAddr := envOrDefault("KFLEET_RECEIVER_LISTEN_ADDR", "127.0.0.1:9099")
	secret := os.Getenv("KFLEET_RECEIVER_SECRET")
	failRemaining, err := envInt("KFLEET_RECEIVER_FAIL_FIRST", 0)
	if err != nil || failRemaining < 0 {
		fatal("KFLEET_RECEIVER_FAIL_FIRST must be a non-negative integer")
	}
	failStatus, err := envInt("KFLEET_RECEIVER_FAIL_STATUS", http.StatusServiceUnavailable)
	if err != nil || failStatus < 400 || failStatus > 599 {
		fatal("KFLEET_RECEIVER_FAIL_STATUS must be an HTTP error status from 400 through 599")
	}
	if secret == "" {
		fatal("KFLEET_RECEIVER_SECRET is required")
	}
	if !isLoopbackAddress(listenAddr) {
		fatal("KFLEET_RECEIVER_LISTEN_ADDR must use a loopback host")
	}

	state := &receiver{
		secret:        secret,
		failRemaining: failRemaining,
		failStatus:    failStatus,
		received:      make([]receivedWebhook, 0),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhook", state.handleWebhook)
	mux.HandleFunc("GET /received", state.handleReceived)
	mux.HandleFunc("POST /control/failures", state.handleFailures)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	slog.Info("starting loopback webhook receiver",
		"address", listenAddr,
		"webhook_path", "/webhook",
		"fail_first", failRemaining,
		"fail_status", failStatus,
	)
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		slog.Error("webhook receiver stopped", "error", err)
		os.Exit(1)
	}
}

func (receiver *receiver) handleWebhook(w http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, request.Body, maxWebhookBodyBytes))
	if err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	timestamp := request.Header.Get("X-Kfleet-Timestamp")
	unixTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || time.Since(time.Unix(unixTimestamp, 0)).Abs() > 5*time.Minute {
		http.Error(w, "invalid timestamp", http.StatusUnauthorized)
		return
	}
	wantSignature := alerts.Signature(receiver.secret, timestamp, body)
	gotSignature := request.Header.Get("X-Kfleet-Signature")
	if len(wantSignature) != len(gotSignature) ||
		subtle.ConstantTimeCompare([]byte(wantSignature), []byte(gotSignature)) != 1 {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}
	if !json.Valid(body) {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	receiver.mu.Lock()
	defer receiver.mu.Unlock()
	receiver.received = append(receiver.received, receivedWebhook{
		ReceivedAt: time.Now().UTC(),
		DeliveryID: request.Header.Get("X-Kfleet-Delivery"),
		Event:      request.Header.Get("X-Kfleet-Event"),
		Body:       append(json.RawMessage(nil), body...),
	})
	if receiver.failRemaining > 0 {
		receiver.failRemaining--
		http.Error(w, "injected receiver failure", receiver.failStatus)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (receiver *receiver) handleReceived(w http.ResponseWriter, _ *http.Request) {
	receiver.mu.Lock()
	received := append([]receivedWebhook(nil), receiver.received...)
	receiver.mu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"received": received})
}

func (receiver *receiver) handleFailures(w http.ResponseWriter, request *http.Request) {
	var update struct {
		Count  int `json:"count"`
		Status int `json:"status"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, request.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&update); err != nil || update.Count < 0 ||
		(update.Status != 0 && (update.Status < 400 || update.Status > 599)) {
		http.Error(w, "expected {\"count\": non-negative integer, \"status\": 400..599}", http.StatusBadRequest)
		return
	}
	receiver.mu.Lock()
	receiver.failRemaining = update.Count
	if update.Status != 0 {
		receiver.failStatus = update.Status
	}
	response := map[string]int{
		"failRemaining": receiver.failRemaining,
		"failStatus":    receiver.failStatus,
	}
	receiver.mu.Unlock()
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to write receiver response", "error", err)
	}
}

func isLoopbackAddress(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func envInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func fatal(message string) {
	_, _ = fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
