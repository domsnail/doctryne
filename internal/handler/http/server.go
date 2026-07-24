package http

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/service"
)

type Handler struct {
	service service.IInspectionService
	config  cfg.Server
}

const pathPrefix = "/api/v1"

func NewHandler(service service.IInspectionService, config cfg.Server) *Handler {
	if service == nil {
		panic("inspection service is nil")
	}

	return &Handler{service: service, config: config}
}

func (h *Handler) RunServer(ctx context.Context) {
	var mux = http.NewServeMux()
	mux.HandleFunc("/upload", h.handleManifestUpload)
	mux.HandleFunc("/", h.static())

	srv := http.Server{
		Addr:    net.JoinHostPort(h.config.Host, strconv.Itoa(int(h.config.Port))),
		Handler: defaultSlogMiddleware()(mux),
	}

	go func() {
		slog.InfoContext(ctx, "starting http server...",
			slog.String("host", net.JoinHostPort(h.config.Host, strconv.Itoa(int(h.config.Port)))),
		)

		err := srv.ListenAndServe()
		if err != nil {
			return
		}
	}()

	<-ctx.Done()
	slog.WarnContext(ctx, "shutting down server...", slog.String("error", ctx.Err().Error()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to shutdown server", slog.String("error", err.Error()))
		return
	}

	slog.WarnContext(ctx, "server shutdown successfully complete")
	return
}
