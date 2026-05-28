package main

import (
	"fmt"
	"os"

	"mihomo-cli/internal/cli"
)

func main() {
	if _, _, _, _, err := cli.InitManagers(); err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
	}
	cli.Execute()
}
