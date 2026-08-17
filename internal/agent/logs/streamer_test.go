package logs

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestStreamerReadsFakeClientsetLogs(t *testing.T) {
	clientset := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "apps"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "api"}}},
	})
	streamer := NewStreamer(clientset)

	var lines []string
	err := streamer.Stream(context.Background(), Request{Namespace: "apps", Pod: "api", TailLines: 10}, func(line string) error {
		lines = append(lines, line)
		return nil
	})
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	if len(lines) == 0 {
		t.Fatalf("Stream() produced no lines")
	}
}

type stubOpener struct {
	reader  io.ReadCloser
	options *corev1.PodLogOptions
	err     error
}

func (o *stubOpener) OpenPodLogs(_ context.Context, _, _ string, options *corev1.PodLogOptions) (io.ReadCloser, error) {
	o.options = options
	if o.err != nil {
		return nil, o.err
	}
	return o.reader, nil
}

type nopCloser struct{ io.Reader }

func (nopCloser) Close() error { return nil }

func TestStreamerSplitsLinesAndPassesOptions(t *testing.T) {
	opener := &stubOpener{reader: nopCloser{strings.NewReader("first\nsecond\n\nthird")}}
	streamer := NewStreamerWithOpener(opener)

	var lines []string
	err := streamer.Stream(context.Background(), Request{
		Namespace: "apps", Pod: "api", Container: "sidecar", Follow: true, TailLines: 25,
	}, func(line string) error {
		lines = append(lines, line)
		return nil
	})
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	want := []string{"first", "second", "", "third"}
	if strings.Join(lines, "|") != strings.Join(want, "|") {
		t.Fatalf("lines = %#v, want %#v", lines, want)
	}
	if opener.options.Container != "sidecar" || !opener.options.Follow {
		t.Fatalf("options = %#v, want sidecar/follow", opener.options)
	}
	if opener.options.TailLines == nil || *opener.options.TailLines != 25 {
		t.Fatalf("tail lines = %v, want 25", opener.options.TailLines)
	}
}

func TestStreamerRequiresNamespaceAndPod(t *testing.T) {
	streamer := NewStreamerWithOpener(&stubOpener{reader: nopCloser{strings.NewReader("")}})
	if err := streamer.Stream(context.Background(), Request{Pod: "api"}, func(string) error { return nil }); err == nil {
		t.Fatalf("Stream() error = nil, want validation error")
	}
}

func TestStreamerReportsOpenFailure(t *testing.T) {
	streamer := NewStreamerWithOpener(&stubOpener{err: errors.New("forbidden")})
	err := streamer.Stream(context.Background(), Request{Namespace: "apps", Pod: "api"}, func(string) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "forbidden") {
		t.Fatalf("Stream() error = %v, want open failure", err)
	}
}

func TestStreamerTruncatesLongLines(t *testing.T) {
	opener := &stubOpener{reader: nopCloser{strings.NewReader(strings.Repeat("x", maxLineBytes*2) + "\ntail\n")}}
	streamer := NewStreamerWithOpener(opener)

	var lines []string
	if err := streamer.Stream(context.Background(), Request{Namespace: "apps", Pod: "api"}, func(line string) error {
		lines = append(lines, line)
		return nil
	}); err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	if len(lines) != 2 || len(lines[0]) != maxLineBytes || lines[1] != "tail" {
		t.Fatalf("lines = %d, first length = %d", len(lines), len(lines[0]))
	}
}

// blockingReader never returns data until it is closed, mimicking a
// follow=true stream sitting idle.
type blockingReader struct {
	closeOnce sync.Once
	closed    chan struct{}
}

func newBlockingReader() *blockingReader {
	return &blockingReader{closed: make(chan struct{})}
}

func (r *blockingReader) Read([]byte) (int, error) {
	<-r.closed
	return 0, io.EOF
}

func (r *blockingReader) Close() error {
	r.closeOnce.Do(func() { close(r.closed) })
	return nil
}

func TestStreamerStopsWhenContextIsCancelled(t *testing.T) {
	reader := newBlockingReader()
	streamer := NewStreamerWithOpener(&stubOpener{reader: reader})
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- streamer.Stream(ctx, Request{Namespace: "apps", Pod: "api", Follow: true}, func(string) error { return nil })
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Stream() error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Stream() did not stop after cancellation")
	}
	select {
	case <-reader.closed:
	default:
		t.Fatal("Stream() left the Kubernetes reader open")
	}
}

func TestLogChannelEndpoint(t *testing.T) {
	tests := []struct {
		hubURL string
		want   string
	}{
		{"http://hub.example:8080", "ws://hub.example:8080/api/v1/agents/prod/logs"},
		{"https://hub.example/", "wss://hub.example/api/v1/agents/prod/logs"},
	}
	for _, tt := range tests {
		if got := logChannelEndpoint(tt.hubURL, "prod"); got != tt.want {
			t.Fatalf("logChannelEndpoint(%q) = %q, want %q", tt.hubURL, got, tt.want)
		}
	}
}
