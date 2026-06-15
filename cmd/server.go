package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"github.com/domsnail/doctryne/cfg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const messageMaxSize int = 1024 * 1024 * 1024 * 128

type Server struct {
	srv *grpc.Server
	cfg *cfg.Server
}

func CreateServer(ctx context.Context, config *cfg.Server) (*Server, error) {
	var server = Server{
		cfg: config,
	}

	if server.cfg == nil || server.cfg.Host == "" || server.cfg.Port == 0 {
		return nil, errors.New("invalid server config: missing host address or port")
	}

	var opts = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(messageMaxSize),
		grpc.Creds(insecure.NewCredentials()),
	}

	server.srv = grpc.NewServer(opts...)

	if !config.DisableReflect {
		slog.WarnContext(ctx, "grpc server reflection enabled")
		reflection.Register(server.srv)
	}

	if !config.DisableReflect {
		slog.WarnContext(ctx, "grpc server reflection enabled")
	}

	if !config.DisableHealth {
		slog.Warn("grpc server health check enabled")
		grpc_health_v1.RegisterHealthServer(server.srv, health.NewServer())
	}

	return &server, nil
}

func (server *Server) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "starting server...")

	listener, err := net.Listen("tcp", net.JoinHostPort(server.cfg.Host, strconv.Itoa(int(server.cfg.Port))))
	if err != nil {
		return fmt.Errorf("failed to start tcp listener: %w", err)
	}

	go func() {
		err = server.srv.Serve(listener)
	}()

	return err
}

func (server *Server) GracefulStop(ctx context.Context) error {
	slog.WarnContext(ctx, "gracefully stopping grpc server...")

	server.srv.GracefulStop()
	return nil
}

func (server *Server) Stop(ctx context.Context) error {
	slog.WarnContext(ctx, "stopping grpc server...")

	server.srv.Stop()
	return nil
}
