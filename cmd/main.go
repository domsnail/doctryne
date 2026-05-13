package main

import (
	"context"
	"log/slog"

	"github.com/domsnail/doctryne/cfg"
)

func main() {
	rootCtx := context.Background()

	config, _ := cfg.NewConfigFromFlags()

	slog.DebugContext(rootCtx, "loaded configuration variables",
		slog.Any("config", config),
	)
}
