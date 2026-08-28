package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Release identity, injected by the release build via
// -ldflags "-X github.com/joshpeak/screenz/internal/cli.version=v1.2.3 …"
// using the goreleaser variable naming convention (G6).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "source"
)

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "screenz %s (commit %s, built %s by %s)\n", version, commit, date, builtBy)
}

// osVersionParts parses "26.6.2" into major and minor; ok is false when the
// string is not a version.
func osVersionParts(v string) (major, minor int, ok bool) {
	parts := strings.Split(v, ".")
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}
	if len(parts) > 1 {
		if minor, err = strconv.Atoi(parts[1]); err != nil {
			return 0, 0, false
		}
	}
	return major, minor, true
}
