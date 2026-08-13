package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var version = "dev"

const (
	hubChart   = "oci://ghcr.io/1solomonwakhungu/charts/kfleet-hub"
	agentChart = "oci://ghcr.io/1solomonwakhungu/charts/kfleet-agent"
)

type state struct {
	Clusters      int    `json:"clusters"`
	Port          int    `json:"port"`
	PortForwardID int    `json:"portForwardPid"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Token         string `json:"token"`
}

type commandRunner interface {
	LookPath(string) error
	Run(context.Context, string, ...string) error
	Output(context.Context, string, ...string) ([]byte, error)
	Start(context.Context, string, ...string) (int, error)
}

type execRunner struct {
	stdout io.Writer
	stderr io.Writer
}

func (r execRunner) LookPath(name string) error {
	_, err := exec.LookPath(name)
	return err
}

func (r execRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", name, err)
	}
	return nil
}

func (r execRunner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	output, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("%s failed: %w: %s", name, err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func (r execRunner) Start(ctx context.Context, name string, args ...string) (int, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start %s: %w", name, err)
	}
	go func() { _ = cmd.Wait() }()
	return cmd.Process.Pid, nil
}

type app struct {
	runner    commandRunner
	stdout    io.Writer
	stderr    io.Writer
	statePath string
	http      *http.Client
	open      func(string) error
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	statePath, err := defaultStatePath()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	a := app{
		runner:    execRunner{stdout: os.Stdout, stderr: os.Stderr},
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		statePath: statePath,
		http:      &http.Client{Timeout: 5 * time.Second},
		open:      openBrowser,
	}
	if err := a.run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func (a app) run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		a.usage()
		return nil
	}
	switch args[0] {
	case "quickstart":
		return a.quickstart(ctx, args[1:])
	case "cleanup":
		return a.cleanup(ctx, args[1:])
	case "status":
		return a.status(ctx, args[1:])
	case "open":
		return a.openUI(args[1:])
	case "version", "--version", "-v":
		fmt.Fprintf(a.stdout, "kfleet %s\n", version)
		return nil
	case "help", "--help", "-h":
		a.usage()
		return nil
	default:
		return fmt.Errorf("unknown command %q; run 'kfleet help'", args[0])
	}
}

func (a app) usage() {
	fmt.Fprintln(a.stdout, `kfleet manages a local multi-cluster demo.

Usage:
  kfleet quickstart [--clusters 3] [--port 8080] [--version VERSION]
  kfleet status
  kfleet open
  kfleet cleanup [--clusters 3]
  kfleet version`)
}

func (a app) quickstart(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("quickstart", flag.ContinueOnError)
	flags.SetOutput(a.stderr)
	clusters := flags.Int("clusters", envInt("KFLEET_CLUSTERS", 3), "number of kind clusters")
	port := flags.Int("port", envInt("KFLEET_HUB_PORT", 8080), "local hub port")
	releaseVersion := flags.String("version", defaultReleaseVersion(), "kfleet release version")
	username := flags.String("username", envString("KFLEET_ADMIN_USERNAME", "admin"), "admin username")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if *clusters < 1 {
		return errors.New("clusters must be at least 1")
	}
	if *port < 1 || *port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	chartVersion := strings.TrimPrefix(*releaseVersion, "v")
	if chartVersion == "" || chartVersion == "dev" {
		return errors.New("a release version is required; use --version or install a released build")
	}
	imageTag := "v" + chartVersion

	for _, dependency := range []string{"docker", "kind", "kubectl", "helm"} {
		if err := a.runner.LookPath(dependency); err != nil {
			return fmt.Errorf("%s is required but was not found in PATH", dependency)
		}
	}
	if err := a.runner.Run(ctx, "docker", "info"); err != nil {
		return errors.New("Docker is installed but is not running")
	}

	previous, _ := a.readState()
	token := previous.Token
	if token == "" {
		generatedToken, randomErr := randomHex(16)
		if randomErr != nil {
			return fmt.Errorf("generate registration token: %w", randomErr)
		}
		token = generatedToken
	}
	password := envString("KFLEET_ADMIN_PASSWORD", previous.Password)
	if password == "" {
		generatedPassword, randomErr := randomHex(16)
		if randomErr != nil {
			return fmt.Errorf("generate admin password: %w", randomErr)
		}
		password = generatedPassword
	}
	if previous.Username != "" && os.Getenv("KFLEET_ADMIN_USERNAME") == "" {
		*username = previous.Username
	}
	hubValues, err := temporaryValues(map[string]any{
		"registration": map[string]any{"token": token},
		"auth": map[string]any{
			"sessionCookieInsecure": true,
			"bootstrapAdmin": map[string]any{
				"username": *username,
				"email":    *username + "@kfleet.local",
				"password": password,
			},
		},
	})
	if err != nil {
		return err
	}
	defer os.Remove(hubValues)

	fmt.Fprintf(a.stdout, "Creating %d kind clusters...\n", *clusters)
	existingOutput, err := a.runner.Output(ctx, "kind", "get", "clusters")
	if err != nil {
		return err
	}
	existing := "\n" + string(existingOutput) + "\n"
	for i := 1; i <= *clusters; i++ {
		name := fmt.Sprintf("kfleet-%d", i)
		if strings.Contains(existing, "\n"+name+"\n") {
			fmt.Fprintf(a.stdout, "%s already exists.\n", name)
			continue
		}
		if err := a.runner.Run(ctx, "kind", "create", "cluster", "--name", name); err != nil {
			return err
		}
	}

	fmt.Fprintln(a.stdout, "Installing the hub...")
	if err := a.runner.Run(ctx, "helm", "upgrade", "--install", "kfleet-hub", hubChart,
		"--version", chartVersion,
		"--kube-context", "kind-kfleet-1",
		"--values", hubValues,
		"--set", "image.repository=ghcr.io/1solomonwakhungu/kfleet/hub",
		"--set", "image.tag="+imageTag,
		"--set", "service.type=ClusterIP",
		"--set", "persistence.enabled=true",
		"--wait", "--timeout", "120s"); err != nil {
		return err
	}

	stopProcess(previous.PortForwardID)
	portForwardArgs := []string{"--context", "kind-kfleet-1", "port-forward", "svc/kfleet-hub", fmt.Sprintf("%d:8080", *port)}
	hubURL := fmt.Sprintf("http://host.docker.internal:%d", *port)
	if runtime.GOOS == "linux" {
		gateway, outputErr := a.runner.Output(ctx, "docker", "network", "inspect", "kind", "--format", "{{(index .IPAM.Config 0).Gateway}}")
		if outputErr != nil {
			return fmt.Errorf("find the kind network gateway: %w", outputErr)
		}
		gatewayAddress := strings.TrimSpace(string(gateway))
		if gatewayAddress == "" {
			return errors.New("the kind network has no gateway address")
		}
		portForwardArgs = append(portForwardArgs, "--address", gatewayAddress+",localhost")
		hubURL = fmt.Sprintf("http://%s:%d", gatewayAddress, *port)
	}
	pid, err := a.runner.Start(context.Background(), "kubectl", portForwardArgs...)
	if err != nil {
		return err
	}
	keepPortForward := false
	defer func() {
		if !keepPortForward {
			stopProcess(pid)
		}
	}()
	current := state{Clusters: *clusters, Port: *port, PortForwardID: pid, Username: *username, Password: password, Token: token}
	if err := a.writeState(current); err != nil {
		stopProcess(pid)
		return err
	}
	if err := a.waitForHub(ctx, *port); err != nil {
		return err
	}

	for i := 1; i <= *clusters; i++ {
		name := fmt.Sprintf("kfleet-%d", i)
		fmt.Fprintf(a.stdout, "Installing the agent on %s...\n", name)
		agentValues, valuesErr := temporaryValues(map[string]any{
			"hub":     map[string]any{"url": hubURL, "token": token},
			"cluster": map[string]any{"name": name},
		})
		if valuesErr != nil {
			return valuesErr
		}
		if err := a.runner.Run(ctx, "helm", "upgrade", "--install", "kfleet-agent", agentChart,
			"--version", chartVersion,
			"--kube-context", "kind-"+name,
			"--values", agentValues,
			"--set", "image.repository=ghcr.io/1solomonwakhungu/kfleet/agent",
			"--set", "image.tag="+imageTag,
			"--wait", "--timeout", "120s"); err != nil {
			_ = os.Remove(agentValues)
			return err
		}
		_ = os.Remove(agentValues)
	}
	if err := a.waitForAgents(ctx, *port, *username, password, *clusters); err != nil {
		return err
	}
	keepPortForward = true

	fmt.Fprintf(a.stdout, "\nkfleet is ready at http://localhost:%d\nUsername: %s\nPassword: %s\nCleanup: kfleet cleanup\n", *port, *username, password)
	return nil
}

func (a app) waitForAgents(ctx context.Context, port int, username, password string, expected int) error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	client := *a.http
	client.Jar = jar
	loginBody, err := json.Marshal(map[string]string{"username": username, "password": password})
	if err != nil {
		return err
	}
	loginRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://localhost:%d/api/v1/auth/login", port), bytes.NewReader(loginBody))
	if err != nil {
		return err
	}
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse, err := client.Do(loginRequest)
	if err != nil {
		return fmt.Errorf("log in to the hub: %w", err)
	}
	_ = loginResponse.Body.Close()
	if loginResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("log in to the hub: HTTP %d", loginResponse.StatusCode)
	}

	fmt.Fprintln(a.stdout, "Waiting for agents to register...")
	for attempt := 0; attempt < 60; attempt++ {
		request, requestErr := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://localhost:%d/api/v1/clusters", port), nil)
		if requestErr != nil {
			return requestErr
		}
		response, responseErr := client.Do(request)
		if responseErr == nil {
			var result struct {
				Clusters []json.RawMessage `json:"clusters"`
			}
			decodeErr := json.NewDecoder(response.Body).Decode(&result)
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK && decodeErr == nil && len(result.Clusters) >= expected {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return fmt.Errorf("%d agents did not register within 2 minutes", expected)
}

func (a app) waitForHub(ctx context.Context, port int) error {
	url := fmt.Sprintf("http://localhost:%d/readyz", port)
	for attempt := 0; attempt < 30; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		response, err := a.http.Do(req)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return errors.New("hub did not become ready within 30 seconds")
}

func (a app) cleanup(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("cleanup", flag.ContinueOnError)
	flags.SetOutput(a.stderr)
	configuredClusters := flags.Int("clusters", envInt("KFLEET_CLUSTERS", 0), "number of kind clusters")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	current, _ := a.readState()
	clusters := *configuredClusters
	if clusters == 0 {
		clusters = current.Clusters
	}
	if clusters == 0 {
		clusters = 3
	}

	fmt.Fprintln(a.stdout, "Cleaning up kfleet...")
	stopProcess(current.PortForwardID)
	stopLegacyPortForward()
	for i := 1; i <= clusters; i++ {
		name := fmt.Sprintf("kfleet-%d", i)
		_, _ = a.runner.Output(ctx, "helm", "uninstall", "kfleet-agent", "--kube-context", "kind-"+name)
		_, _ = a.runner.Output(ctx, "helm", "uninstall", "kfleet-hub", "--kube-context", "kind-"+name)
		_, _ = a.runner.Output(ctx, "kind", "delete", "cluster", "--name", name)
	}
	_ = os.Remove(a.statePath)
	_ = os.Remove("/tmp/kfleet-quickstart-cookie.txt")
	fmt.Fprintln(a.stdout, "Cleanup complete.")
	return nil
}

func (a app) status(ctx context.Context, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("status accepts no arguments")
	}
	current, err := a.readState()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(a.stdout, "No kfleet quickstart is recorded. Run 'kfleet quickstart'.")
			return nil
		}
		return err
	}
	url := fmt.Sprintf("http://localhost:%d/readyz", current.Port)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := a.http.Do(req)
	if err != nil {
		fmt.Fprintf(a.stdout, "kfleet is not reachable at http://localhost:%d\n", current.Port)
		return nil
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		fmt.Fprintf(a.stdout, "kfleet returned HTTP %d at http://localhost:%d\n", response.StatusCode, current.Port)
		return nil
	}
	fmt.Fprintf(a.stdout, "kfleet is ready at http://localhost:%d with %d clusters.\n", current.Port, current.Clusters)
	return nil
}

func (a app) openUI(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("open accepts no arguments")
	}
	current, err := a.readState()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errors.New("no kfleet quickstart is recorded; run 'kfleet quickstart'")
		}
		return err
	}
	return a.open(fmt.Sprintf("http://localhost:%d", current.Port))
}

func (a app) readState() (state, error) {
	data, err := os.ReadFile(a.statePath)
	if err != nil {
		return state{}, err
	}
	var result state
	if err := json.Unmarshal(data, &result); err != nil {
		return state{}, fmt.Errorf("read quickstart state: %w", err)
	}
	return result, nil
}

func (a app) writeState(value state) error {
	if err := os.MkdirAll(filepath.Dir(a.statePath), 0o700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(a.statePath, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write quickstart state: %w", err)
	}
	return nil
}

func defaultStatePath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config directory: %w", err)
	}
	return filepath.Join(directory, "kfleet", "quickstart.json"), nil
}

func defaultReleaseVersion() string {
	if value := os.Getenv("KFLEET_VERSION"); value != "" {
		return value
	}
	return version
}

func envString(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func envInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func randomHex(bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func temporaryValues(values map[string]any) (string, error) {
	data, err := json.Marshal(values)
	if err != nil {
		return "", fmt.Errorf("encode Helm values: %w", err)
	}
	file, err := os.CreateTemp("", "kfleet-values-*.json")
	if err != nil {
		return "", fmt.Errorf("create temporary Helm values: %w", err)
	}
	path := file.Name()
	if chmodErr := file.Chmod(0o600); chmodErr != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("protect temporary Helm values: %w", chmodErr)
	}
	if _, writeErr := file.Write(data); writeErr != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("write temporary Helm values: %w", writeErr)
	}
	if closeErr := file.Close(); closeErr != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close temporary Helm values: %w", closeErr)
	}
	return path, nil
}

func stopProcess(pid int) {
	if pid <= 0 {
		return
	}
	output, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=").Output()
	if err != nil || !strings.Contains(string(output), "kubectl") || !strings.Contains(string(output), "port-forward") || !strings.Contains(string(output), "svc/kfleet-hub") {
		return
	}
	process, err := os.FindProcess(pid)
	if err == nil {
		_ = process.Signal(syscall.SIGTERM)
	}
}

func stopLegacyPortForward() {
	const path = "/tmp/kfleet-pf.pid"
	data, err := os.ReadFile(path)
	if err == nil {
		if pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data))); parseErr == nil {
			stopProcess(pid)
		}
	}
	_ = os.Remove(path)
}

func openBrowser(url string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name, args = "open", []string{url}
	case "linux":
		name, args = "xdg-open", []string{url}
	default:
		return fmt.Errorf("opening a browser is not supported on %s; open %s manually", runtime.GOOS, url)
	}
	if err := exec.Command(name, args...).Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
