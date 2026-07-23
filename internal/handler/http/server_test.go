package http

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/service/github_service"
	"github.com/domsnail/doctryne/internal/service/inspect_service"
	"github.com/domsnail/doctryne/internal/service/manifest_service"
	"github.com/domsnail/doctryne/internal/service/registry_service"
	"github.com/domsnail/doctryne/pkg/stack_exchange"
	"github.com/stretchr/testify/require"
)

func Test_InspectionHandler_JavaScript(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.Level(-8), // trace cached http
	})))

	config, err := cfg.NewConfigFromEnv()
	require.NoError(t, err)

	config.Insecure = true
	config.Languages.JavaScript.CheckDevDependencies = true
	config.Languages.JavaScript.CheckOptionalDependencies = true
	cfg.SetGlobalConfig(config)

	handler := NewInspectionHTTPHandler(inspect_service.NewInspectionService(
		manifest_service.NewManifestServiceImpl(),
		github_service.NewGithubServiceImpl(github_service.GithubServiceOpts{}),
		stack_exchange.NewClient(stack_exchange.Options{}),
		registry_service.NewRegistryServiceImpl(registry_service.RegistryServiceOpts{}),
	), config.Server)

	t.Run("start test http server", func(t *testing.T) {
		handler.RunServer(context.Background())
	})
}
