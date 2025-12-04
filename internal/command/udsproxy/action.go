package udsproxy

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/lwmacct/251124-uds-proxy/internal/proxy"
	"github.com/lwmacct/251125-go-mod-logger/pkg/logger"
	"github.com/urfave/cli/v3"
)

func action(ctx context.Context, cmd *cli.Command) error {
	if err := logger.InitEnv(); err != nil {
		slog.Warn("初始化日志系统失败，使用默认配置", "error", err)
	}

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
		sig := <-sigChan
		slog.Info("收到关闭信号", "signal", sig.String())
		server.Shutdown()
		os.Exit(0)
	}()

	return server.Run()
}
