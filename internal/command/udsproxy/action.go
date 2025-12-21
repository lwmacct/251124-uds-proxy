package udsproxy

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/lwmacct/251124-uds-proxy/internal/config"
	"github.com/lwmacct/251124-uds-proxy/internal/proxy"
	"github.com/lwmacct/251207-go-pkg-cfgm/pkg/cfgm"
	"github.com/lwmacct/251207-go-pkg-version/pkg/version"
	"github.com/lwmacct/251219-go-pkg-logm/pkg/logm"
	"github.com/urfave/cli/v3"
)

// 配置优先级 (从低到高)：
// 1. 默认值 (config.defaultConfig)
// 2. 配置文件 (config.yaml)
// 3. 环境变量 (UDS_PROXY_*)
// 4. CLI flags (用户明确指定)

func action(ctx context.Context, cmd *cli.Command) error {
	logm.MustInit(logm.PresetAuto()...)

	cfg := cfgm.MustLoadCmd(cmd, config.DefaultConfig(), version.GetAppRawName())

	// 启动服务器
	server, err := proxy.NewServer(cfg)
	if err != nil {
		return err
	}

	// 优雅关闭
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
