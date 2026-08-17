package api

// Pod log streaming uses a reverse channel: the agent holds an outbound
// WebSocket to the hub, and the hub pushes log stream requests down it. The
// agent answers with data frames that the hub relays to the browser as SSE.
// Both directions use LogStreamMessage so the wire format stays symmetric.
const (
	// LogStreamStart is sent hub -> agent to open a pod log stream.
	LogStreamStart = "start"
	// LogStreamStop is sent hub -> agent when the browser disconnects.
	LogStreamStop = "stop"
	// LogStreamData is sent agent -> hub for each log line.
	LogStreamData = "data"
	// LogStreamEnd is sent agent -> hub when a stream finishes or fails.
	LogStreamEnd = "end"
)

// LogStreamMessage is a single frame on the agent reverse channel.
type LogStreamMessage struct {
	Type      string `json:"type"`
	StreamID  string `json:"streamId"`
	Namespace string `json:"namespace,omitempty"`
	Pod       string `json:"pod,omitempty"`
	Container string `json:"container,omitempty"`
	Follow    bool   `json:"follow,omitempty"`
	TailLines int64  `json:"tailLines,omitempty"`
	Line      string `json:"line,omitempty"`
	Error     string `json:"error,omitempty"`
}
