package cli

// Version is the current version of the app
const Version = "0.0.1"

// GitSHA is the git commit SHA populated during build
// - We want to replace this variable at build time with "-ldflags -X cli.GitSHA=xxx", where const is not supported.
var GitSHA = ""
