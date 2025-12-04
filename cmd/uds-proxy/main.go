package main

import (
	"context"
	"fmt"
	"os"

	"github.com/lwmacct/251124-uds-proxy/internal/command/udsproxy"
)

func main() {
	if err := udsproxy.Command.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
