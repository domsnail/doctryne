package cli

import (
	"log/slog"
	"os"
	"testing"
)

func TestNewInspectionHandler(t *testing.T) {

}

func TestInspection(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	t.Run("test with single file", func(t *testing.T) {

	})
}
