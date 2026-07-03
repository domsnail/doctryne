package github_service

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/domsnail/doctryne/cfg"
	"github.com/stretchr/testify/require"
)

func TestGithubServiceImpl_Ping(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	t.Run("github client prepare test", func(t *testing.T) {
		service := NewGithubServiceImpl(GithubServiceOpts{
			LatestActivityPeriod: 0,
		})

		require.NotNil(t, service)

		service = NewGithubServiceImpl(GithubServiceOpts{
			LatestActivityPeriod: time.Hour,
			AccessToken:          "token",
		})

		require.NotNil(t, service)
	})

	t.Run("github client ping test", func(t *testing.T) {
		service := NewGithubServiceImpl(GithubServiceOpts{})

		require.NotNil(t, service)
		require.Error(t, service.Ping(context.Background()))
	})
}

func TestGithubServiceImpl_GetRepositoryInfo(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	config, err := cfg.NewConfigFromEnv()
	require.NoError(t, err)

	cfg.SetGlobalConfig(config)

	service := NewGithubServiceImpl(GithubServiceOpts{})

	require.NoError(t, err)
	require.NotNil(t, service)

	t.Run("get user repository info", func(t *testing.T) {
		repo, err := service.GetRepositoryByName(context.Background(), "Qvineox", "cyclonedx-ui")
		require.NoError(t, err)
		require.NotNil(t, repo)

		require.EqualValues(t, "qvineox/cyclonedx-ui", repo.Name)

		require.NotNil(t, repo.Owner)
		require.EqualValues(t, "Qvineox", repo.Owner.Username)
		require.NotNil(t, repo.Owner)
		require.False(t, repo.Owner.IsPrivate)

		require.Nil(t, repo.Org)
	})

	t.Run("get organization repository info", func(t *testing.T) {
		repo, err := service.GetRepositoryByName(context.Background(), "domsnail", "doctryne")
		require.NoError(t, err)
		require.NotNil(t, repo)

		require.EqualValues(t, "doctryne", repo.Name)

		require.NotNil(t, repo.Owner)
		require.EqualValues(t, "domsnail", repo.Owner.Username)
		require.NotNil(t, repo.Owner)
		require.False(t, repo.Owner.IsPrivate)

		require.NotNil(t, repo.Org)
		require.EqualValues(t, "domsnail", repo.Org.Name)
		require.EqualValues(t, "domsnail", repo.Org.Login)
	})

	t.Run("get repository by git url", func(t *testing.T) {
		link, _ := url.Parse("https://github.com/facebook/react.git")

		repo, err := service.GetRepositoryByURL(context.Background(), link)
		require.NoError(t, err)
		require.NotNil(t, repo)

		require.EqualValues(t, "react", repo.Name)

		require.NotNil(t, repo)
		require.EqualValues(t, int64(10270250), repo.ID)
		require.EqualValues(t, "https://react.dev", repo.Homepage)

		require.NotNil(t, repo.Owner)
		require.NotNil(t, repo.Org)
		require.EqualValues(t, int64(102812), repo.Owner.ID)
		require.EqualValues(t, "react", repo.Owner.Username)
		require.EqualValues(t, repo.Org.Login, repo.Owner.Username)
	})

	t.Run("get repository contributors by git url", func(t *testing.T) {
		devs, err := service.GetRepositoryContributors(context.Background(), "react", "react")
		require.NoError(t, err)
		require.NotNil(t, devs)

		require.GreaterOrEqual(t, len(devs), 200)

		var uniquenessCheck = make(map[string]bool)
		for _, dev := range devs {
			if _, ok := uniquenessCheck[dev.Username]; ok {
				require.Fail(t, "duplicate repository contributor")
			}
		}
	})
}

func TestGithubServiceImpl_GetUserInfo(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	config, err := cfg.NewConfigFromEnv()
	require.NoError(t, err)

	cfg.SetGlobalConfig(config)

	service := NewGithubServiceImpl(GithubServiceOpts{})
	require.NotNil(t, service)

	t.Run("get user info", func(t *testing.T) {
		user, err := service.GetUserByUsername(context.Background(), "Qvineox")
		require.NoError(t, err)
		require.NotNil(t, user)

		require.EqualValues(t, int64(43321560), user.ID)
		require.EqualValues(t, "qvineox", user.Username)
		require.EqualValues(t, "Lysak Yaroslav", user.Fullname)
		require.EqualValues(t, "Moscow, Russia", user.Location)

		require.GreaterOrEqual(t, user.PublicReposCount, uint64(30))

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

	config, err := cfg.NewConfigFromEnv()
	require.NoError(t, err)

	cfg.SetGlobalConfig(config)

	service := NewGithubServiceImpl(GithubServiceOpts{})
	require.NotNil(t, service)

	t.Run("get organization info", func(t *testing.T) {
		org, err := service.GetOrganizationByName(context.Background(), "domsnail")
		require.NoError(t, err)
		require.NotNil(t, org)

		require.EqualValues(t, int64(227165210), org.ID)
		require.EqualValues(t, "domsnail", org.Login)
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

		require.EqualValues(t, int64(168166), org.ID)
		require.EqualValues(t, "atlassian", org.Login)
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

	config, err := cfg.NewConfigFromEnv()
	require.NoError(t, err)

	cfg.SetGlobalConfig(config)

	service := NewGithubServiceImpl(GithubServiceOpts{})
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

	config, err := cfg.NewConfigFromEnv()
	require.NoError(t, err)

	cfg.SetGlobalConfig(config)

	service := NewGithubServiceImpl(GithubServiceOpts{})
	require.NotNil(t, service)

	t.Run("get users with company rgs in bio", func(t *testing.T) {
		_, err := service.GetCompanyUsers(context.Background(), "PJSC IC RGS")
		require.NoError(t, err)
	})
}
