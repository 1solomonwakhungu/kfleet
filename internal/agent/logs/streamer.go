// Package logs streams pod logs from the local cluster for the kfleet agent.
//
// The hub has no Kubernetes credentials of its own, so the agent is the only
// component that can read pod logs. Client holds an outbound WebSocket to the
// hub and answers the log stream requests the hub pushes down it.
package logs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// maxLineBytes truncates pathologically long log lines so a single line
// cannot exhaust memory or overflow a WebSocket frame.
const maxLineBytes = 16 << 10

// Request describes one pod log stream.
type Request struct {
	Namespace string
	Pod       string
	Container string
	Follow    bool
	TailLines int64
}

// PodLogOpener opens a pod log stream. The clientset implementation is used
// in production; tests substitute a fake.
type PodLogOpener interface {
	OpenPodLogs(ctx context.Context, namespace, pod string, options *corev1.PodLogOptions) (io.ReadCloser, error)
}

type clientsetOpener struct {
	client kubernetes.Interface
}

func (o clientsetOpener) OpenPodLogs(ctx context.Context, namespace, pod string, options *corev1.PodLogOptions) (io.ReadCloser, error) {
	return o.client.CoreV1().Pods(namespace).GetLogs(pod, options).Stream(ctx)
}

// Streamer reads pod logs line by line.
type Streamer struct {
	opener PodLogOpener
}

// NewStreamer builds a streamer backed by a client-go clientset.
func NewStreamer(client kubernetes.Interface) *Streamer {
	return &Streamer{opener: clientsetOpener{client: client}}
}

// NewStreamerWithOpener builds a streamer over an explicit opener.
func NewStreamerWithOpener(opener PodLogOpener) *Streamer {
	return &Streamer{opener: opener}
}

// Stream reads pod logs and invokes emit for every line until the log ends,
// emit fails, or ctx is cancelled. Cancelling ctx closes the underlying
// Kubernetes stream, so a browser disconnect cannot leak a reader.
func (s *Streamer) Stream(ctx context.Context, request Request, emit func(line string) error) error {
	if strings.TrimSpace(request.Namespace) == "" || strings.TrimSpace(request.Pod) == "" {
		return errors.New("namespace and pod are required")
	}
	options := &corev1.PodLogOptions{
		Container: request.Container,
		Follow:    request.Follow,
	}
	if request.TailLines > 0 {
		tail := request.TailLines
		options.TailLines = &tail
	}

	reader, err := s.opener.OpenPodLogs(ctx, request.Namespace, request.Pod, options)
	if err != nil {
		return fmt.Errorf("open pod logs: %w", err)
	}
	finished := make(chan struct{})
	closeDone := make(chan struct{})
	go func() {
		defer close(closeDone)
		select {
		case <-ctx.Done():
			// Unblock a follow read that is parked waiting for output.
		case <-finished:
		}
		_ = reader.Close()
	}()
	defer func() {
		close(finished)
		<-closeDone
	}()

	buffered := bufio.NewReaderSize(reader, 32<<10)
	for {
		line, err := readLine(buffered)
		if err == nil || line != "" {
			if emitErr := emit(line); emitErr != nil {
				return emitErr
			}
		}
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read pod logs: %w", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// readLine returns the next line without its terminator, truncating lines
// longer than maxLineBytes and discarding the remainder of the line.
func readLine(reader *bufio.Reader) (string, error) {
	var builder strings.Builder
	for {
		chunk, isPrefix, err := reader.ReadLine()
		if len(chunk) > 0 && builder.Len() < maxLineBytes {
			remaining := maxLineBytes - builder.Len()
			if len(chunk) > remaining {
				chunk = chunk[:remaining]
			}
			builder.Write(chunk)
		}
		if err != nil {
			return builder.String(), err
		}
		if !isPrefix {
			return builder.String(), nil
		}
	}
}
