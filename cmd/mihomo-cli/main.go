package main

import (
	"fmt"
	"os"

	"mihomo-cli/internal/cli"
)

func main() {
	// Skip heavy init for help/version requests
	skipInit := false
	for _, a := range os.Args[1:] {
		if a == "--help" || a == "-h" || a == "help" || a == "--version" || a == "version" {
			skipInit = true
			break
		}
	}
	if !skipInit {
		if _, _, _, _, err := cli.InitManagers(); err != nil {
			fmt.Fprintf(os.Stderr, "init: %v\n", err)
		}
	}
	cli.Execute()
}
