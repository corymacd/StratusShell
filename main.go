package main

import "github.com/corymacd/StratusShell/cmd"

// Version information populated via ldflags during build by GoReleaser:
//   -ldflags "-X main.version=1.0.0 -X main.commit=abc123 -X main.date=2024-01-01"
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	cmd.Execute()
}
