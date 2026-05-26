package github_service

import (
	"context"
	"log/slog"
	"net/url"
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
		repo, err := service.GetRepositoryByName(context.Background(), "Qvineox", "cyclonedx-ui")
		require.NoError(t, err)
		require.NotNil(t, repo)

		require.EqualValues(t, "qvineox/cyclonedx-ui", repo.Name)

		require.NotNil(t, repo.Owner)
		require.EqualValues(t, "Qvineox", repo.Owner.Username)
		require.False(t, repo.Owner.IsPrivate)

		require.Nil(t, repo.Organization)
	})

	t.Run("get organization repository info", func(t *testing.T) {
		repo, err := service.GetRepositoryByName(context.Background(), "domsnail", "doctryne")
		require.NoError(t, err)
		require.NotNil(t, repo)

		require.EqualValues(t, "domsnail/doctryne", repo.Name)

		require.NotNil(t, repo.Owner)
		require.EqualValues(t, "domsnail", repo.Owner.Username)
		require.False(t, repo.Owner.IsPrivate)

		require.NotNil(t, repo.Organization)
		require.EqualValues(t, "domsnail", repo.Organization.Username)
	})

	t.Run("get repository by git url", func(t *testing.T) {
		link, _ := url.Parse("https://github.com/facebook/react.git")

		repo, err := service.GetRepositoryByURL(context.Background(), link)
		require.NoError(t, err)
		require.NotNil(t, repo)
	})
}

func TestGithubServiceImpl_GetUserInfo(t *testing.T) {
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

	t.Run("get user info", func(t *testing.T) {
		user, err := service.GetUserByUsername(context.Background(), "Qvineox")
		require.NoError(t, err)
		require.NotNil(t, user)

		require.EqualValues(t, 43321560, *user.GithubID)
		require.EqualValues(t, "qvineox", user.Username)
		require.EqualValues(t, "lysak yaroslav", user.Name)
		require.EqualValues(t, "Moscow, Russia", user.Location)

		require.GreaterOrEqual(t, 30, user.PublicReposCount)
		require.GreaterOrEqual(t, 30, user.PublicReposCount)

		require.False(t, user.IsPrivate)
		require.False(t, user.IsHireable)
		require.False(t, user.IsSiteAdmin)

		require.NotNil(t, user.CreatedAt)
		require.NotNil(t, user.UpdatedAt)
		require.Nil(t, user.SuspendedAt)
	})
}

func TestGithubServiceImpl_GetOrganizationInfo(t *testing.T) {
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

	t.Run("get organization info", func(t *testing.T) {
		org, err := service.GetOrganizationByName(context.Background(), "domsnail")
		require.NoError(t, err)
		require.NotNil(t, org)

		require.EqualValues(t, 227165210, *org.GithubID)
		require.EqualValues(t, "domsnail", org.Username)
		require.EqualValues(t, "Domsnail", org.Name)
		require.EqualValues(t, "Russian Federation", org.Location)

		require.GreaterOrEqual(t, uint64(2), org.PublicReposCount)
		require.GreaterOrEqual(t, uint64(1), org.FollowersCount)

		require.False(t, org.IsVerified)

		require.NotNil(t, org.CreatedAt)
		require.NotNil(t, org.UpdatedAt)
		require.Nil(t, org.ArchivedAt)
	})

	t.Run("get organization info", func(t *testing.T) {
		org, err := service.GetOrganizationByName(context.Background(), "atlassian")
		require.NoError(t, err)
		require.NotNil(t, org)

		require.EqualValues(t, 168166, *org.GithubID)
		require.EqualValues(t, "atlassian", org.Username)
		require.EqualValues(t, "Atlassian", org.Name)
		require.EqualValues(t, "Australia", org.Location)

		require.Len(t, org.Emails, 1)

		require.GreaterOrEqual(t, uint64(1600), org.FollowersCount)

		require.True(t, org.IsVerified)

		require.NotNil(t, org.CreatedAt)
		require.NotNil(t, org.UpdatedAt)
		require.Nil(t, org.ArchivedAt)
	})
}

func TestGithubServiceImpl_GetUserActivity(t *testing.T) {
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

	t.Run("get user latest activity info", func(t *testing.T) {
		_, err := service.GetUserActivity(context.Background(), "qvineox")
		require.NoError(t, err)

	})
}

func TestGithubServiceImpl_GetCompanyUsers(t *testing.T) {
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

	t.Run("get users with company rgs in bio", func(t *testing.T) {
		_, err := service.GetCompanyUsers(context.Background(), "PJSC IC RGS")
		require.NoError(t, err)
	})
}
