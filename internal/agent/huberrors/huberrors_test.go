package huberrors

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDetail(t *testing.T) {
	tests := []struct {
		name     string
		response *http.Response
		want     string
	}{
		{
			name:     "nil response",
			response: nil,
			want:     "",
		},
		{
			name:     "json error contract",
			response: responseWithBody(`{"error":"agent is pending approval","code":403}`),
			want:     "agent is pending approval",
		},
		{
			name:     "json without error field",
			response: responseWithBody(`{"code":403}`),
			want:     `{"code":403}`,
		},
		{
			name:     "plain body",
			response: responseWithBody("rate limited\n"),
			want:     "rate limited",
		},
		{
			name:     "empty body",
			response: responseWithBody(""),
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Detail(tt.response); got != tt.want {
				t.Errorf("Detail() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetailBoundsHugeBody(t *testing.T) {
	body := strings.Repeat("x", maxDetailBytes*4)
	got := Detail(responseWithBody(body))
	if len(got) != maxDetailBytes {
		t.Errorf("Detail() length = %d, want bounded to %d", len(got), maxDetailBytes)
	}
}

func TestWithDetail(t *testing.T) {
	if got := WithDetail("hub returned status 403 Forbidden", responseWithBody(`{"error":"agent is pending approval"}`)); got != "hub returned status 403 Forbidden: agent is pending approval" {
		t.Errorf("WithDetail() = %q, want joined message", got)
	}
	if got := WithDetail("hub returned status 401 Unauthorized", responseWithBody("")); got != "hub returned status 401 Unauthorized" {
		t.Errorf("WithDetail() = %q, want base unchanged without a body", got)
	}
}

func responseWithBody(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusForbidden,
		Status:     "403 Forbidden",
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{},
	}
}
