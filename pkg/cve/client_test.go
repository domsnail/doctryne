package cve

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_GetRecords(t *testing.T) {
	t.Run("test cve query records", func(t *testing.T) {
		c := NewClient()

		records, err := c.GetRecords(context.Background(), RecordsQueryOptions{}, 0, 10)
		require.NoError(t, err)
		require.EqualValues(t, 0, records.StartIndex)
		require.EqualValues(t, 10, records.ResultsPerPage)
		require.NotZero(t, records.TotalResults)

		require.Len(t, records.Vulnerabilities, 10)
	})

	t.Run("test single cve query record", func(t *testing.T) {
		c := NewClient()

		records, err := c.GetRecords(context.Background(), RecordsQueryOptions{
			CveIDs: []string{"CVE-2026-85050"},
		}, 0, 1)

		require.NoError(t, err)
		require.EqualValues(t, 0, records.StartIndex)
		require.EqualValues(t, 1, records.ResultsPerPage)
		require.EqualValues(t, 1, records.TotalResults)

		require.Len(t, records.Vulnerabilities, 1)
	})
}
