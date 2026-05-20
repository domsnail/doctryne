package service

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGithubServiceImpl_Ping(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	t.Run("github client prepare test", func(t *testing.T) {
		service, err := NewGithubServiceImpl(GithubServiceOpts{
			Timeout: 0,
		})

		require.Error(t, err)
		require.Nil(t, service)

		service, err = NewGithubServiceImpl(GithubServiceOpts{
			Timeout: time.Second * 30,
		})

		require.NoError(t, err)
		require.NotNil(t, service)
	})

	t.Run("github client ping test", func(t *testing.T) {
		service, err := NewGithubServiceImpl(GithubServiceOpts{
			Timeout: 0,
		})

		require.Error(t, err)
		require.Nil(t, service)

		service, err = NewGithubServiceImpl(GithubServiceOpts{
			Timeout: time.Second * 30,
		})

		require.NoError(t, err)
		require.NotNil(t, service)
		require.Error(t, service.Ping(context.Background()))
	})

	t.Run("github client authorized ping test", func(t *testing.T) {
		require.NotEmpty(t, os.Getenv("GITHUB_API_KEY"))

		service, err := NewGithubServiceImpl(GithubServiceOpts{
			Timeout:     time.Second * 30,
			AccessToken: os.Getenv("GITHUB_API_KEY"),
		})

		require.NoError(t, err)
		require.NotNil(t, service)
		require.NoError(t, service.Ping(context.Background()))
	})
}

func TestGithubServiceImpl_GetRepositoryInfo(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	service, err := NewGithubServiceImpl(GithubServiceOpts{
		Timeout:     time.Second * 30,
		AccessToken: os.Getenv("GITHUB_API_KEY"),
	})

	require.NoError(t, err)
	require.NotNil(t, service)

	t.Run("get user repository info", func(t *testing.T) {
		repo, resp, err := service.c.Repositories.Get(context.Background(), "Qvineox", "cyclonedx-ui")
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotNil(t, repo)
	})

	t.Run("get organization repository info", func(t *testing.T) {
		repo, resp, err := service.c.Repositories.Get(context.Background(), "domsnail", "doctryne")
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotNil(t, repo)
	})

	t.Run("get repository by git url", func(t *testing.T) {
		repo, resp, err := service.c.Repositories.Get(context.Background(), "domsnail", "doctryne")
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotNil(t, repo)
	})
}
