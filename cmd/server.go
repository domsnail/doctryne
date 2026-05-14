package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/domsnail/doctryne/cfg"
)

func RunServer(ctx context.Context, config cfg.Server) error {
	slog.InfoContext(ctx, "starting server...")

	return errors.New("not implemented")
}
