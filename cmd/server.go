package cmd

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/domsnail/doctryne/cfg"
	http_handler "github.com/domsnail/doctryne/internal/handler/http"
	"github.com/domsnail/doctryne/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const messageMaxSize int = 1024 * 1024 * 1024 * 128

type Server struct {
	mux *http.ServeMux
	srv *http.Server

	cfg *cfg.ServerConfig
}

type ServerOptions struct {
	Config *cfg.ServerConfig

	InspectionService service.IInspectionService
	DeveloperService  service.IDeveloperService
}

func CreateServer(opts ServerOptions) (*Server, error) {
	var server = Server{
		cfg: opts.Config,
		mux: http.NewServeMux(),
	}

	httpServer := http.NewServeMux()

	if server.cfg == nil || server.cfg.Host == "" || server.cfg.Port == 0 {
		return nil, errors.New("invalid server config: missing host address or port")
	}

	var grpcOpts = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(messageMaxSize),
		grpc.Creds(insecure.NewCredentials()),
	}

	grpcServer := grpc.NewServer(grpcOpts...)

	if !opts.Config.DisableReflect {
		slog.Warn("grpc server reflection enabled")
		reflection.Register(grpcServer)
	}

	if !opts.Config.DisableHealth {
		slog.Warn("grpc server health check enabled")
		grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())
	}

	if !opts.Config.DisableWebUI {
		slog.Warn("server web user interface enabled")

		httpHandler := http_handler.NewHandler(&http_handler.HandlerOptions{
			InspectionService: opts.InspectionService,
			DeveloperService:  opts.DeveloperService,
			Config:            opts.Config,
		})

		httpHandler.HandleMux(httpServer)
	}

	server.mux.Handle("/", grpcHandlerFunc(grpcServer, httpServer))

	return &server, nil
}

func (server *Server) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "starting server...")

	var protocols = http.Protocols{}
	protocols.SetHTTP1(true)
	protocols.SetHTTP2(true)
	protocols.SetUnencryptedHTTP2(true)

	server.srv = &http.Server{
		Addr:      net.JoinHostPort(server.cfg.Host, strconv.Itoa(int(server.cfg.Port))),
		Handler:   defaultSlogMiddleware()(server.mux),
		Protocols: &protocols,
	}

	go func() {
		err := server.srv.ListenAndServe()
		if err != nil {
			slog.ErrorContext(ctx, "failed to start http server", slog.String("error", err.Error()))
			return
		}
	}()

	return nil
}

func (server *Server) GracefulStop(ctx context.Context) error {
	slog.WarnContext(ctx, "gracefully stopping server...")

	return server.srv.Shutdown(ctx)
}

func (server *Server) Stop(ctx context.Context) error {
	slog.WarnContext(ctx, "stopping server...")

	//server.srv.Stop()
	return nil
}

func grpcHandlerFunc(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpHandler.ServeHTTP(w, r)
		}
	})
}

func defaultSlogMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			slog.InfoContext(r.Context(), "http request",
				slog.String("method", r.Method),
				slog.String("proto", r.Proto),
				slog.String("path", r.URL.Path),
				slog.Duration("latency", time.Since(start)),
				slog.String("ip", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)

			next.ServeHTTP(w, r)
		})
	}
}
