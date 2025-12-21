package udsproxy

import (
	"github.com/lwmacct/251124-uds-proxy/internal/config"
	"github.com/lwmacct/251207-go-pkg-version/pkg/version"
	"github.com/urfave/cli/v3"
)

// 默认配置 - 单一来源 (Single Source of Truth)
var defaults = config.DefaultConfig()

// Command returns the uds-proxy CLI command
var Command = &cli.Command{
	Name:     "uds-proxy",
	Usage:    "HTTP server that proxies requests to Unix domain sockets",
	Action:   action,
	Commands: []*cli.Command{version.Command},
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "host",
			Aliases: []string{"H"},
			Value:   defaults.Host,
			Usage:   "listen host address",
		},
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Value:   defaults.Port,
			Usage:   "listen port (0 for auto-assign)",
		},
		&cli.StringFlag{
			Name:  "port-file",
			Value: defaults.PortFile,
			Usage: "file to write actual port",
		},
		&cli.IntFlag{
			Name:  "timeout",
			Value: defaults.Timeout,
			Usage: "request timeout in milliseconds",
		},
		&cli.IntFlag{
			Name:  "max-conns",
			Value: defaults.MaxConns,
			Usage: "maximum connections per socket",
		},
		&cli.IntFlag{
			Name:  "max-idle-conns",
			Value: defaults.MaxIdleConns,
			Usage: "maximum idle connections per socket",
		},
		&cli.BoolFlag{
			Name:  "no-access-log",
			Value: defaults.NoAccessLog,
			Usage: "disable access logging",
		},
	},
}
