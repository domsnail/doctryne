package stack_exchange

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStackExchangeUsers(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug - 4,
	})))

	t.Run("stack exchange client prepare test", func(t *testing.T) {
		client, err := NewClient(Options{})

		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("stack exchange client get about me", func(t *testing.T) {
		client, err := NewClient(Options{
			AccessToken: os.Getenv("STACK_EXCHANGE_API_KEY"),
		})

		require.NoError(t, err)
		require.NotNil(t, client)

		me, raw, err := client.GetMe(context.Background())
		require.NoError(t, err)
		require.NotNil(t, raw)
		require.NotNil(t, me)
	})

	t.Run("stack exchange client get user test", func(t *testing.T) {
		client, err := NewClient(Options{
			AccessToken: os.Getenv("STACK_EXCHANGE_API_KEY"),
		})

		require.NoError(t, err)
		require.NotNil(t, client)

		users, raw, err := client.GetUsersByUsername(context.Background(), "qvineox")
		require.NoError(t, err)
		require.NotNil(t, raw)
		require.NotNil(t, users)

		require.Len(t, users, 1)
		require.EqualValues(t, 31134943, users[0].UserID)
		require.EqualValues(t, 39876030, users[0].AccountID)
		require.EqualValues(t, "qvineox", users[0].DisplayName)
		require.EqualValues(t, "registered", users[0].UserType)
		require.EqualValues(t, false, users[0].IsEmployee)

		require.NotEmpty(t, users[0].CreationDate)
		require.EqualValues(t, 1753356250, users[0].CreationDate.Unix())
		require.EqualValues(t, "2025-07-24 11:24:10 +0000 UTC", users[0].CreationDate.UTC().String())

		users, raw, err = client.GetUsersByUsername(context.Background(), "lysak")
		require.Len(t, users, 1)
		require.NoError(t, err)
		require.NotNil(t, raw)
		require.NotNil(t, users)
	})

	t.Run("stack exchange client get user by id test", func(t *testing.T) {
		client, err := NewClient(Options{
			AccessToken: os.Getenv("STACK_EXCHANGE_API_KEY"),
		})

		require.NoError(t, err)
		require.NotNil(t, client)

		user, raw, err := client.GetUserByID(context.Background(), 13775941)
		require.NoError(t, err)
		require.NotNil(t, raw)
		require.NotNil(t, user)
		require.EqualValues(t, "Lysak Yaroslav", user.DisplayName)
		require.EqualValues(t, "registered", user.UserType)
	})
}
