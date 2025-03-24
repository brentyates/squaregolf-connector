package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of the application
	Version = "0.1.0-alpha.7"

	// BuildTime is the time the binary was built
	BuildTime = "unknown"

	// GitCommit is the git commit hash
	GitCommit = "unknown"

	// GoVersion is the version of Go used to build the binary
	GoVersion = runtime.Version()
)

// GetVersion returns the full version string including build information
func GetVersion() string {
	return fmt.Sprintf("%s (build: %s, commit: %s, go: %s)",
		Version,
		BuildTime,
		GitCommit,
		GoVersion,
	)
}

// GetShortVersion returns just the version number
func GetShortVersion() string {
	return Version
}
