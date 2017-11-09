package cli

import (
	"fmt"
	"runtime"
)

// Version is the current version of the app
const Version = "0.0.2"

// GitSHA is the git commit SHA populated during build
// - We want to replace this variable at build time with "-ldflags -X cli.GitSHA=xxx", where const is not supported.
var GitSHA = ""

func (cli *CLI) version() {
	fmt.Printf("blockchain %s (Git SHA: %s, Go Version: %s)\n", Version, GitSHA, runtime.Version())
}
