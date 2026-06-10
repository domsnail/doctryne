package cli

import (
	"log/slog"
	"os"
	"testing"

	"github.com/domsnail/doctryne/internal/service/inspect_service"
	"github.com/domsnail/doctryne/internal/service/manifest_service"
)

func TestNewInspectionHandler(t *testing.T) {

}

func TestInspection(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	handler := NewInspectionCommandLineHandler(
		inspect_service.NewInspectionService(
			manifest_service.NewManifestServiceImpl(),
			nil,
			nil,
		),
	)

	t.Run("test with single file", func(t *testing.T) {

	})
}
