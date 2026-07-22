package npm

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_GetPackage(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug - 4,
	})))

	t.Run("npm client prepare test", func(t *testing.T) {
		client, err := NewClient(Options{})

		require.Error(t, err)
		require.Nil(t, client)
	})

	var (
		testPackage1 *Package
		testPackage2 *Package
	)

	t.Run("npm client get package test", func(t *testing.T) {
		client, err := NewClient(Options{})

		require.NoError(t, err)
		require.NotNil(t, client)

		err = client.Ping(context.Background())
		require.Error(t, err, "must be unauthorized")

		var raw json.RawMessage
		testPackage1, raw, err = client.GetPackage(context.Background(), "detect-libc")
		require.NoError(t, err)
		require.NotNil(t, raw)
		require.NotNil(t, testPackage1)

		testPackage2, raw, err = client.GetPackage(context.Background(), "react")
		require.NoError(t, err)
		require.NotNil(t, raw)
		require.NotNil(t, testPackage2)

		pkg, raw, err := client.GetPackage(context.Background(), "detect-libc")
		require.NoError(t, err)
		require.NotNil(t, raw)
		require.NotNil(t, pkg)

		pkg, raw, err = client.GetPackage(context.Background(), "package-that-not-exist")
		require.Error(t, err)
		require.Nil(t, raw)
		require.Nil(t, pkg)
	})

	t.Run("npm client get package test with bearer token", func(t *testing.T) {
		require.NotEmpty(t, os.Getenv("NPM_API_KEY"))

		client, err := NewClient(Options{
			BearerToken: os.Getenv("NPM_API_KEY"),
		})
		require.NoError(t, err, "must be authorized")

		err = client.Ping(context.Background())
		require.NoError(t, err)
	})

	t.Run("npm client get package emails", func(t *testing.T) {
		emails := testPackage1.GetUniqueEmails()
		require.Len(t, emails, 4)
	})

	t.Run("npm client get package urls", func(t *testing.T) {
		urls := testPackage1.GetUniqueURLs()
		require.Len(t, urls, 26)
	})

	t.Run("npm client get creation date", func(t *testing.T) {
		at := testPackage1.GetCreatedAt()
		require.EqualValues(t, "2017-07-03 20:58:20.378 +0000 UTC", at.String())
	})

	t.Run("npm client get modification date", func(t *testing.T) {
		at := testPackage1.GetModifiedAt()
		require.EqualValues(t, "2025-10-05 12:46:33.558 +0000 UTC", at.String())
	})

	t.Run("npm client get git repository url", func(t *testing.T) {
		git := testPackage1.GetGitURL()
		require.EqualValues(t, "git://github.com/lovell/detect-libc.git", git.String())
	})

	t.Run("npm client get git repository url (2 schemas)", func(t *testing.T) {
		git := testPackage2.GetGitURL()
		require.EqualValues(t, "git+https://github.com/facebook/react.git", git.String())
	})

	t.Run("npm client get package stats", func(t *testing.T) {
		client, err := NewClient(Options{})

		require.NoError(t, err)
		require.NotNil(t, client)

		stats, err := client.GetPackageStats(context.Background(), "", time.Hour*72)
		require.Error(t, err)
		require.Nil(t, stats)

		stats, err = client.GetPackageStats(context.Background(), "detect-libc", time.Hour*72)
		require.NoError(t, err)
		require.NotNil(t, stats)
	})
}
