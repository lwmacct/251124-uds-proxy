package udsproxy

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lwmacct/251124-uds-proxy/internal/proxy"
	"github.com/urfave/cli/v3"
)

// Command returns the uds-proxy CLI command
func Command(version string) *cli.Command {
	return &cli.Command{
		Name:    "uds-proxy",
		Usage:   "HTTP server that proxies requests to Unix domain sockets",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "host",
				Aliases: []string{"H"},
				Value:   "127.0.0.1",
				Usage:   "listen host address",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Value:   0,
				Usage:   "listen port (0 for auto-assign)",
			},
			&cli.StringFlag{
				Name:  "port-file",
				Value: "/tmp/uds-proxy.port",
				Usage: "file to write actual port",
			},
			&cli.IntFlag{
				Name:  "timeout",
				Value: 10000,
				Usage: "request timeout in milliseconds",
			},
			&cli.IntFlag{
				Name:  "max-conns",
				Value: 10,
				Usage: "maximum connections per socket",
			},
			&cli.IntFlag{
				Name:  "max-idle-conns",
				Value: 5,
				Usage: "maximum idle connections per socket",
			},
			&cli.BoolFlag{
				Name:  "no-access-log",
				Value: false,
				Usage: "disable access logging",
			},
		},
		Action: runServer,
	}
}

func runServer(ctx context.Context, cmd *cli.Command) error {
	cfg := proxy.Config{
		Host:         cmd.String("host"),
		Port:         int(cmd.Int("port")),
		PortFile:     cmd.String("port-file"),
		Timeout:      int(cmd.Int("timeout")),
		MaxConns:     int(cmd.Int("max-conns")),
		MaxIdleConns: int(cmd.Int("max-idle-conns")),
		NoAccessLog:  cmd.Bool("no-access-log"),
	}

	server, err := proxy.NewServer(cfg)
	if err != nil {
		return err
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		server.Shutdown()
		os.Exit(0)
	}()

	return server.Run()
}
