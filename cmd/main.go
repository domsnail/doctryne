package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/domsnail/doctryne/cfg"
)

func main() {
	rootCtx := context.Background()

	config, _ := cfg.NewConfigFromFlags(rootCtx)

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
		slog.ErrorContext(rootCtx, fmt.Sprintf("invalid logging format: '%s'", config.Logging.Format))
		os.Exit(1)
	}

	slog.SetDefault(slog.New(handler))

	slog.DebugContext(rootCtx, "loaded configuration variables",
		slog.String("config_file_path", config.FilePath),
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
}
