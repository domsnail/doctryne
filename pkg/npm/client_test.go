package npm

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_GetPackage(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	t.Run("npm client prepare test", func(t *testing.T) {
		client, err := NewClient(Options{
			Timeout: 0,
		})

		require.Error(t, err)
		require.Nil(t, client)

		client, err = NewClient(Options{
			Timeout: time.Second * 30,
		})

		require.NoError(t, err)
		require.NotNil(t, client)
	})

	var testPackage *Package

	t.Run("npm client get package test", func(t *testing.T) {
		client, err := NewClient(Options{
			Timeout: time.Second * 30,
		})

		require.NoError(t, err)
		require.NotNil(t, client)

		err = client.Ping(context.Background())
		require.Error(t, err, "must be unauthorized")

		testPackage, err = client.GetPackage(context.Background(), "detect-libc", "")
		require.NoError(t, err)
		require.NotNil(t, testPackage)

		pkg, err := client.GetPackage(context.Background(), "detect-libc", "2.0.1")
		require.NoError(t, err)
		require.NotNil(t, pkg)

		pkg, err = client.GetPackage(context.Background(), "package-that-not-exist", "9999")
		require.NoError(t, err)
		require.Nil(t, pkg)
	})

	t.Run("npm client get package test with bearer token", func(t *testing.T) {
		client, err := NewClient(Options{
			Timeout:     time.Second * 30,
			BearerToken: os.Getenv("NPM_API_KEY"),
		})
		require.NoError(t, err, "must be authorized")

		err = client.Ping(context.Background())
		require.NoError(t, err)
	})

	t.Run("npm client get package emails", func(t *testing.T) {
		emails := testPackage.GetUniqueEmails()
		require.Len(t, emails, 4)
	})

	t.Run("npm client get package urls", func(t *testing.T) {
		urls := testPackage.GetUniqueURLs()
		require.Len(t, urls, 26)
	})

	t.Run("npm client get creation date", func(t *testing.T) {
		at, err := testPackage.GetCreatedAt()
		require.NoError(t, err)
		require.EqualValues(t, "2017-07-03 20:58:20.378 +0000 UTC", at.String())
	})

	t.Run("npm client get modification date", func(t *testing.T) {
		at, err := testPackage.GetModifiedAt()
		require.NoError(t, err)
		require.EqualValues(t, "2025-10-05 12:46:33.558 +0000 UTC", at.String())
	})

	t.Run("npm client get package stats", func(t *testing.T) {
		client, err := NewClient(Options{
			Timeout: time.Second * 30,
		})

		require.NoError(t, err)
		require.NotNil(t, client)

		stats, err := client.GetPackageStats(context.Background(), "", "")
		require.Error(t, err)
		require.Nil(t, stats)

		stats, err = client.GetPackageStats(context.Background(), "detect-libc", "last-month")
		require.NoError(t, err)
		require.NotNil(t, stats)
	})
}
