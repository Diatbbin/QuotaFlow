package sqlc_test

import (
	"context"
	"database/sql"
	"testing"

	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	"github.com/diatbbin/QuotaFlow/util"
	"github.com/stretchr/testify/require"
)

func createRandomUsageEvent(t *testing.T) db.UsageEvent {
	arg := db.CreateUsageEventParams{
		AiToolID: createRandomAiTool(t).ID,
		Tokens:   util.RandomTokenUsed() + 1,
	}

	event, err := testQueries.CreateUsageEvent(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, event)

	require.Equal(t, arg.AiToolID, event.AiToolID)
	require.Equal(t, arg.Tokens, event.Tokens)

	require.NotZero(t, event.ID)
	require.NotZero(t, event.CreatedAt)

	return event
}

func TestCreateUsageEvent(t *testing.T) {
	createRandomUsageEvent(t)
}

func TestGetUsageEvent(t *testing.T) {
	event1 := createRandomUsageEvent(t)

	event2, err := testQueries.GetUsageEvent(context.Background(), event1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, event2)

	require.Equal(t, event1.ID, event2.ID)
	require.Equal(t, event1.AiToolID, event2.AiToolID)
	require.Equal(t, event1.Tokens, event2.Tokens)
	require.WithinDuration(t, event1.CreatedAt, event2.CreatedAt, 0)
}

func TestUpdateUsageEventTokens(t *testing.T) {
	event1 := createRandomUsageEvent(t)

	arg := db.UpdateUsageEventTokensParams{
		Tokens: util.RandomTokenUsed() + 1,
		ID:     event1.ID,
	}

	event2, err := testQueries.UpdateUsageEventTokens(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, event2)

	require.Equal(t, event1.ID, event2.ID)
	require.Equal(t, event1.AiToolID, event2.AiToolID)
	require.Equal(t, arg.Tokens, event2.Tokens)
	require.WithinDuration(t, event1.CreatedAt, event2.CreatedAt, 0)
}

func TestDeleteUsageEvent(t *testing.T) {
	event1 := createRandomUsageEvent(t)
	err := testQueries.DeleteUsageEvent(context.Background(), event1.ID)
	require.NoError(t, err)

	event2, err := testQueries.GetUsageEvent(context.Background(), event1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, event2)
}

func TestListUsageEvents(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomUsageEvent(t)
	}

	arg := db.ListUsageEventsParams{
		Limit:  10,
		Offset: 5,
	}

	events, err := testQueries.ListUsageEvents(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, events, 10)

	for _, event := range events {
		require.NotEmpty(t, event)
	}
}
