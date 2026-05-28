package main

import (
	"fmt"
	"os"

	"mihomo-cli/internal/cli"
)

func main() {
	if err := cli.InitBase(); err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}
	cli.Execute()
}
