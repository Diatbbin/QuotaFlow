package sqlc_test

import (
	"context"
	"database/sql"
	"testing"

	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	"github.com/diatbbin/QuotaFlow/util"
	"github.com/stretchr/testify/require"
)

func createRandomTokenTransfer(t *testing.T) db.TokenTransfer {
	from := createAiToolForRandomUser(t, util.RandomTool())
	to := createAiToolForRandomUser(t, util.RandomTool())

	arg := db.CreateTokenTransferParams{
		FromAiToolID: from.ID,
		ToAiToolID:   to.ID,
		Tokens:       util.RandomTokenUsed() + 1,
	}

	transfer, err := testQueries.CreateTokenTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAiToolID, transfer.FromAiToolID)
	require.Equal(t, arg.ToAiToolID, transfer.ToAiToolID)
	require.Equal(t, arg.Tokens, transfer.Tokens)

	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

func TestGetTokenTransfer(t *testing.T) {
	transfer1 := createRandomTokenTransfer(t)

	transfer2, err := testQueries.GetTokenTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAiToolID, transfer2.FromAiToolID)
	require.Equal(t, transfer1.ToAiToolID, transfer2.ToAiToolID)
	require.Equal(t, transfer1.Tokens, transfer2.Tokens)
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, 0)
}

func TestUpdateTokenTransferTokens(t *testing.T) {
	transfer1 := createRandomTokenTransfer(t)

	arg := db.UpdateTokenTransferTokensParams{
		Tokens: util.RandomTokenUsed() + 1,
		ID:     transfer1.ID,
	}

	transfer2, err := testQueries.UpdateTokenTransferTokens(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAiToolID, transfer2.FromAiToolID)
	require.Equal(t, transfer1.ToAiToolID, transfer2.ToAiToolID)
	require.Equal(t, arg.Tokens, transfer2.Tokens)
	require.WithinDuration(t, transfer1.CreatedAt, transfer2.CreatedAt, 0)
}

func TestDeleteTokenTransfer(t *testing.T) {
	transfer1 := createRandomTokenTransfer(t)
	err := testQueries.DeleteTokenTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)

	transfer2, err := testQueries.GetTokenTransfer(context.Background(), transfer1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, transfer2)
}

func TestListTokenTransfers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomTokenTransfer(t)
	}

	arg := db.ListTokenTransfersParams{
		Limit:  10,
		Offset: 5,
	}

	transfers, err := testQueries.ListTokenTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 10)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}
