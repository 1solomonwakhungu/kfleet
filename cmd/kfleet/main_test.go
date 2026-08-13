package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRunner struct {
	missing string
	runs    []string
	output  []byte
	pid     int
}

func (r *fakeRunner) LookPath(name string) error {
	if name == r.missing {
		return errors.New("missing")
	}
	return nil
}

func (r *fakeRunner) Run(_ context.Context, name string, args ...string) error {
	r.runs = append(r.runs, strings.Join(append([]string{name}, args...), " "))
	return nil
}

func (r *fakeRunner) Output(_ context.Context, name string, args ...string) ([]byte, error) {
	r.runs = append(r.runs, strings.Join(append([]string{name}, args...), " "))
	return r.output, nil
}

func (r *fakeRunner) Start(_ context.Context, name string, args ...string) (int, error) {
	r.runs = append(r.runs, strings.Join(append([]string{name}, args...), " "))
	return r.pid, nil
}

func newTestApp(t *testing.T, runner commandRunner) (app, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	return app{
		runner:    runner,
		stdout:    stdout,
		stderr:    stderr,
		statePath: filepath.Join(t.TempDir(), "quickstart.json"),
		http:      http.DefaultClient,
		open:      func(string) error { return nil },
	}, stdout, stderr
}

func TestRunHelpAndVersion(t *testing.T) {
	a, stdout, _ := newTestApp(t, &fakeRunner{})
	if err := a.run(context.Background(), []string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "kfleet quickstart") {
		t.Fatalf("help output = %q", stdout.String())
	}
	stdout.Reset()
	if err := a.run(context.Background(), []string{"version"}); err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != "kfleet dev\n" {
		t.Fatalf("version output = %q", got)
	}
}

func TestQuickstartRequiresReleasedVersion(t *testing.T) {
	a, _, _ := newTestApp(t, &fakeRunner{})
	err := a.run(context.Background(), []string{"quickstart"})
	if err == nil || !strings.Contains(err.Error(), "release version is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestQuickstartChecksDependencies(t *testing.T) {
	runner := &fakeRunner{missing: "helm"}
	a, _, _ := newTestApp(t, runner)
	err := a.run(context.Background(), []string{"quickstart", "--version", "1.2.3"})
	if err == nil || !strings.Contains(err.Error(), "helm is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestCleanupUsesRecordedClusterCount(t *testing.T) {
	runner := &fakeRunner{}
	a, stdout, _ := newTestApp(t, runner)
	if err := a.writeState(state{Clusters: 2, Port: 8080}); err != nil {
		t.Fatal(err)
	}
	if err := a.run(context.Background(), []string{"cleanup"}); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(runner.runs, "\n")
	if !strings.Contains(joined, "kind delete cluster --name kfleet-2") {
		t.Fatalf("commands = %s", joined)
	}
	if strings.Contains(joined, "kfleet-3") {
		t.Fatalf("unexpected third cluster command: %s", joined)
	}
	if !strings.Contains(stdout.String(), "Cleanup complete") {
		t.Fatalf("output = %q", stdout.String())
	}
}

func TestStatusReportsReadyHub(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	a, stdout, _ := newTestApp(t, &fakeRunner{})
	a.http = server.Client()
	port := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	var parsedPort int
	if _, err := fmt.Sscanf(port, "%d", &parsedPort); err != nil {
		t.Fatal(err)
	}
	if err := a.writeState(state{Clusters: 3, Port: parsedPort}); err != nil {
		t.Fatal(err)
	}
	if err := a.run(context.Background(), []string{"status"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "with 3 clusters") {
		t.Fatalf("output = %q", stdout.String())
	}
}

func TestOpenUsesRecordedPort(t *testing.T) {
	a, _, _ := newTestApp(t, &fakeRunner{})
	if err := a.writeState(state{Port: 9090}); err != nil {
		t.Fatal(err)
	}
	var opened string
	a.open = func(url string) error {
		opened = url
		return nil
	}
	if err := a.run(context.Background(), []string{"open"}); err != nil {
		t.Fatal(err)
	}
	if opened != "http://localhost:9090" {
		t.Fatalf("opened = %q", opened)
	}
}

func TestTemporaryValuesProtectsSecrets(t *testing.T) {
	path, err := temporaryValues(map[string]any{"token": "secret"})
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("mode = %o, want 600", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"token":"secret"}` {
		t.Fatalf("content = %q", data)
	}
}
