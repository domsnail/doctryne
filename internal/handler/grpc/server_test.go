package grpc

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	inspection_v1 "github.com/domsnail/doctryne/api/gen/go/inspection/v1"
	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/service/github_service"
	"github.com/domsnail/doctryne/internal/service/inspect_service"
	"github.com/domsnail/doctryne/internal/service/manifest_service"
	"github.com/domsnail/doctryne/internal/service/registry_service"
	"github.com/stretchr/testify/require"
)

const testDataDir = "../../../test"

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

	handler := NewInspectionGRPCHandler(inspect_service.NewInspectionService(
		manifest_service.NewManifestServiceImpl(),
		github_service.NewGithubServiceImpl(github_service.GithubServiceOpts{}),
		registry_service.NewRegistryServiceImpl(registry_service.RegistryServiceOpts{}),
	))

	t.Run("test package.json inspection with errors", func(t *testing.T) {
		contents, err := os.ReadFile(filepath.Join(testDataDir, "package.brobot.json"))
		require.NoError(t, err)
		require.NotEmpty(t, contents)

		opts := inspection_v1.InspectionOptions_builder{
			ScanType:        new(inspection_v1.ScanType_SCAN_TYPE_BINARY),
			Mode:            new(inspection_v1.InspectionMode_INSPECTION_MODE_DIRECT),
			ManifestType:    new("package.json"),
			LoadUserProfile: new(false),
		}.Build()

		inspection, err := handler.Inspect(context.Background(), opts)
		require.ErrorContains(t, err, "manifest file is not specified")
		require.Nil(t, inspection)

		opts = inspection_v1.InspectionOptions_builder{
			ScanType:        new(inspection_v1.ScanType_SCAN_TYPE_BINARY),
			Mode:            new(inspection_v1.InspectionMode_INSPECTION_MODE_DIRECT),
			Manifest:        contents,
			LoadUserProfile: new(false),
		}.Build()

		inspection, err = handler.Inspect(context.Background(), opts)
		require.ErrorContains(t, err, "manifest file type is not specified")
		require.Nil(t, inspection)

		opts = inspection_v1.InspectionOptions_builder{
			ScanType:        new(inspection_v1.ScanType_SCAN_TYPE_DIRPATH),
			Mode:            new(inspection_v1.InspectionMode_INSPECTION_MODE_DIRECT),
			ManifestType:    new("package.json"),
			Manifest:        []byte(testDataDir),
			LoadUserProfile: new(false),
		}.Build()

		inspection, err = handler.Inspect(context.Background(), opts)
		require.ErrorContains(t, err, "scan type 'SCAN_TYPE_DIRPATH' not supported in server mode")
		require.Nil(t, inspection)
	})

	t.Run("test package.brobot.json handling", func(t *testing.T) {
		contents, err := os.ReadFile(filepath.Join(testDataDir, "package.brobot.json"))
		require.NoError(t, err)
		require.NotEmpty(t, contents)

		opts := inspection_v1.InspectionOptions_builder{
			ScanType:        new(inspection_v1.ScanType_SCAN_TYPE_BINARY),
			Mode:            new(inspection_v1.InspectionMode_INSPECTION_MODE_DIRECT),
			ManifestType:    new("package.json"),
			Manifest:        contents,
			LoadUserProfile: new(false),
		}.Build()

		inspection, err := handler.Inspect(context.Background(), opts)
		require.NoError(t, err)
		require.NotNil(t, inspection)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		inspection, err = handler.Inspect(ctx, opts)
		require.Error(t, err)
		require.Nil(t, inspection)
	})

	t.Run("test package.cyclonedx_ui.json handling", func(t *testing.T) {
		contents, err := os.ReadFile(filepath.Join(testDataDir, "package.cyclonedx_ui.json"))
		require.NoError(t, err)
		require.NotEmpty(t, contents)

		opts := inspection_v1.InspectionOptions_builder{
			ScanType:        new(inspection_v1.ScanType_SCAN_TYPE_BINARY),
			Mode:            new(inspection_v1.InspectionMode_INSPECTION_MODE_DIRECT),
			ManifestType:    new("package.json"),
			Manifest:        contents,
			LoadUserProfile: new(false),
		}.Build()

		inspection, err := handler.Inspect(context.Background(), opts)
		require.NoError(t, err)
		require.NotNil(t, inspection)
	})

	t.Run("test package.cyclonedx_ui.json handling by url", func(t *testing.T) {
		const rawURL = "https://raw.githubusercontent.com/Qvineox/cyclonedx-ui/refs/heads/main/frontend/package.json"

		opts := inspection_v1.InspectionOptions_builder{
			ScanType:        new(inspection_v1.ScanType_SCAN_TYPE_URL),
			Mode:            new(inspection_v1.InspectionMode_INSPECTION_MODE_DIRECT),
			ManifestType:    new("package.json"),
			Manifest:        []byte(rawURL),
			LoadUserProfile: new(false),
		}.Build()

		inspection, err := handler.Inspect(context.Background(), opts)
		require.NoError(t, err)
		require.NotNil(t, inspection)
	})
}
