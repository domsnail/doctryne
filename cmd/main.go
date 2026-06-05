package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/domsnail/doctryne/cfg"
)

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	config, err := cfg.NewConfigFromFlags(rootCtx)
	if err != nil {
		panic(err.Error())
	}

	var handler slog.Handler
	switch config.Logging.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: config.Logging.AddSource,
			Level:     slog.Level(config.Logging.Level),
		})
	case "text":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: config.Logging.AddSource,
			Level:     slog.Level(config.Logging.Level),
		})
	default:
		panic(fmt.Sprintf("invalid logging format: '%s'", config.Logging.Format))
	}

	slog.SetDefault(slog.New(handler))

	slog.DebugContext(rootCtx, "loaded configuration variables",
		slog.String("config_file_path", config.FilePath),
		slog.String("http_proxy", config.HttpProxy.Redacted()),
		slog.Duration("timeout", config.Timeout),
		slog.Group("output",
			slog.String("format", string(config.Output.Format)),
		),
		slog.Group("server",
			slog.Bool("enabled", config.Server.Enabled),
			slog.String("host", config.Server.Host),
			slog.Int("port", int(config.Server.Port)),
			slog.Bool("require_access_key", config.Server.AccessKey != ""),
		),
		slog.Group("logging",
			slog.Int("level", config.Logging.Level),
			slog.String("format", config.Logging.Format),
			slog.Bool("add_source", config.Logging.AddSource),
		),
	)

	slog.DebugContext(rootCtx, "setting global config...")
	cfg.SetGlobalConfig(config)

	if config.Server.Enabled {
		err := RunServer(rootCtx, config.Server)
		if err != nil {
			return
		}

		<-rootCtx.Done()
	}

	if !config.HasScan() {
		var err error
		config.Scan, err = cfg.NewScanFromArgs(rootCtx)

		if err != nil {
			slog.ErrorContext(rootCtx, fmt.Sprintf("failed to determine scan target(s): %s", err.Error()))
			os.Exit(1)
		}
	}

	err = RunCLI(rootCtx)
	if err != nil {
		return
	}
}
