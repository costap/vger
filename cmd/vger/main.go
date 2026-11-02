package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/costap/vger/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetVersionInfo(version, commit, date)

	// Load .env if present; silently ignore if the file does not exist.
	_ = godotenv.Load()

	if err := cli.Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
