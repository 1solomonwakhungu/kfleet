// Package huberrors turns failed hub HTTP responses into operator-readable
// error text, so kubectl logs explains why the hub rejected an agent call.
package huberrors

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/1solomonwakhungu/kfleet/pkg/api"
)

// maxDetailBytes bounds how much of an error response body is read. Hub error
// bodies are short JSON documents; the limit keeps a misbehaving hub from
// flooding agent logs.
const maxDetailBytes = 8 << 10

// Detail drains and closes the response body and returns the hub's error
// message. The hub's JSON error contract ({"error": "...", "code": ...})
// yields its error field; any other body yields a trimmed snippet. An empty
// or unreadable body yields an empty string.
func Detail(response *http.Response) string {
	if response == nil || response.Body == nil {
		return ""
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxDetailBytes))
	if err != nil {
		return ""
	}
	snippet := strings.TrimSpace(string(body))
	if snippet == "" {
		return ""
	}
	var apiError api.ErrorResponse
	if err := json.Unmarshal(body, &apiError); err == nil && strings.TrimSpace(apiError.Error) != "" {
		return strings.TrimSpace(apiError.Error)
	}
	return snippet
}

// WithDetail appends the response's error detail to a base message as
// "base: detail", leaving base unchanged when the body carries no detail.
// The response body is drained and closed.
func WithDetail(base string, response *http.Response) string {
	if detail := Detail(response); detail != "" {
		return base + ": " + detail
	}
	return base
}
