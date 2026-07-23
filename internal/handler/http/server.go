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

type InspectionHandler struct {
	service service.IInspectionService
	config  cfg.Server
}

func NewInspectionHTTPHandler(service service.IInspectionService, config cfg.Server) *InspectionHandler {
	if service == nil {
		panic("inspection service is nil")
	}

	return &InspectionHandler{service: service, config: config}
}

func (handler *InspectionHandler) RunServer(ctx context.Context) {
	var mux = http.NewServeMux()
	mux.HandleFunc("/upload", handler.handleManifestUpload)

	srv := http.Server{
		Addr:    net.JoinHostPort(handler.config.Host, strconv.Itoa(int(handler.config.Port))),
		Handler: defaultSlogMiddleware()(mux),
	}

	go func() {
		slog.InfoContext(ctx, "starting http server...",
			slog.String("host", net.JoinHostPort(handler.config.Host, strconv.Itoa(int(handler.config.Port)))),
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
