package udsproxy

import (
	"github.com/lwmacct/251124-uds-proxy/internal/version"
	"github.com/urfave/cli/v3"
)

// Command returns the uds-proxy CLI command
var Command = &cli.Command{
	Name:   "uds-proxy",
	Usage:  "HTTP server that proxies requests to Unix domain sockets",
	Action: action,
	Commands: []*cli.Command{
		version.Command,
	},
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
}
