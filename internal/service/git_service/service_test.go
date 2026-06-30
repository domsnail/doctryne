package git_service

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"testing"

	"github.com/domsnail/doctryne/cfg"
	"github.com/stretchr/testify/require"
)

func TestFindManifestFiles(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	t.Run("test git repository in-memory bare clone", func(t *testing.T) {
		config := cfg.GitHistoryInspectionConfig{}

		service := NewGitHistoryServiceImpl(config)

		gitUrl, err := url.Parse("https://github.com/Qvineox/cyclonedx-ui")
		require.NoError(t, err)

		repository, err := service.InspectRepository(context.Background(), gitUrl, "main")
		require.NotNil(t, repository)
		require.NoError(t, err)
	})

	t.Run("test git repository bare clone on disk", func(t *testing.T) {
		config := cfg.GitHistoryInspectionConfig{
			SaveToDisk:  true,
			AlwaysFetch: true,
			Filepath:    "./test",
		}

		service := NewGitHistoryServiceImpl(config)

		gitUrl, err := url.Parse("https://github.com/Qvineox/cyclonedx-ui")
		require.NoError(t, err)

		repository, err := service.InspectRepository(context.Background(), gitUrl, "main")
		require.NotNil(t, repository)
		require.NoError(t, err)
	})

	t.Run("test git repository bare clone with tags and all branches on disk", func(t *testing.T) {
		config := cfg.GitHistoryInspectionConfig{
			SaveToDisk: true,
			Filepath:   "./test",
		}

		service := NewGitHistoryServiceImpl(config)

		gitUrl, err := url.Parse("https://github.com/go-git/go-git")
		require.NoError(t, err)

		repository, err := service.InspectRepository(context.Background(), gitUrl, "")
		require.NotNil(t, repository)
		require.NoError(t, err)
	})

	t.Run("test git repository full clone on disk", func(t *testing.T) {
		config := cfg.GitHistoryInspectionConfig{
			SaveToDisk: true,
			FullClone:  true,
			Filepath:   "./test",
		}

		service := NewGitHistoryServiceImpl(config)

		gitUrl, err := url.Parse("https://github.com/Qvineox/cyclonedx-ui")
		require.NoError(t, err)

		repository, err := service.InspectRepository(context.Background(), gitUrl, "main")
		require.NotNil(t, repository)
		require.NoError(t, err)
	})
}
