package main

import (
	"context"
	"fmt"
	"os"

	udsproxy "github.com/lwmacct/251124-uds-proxy/internal/commands/uds-proxy"
)

var version = "0.1.0"

func main() {
	cmd := udsproxy.Command(version)
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
