package epss

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const downloadURL = "https://gitlab.domsnail.ru/public-data/epss/-/raw/main/epss_scores-current.csv.gz?ref_type=heads"

func TestClient_DownloadDatabase(t *testing.T) {
	slog.SetDefault(
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: false,
			Level:     slog.LevelDebug,
		})),
	)

	t.Run("test download current database", func(t *testing.T) {
		opts := WithDownloadURL(downloadURL)
		c := NewClient(opts)

		db, err := c.DownloadCurrentDatabase(context.Background())
		require.NoError(t, err)
		require.NotEmpty(t, db)

		require.GreaterOrEqual(t, db.Records, 377333)
		require.EqualValues(t, "2026-09-21", db.Date.Format("2006-01-02"))
		require.EqualValues(t, "v2026.06.15", db.Version)
	})

	t.Run("test download dated database", func(t *testing.T) {
		opts := WithDownloadURL(downloadURL)
		c := NewClient(opts)

		db, err := c.DownloadDatabaseOnDate(context.Background(), time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		require.NotEmpty(t, db)

		require.Len(t, db.Records, 377333)
		require.EqualValues(t, "2026-09-21", db.Date.Format("2006-01-02"))
		require.EqualValues(t, "v2026.06.15", db.Version)
	})
}
