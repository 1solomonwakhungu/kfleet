// Package version exposes the build version stamped into kfleet binaries.
package version

// Version is the kfleet build version. Release builds override it with
// -ldflags "-X github.com/1solomonwakhungu/kfleet/internal/version.Version=v1.2.3".
var Version = "dev"

// String returns the current build version, never an empty string.
func String() string {
	if Version == "" {
		return "dev"
	}
	return Version
}
