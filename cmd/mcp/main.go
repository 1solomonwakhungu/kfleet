// Command mcp runs the kfleet MCP server over standard input and output.
package main

import (
	"fmt"
	"os"

	kfleetmcp "github.com/1solomonwakhungu/kfleet/internal/mcp"
)

func main() {
	if err := kfleetmcp.RunStdio(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "kfleet MCP server: %v\n", err)
		os.Exit(1)
	}
}
