package github

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_GetAdvisories(t *testing.T) {
	t.Run("test ghsa query records", func(t *testing.T) {
		c := NewClient()

		records, err := c.GetAdvisories(context.Background(), &AdvisoriesQueryOptions{
			Sort:      "published",
			Direction: "asc",
		}, 25, 0)

		require.NoError(t, err)
		require.Len(t, records, 25)
	})

	t.Run("test ghsa records query with date range", func(t *testing.T) {
		c := NewClient()

		from, _ := time.Parse("2006-01-02", "2024-09-01")
		to, _ := time.Parse("2006-01-02", "2024-09-30")

		records, err := c.GetAdvisories(context.Background(), &AdvisoriesQueryOptions{
			ModifiedAfter:  &from,
			ModifiedBefore: &to,
			Sort:           "published",
			Direction:      "asc",
		}, 10, 0)

		require.NoError(t, err)
		require.Len(t, records, 10)

		for _, record := range records {
			require.True(t, (record.UpdatedAt.After(from) || record.PublishedAt.After(from)) && record.UpdatedAt.Before(to) || record.PublishedAt.Before(to))
		}
	})

	t.Run("test ghsa record query by ghsa id", func(t *testing.T) {
		c := NewClient()

		records, err := c.GetAdvisories(context.Background(), &AdvisoriesQueryOptions{
			GhsaID:    "GHSA-6cwv-x26c-w2q4",
			Sort:      "published",
			Direction: "asc",
		}, 2, 0)

		require.NoError(t, err)
		require.Len(t, records, 1)

		require.EqualValues(t, "GHSA-6cwv-x26c-w2q4", records[0].GhsaId)
		require.EqualValues(t, "CVE-2018-8768", records[0].CveId)
	})

	t.Run("test ghsa record query by cve id", func(t *testing.T) {
		c := NewClient()

		records, err := c.GetAdvisories(context.Background(), &AdvisoriesQueryOptions{
			CveID:     "CVE-2022-26332",
			Sort:      "published",
			Direction: "asc",
		}, 2, 0)

		require.NoError(t, err)
		require.Len(t, records, 1)

		require.EqualValues(t, "CVE-2022-26332", records[0].CveId)
	})
}
