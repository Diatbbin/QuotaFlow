package sqlc_test

import (
	"context"
	"database/sql"
	"testing"

	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	"github.com/diatbbin/QuotaFlow/util"
	"github.com/stretchr/testify/require"
)

func createAiToolForSpecificUserAndTool(t *testing.T, username string, tool string) db.AiTool {
	arg := db.CreateAiToolParams{
		Username:   username,
		Tool:       tool,
		TokenLimit: util.RandomTokenLimit(),
	}

	aiTool, err := testQueries.CreateAiTool(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, aiTool)

	require.Equal(t, arg.Username, aiTool.Username)
	require.Equal(t, arg.Tool, aiTool.Tool)
	require.Equal(t, arg.TokenLimit, aiTool.TokenLimit)
	require.Equal(t, int64(0), aiTool.TokensUsed)

	require.NotZero(t, aiTool.ID)
	require.NotZero(t, aiTool.CreatedAt)

	return aiTool
}

func createAiToolForRandomUser(t *testing.T, tool string) db.AiTool {
	return createAiToolForSpecificUserAndTool(t, createRandomUser(t).Username, tool)
}

func TestGetAiTool(t *testing.T) {
	aiTool1 := createAiToolForRandomUser(t, util.RandomTool())

	aiTool2, err := testQueries.GetAiTool(context.Background(), aiTool1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, aiTool2)

	require.Equal(t, aiTool1.ID, aiTool2.ID)
	require.Equal(t, aiTool1.Username, aiTool2.Username)
	require.Equal(t, aiTool1.Tool, aiTool2.Tool)
	require.Equal(t, aiTool1.TokenLimit, aiTool2.TokenLimit)
	require.Equal(t, aiTool1.TokensUsed, aiTool2.TokensUsed)
	require.WithinDuration(t, aiTool1.CreatedAt, aiTool2.CreatedAt, 0)
}

func TestUpdateAiToolTokenLimit(t *testing.T) {
	aiTool1 := createAiToolForRandomUser(t, util.RandomTool())
	newLimit := aiTool1.TokenLimit + 100

	aiTool2, err := testQueries.UpdateAiToolTokenLimit(context.Background(), db.UpdateAiToolTokenLimitParams{
		TokenLimit: newLimit,
		ID:         aiTool1.ID,
		Username:   aiTool1.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, aiTool2)

	require.Equal(t, aiTool1.ID, aiTool2.ID)
	require.Equal(t, aiTool1.Username, aiTool2.Username)
	require.Equal(t, aiTool1.Tool, aiTool2.Tool)
	require.Equal(t, newLimit, aiTool2.TokenLimit)
	require.Equal(t, aiTool1.TokensUsed, aiTool2.TokensUsed)
	require.WithinDuration(t, aiTool1.CreatedAt, aiTool2.CreatedAt, 0)

	aiTool3, err := testQueries.GetAiTool(context.Background(), aiTool1.ID)
	require.NoError(t, err)
	require.Equal(t, newLimit, aiTool3.TokenLimit)
}

func TestUpdateAiToolTokenLimitWrongUser(t *testing.T) {
	aiTool1 := createAiToolForRandomUser(t, util.RandomTool())
	otherUser := createRandomUser(t)

	_, err := testQueries.UpdateAiToolTokenLimit(context.Background(), db.UpdateAiToolTokenLimitParams{
		TokenLimit: aiTool1.TokenLimit + 50,
		ID:         aiTool1.ID,
		Username:   otherUser.Username,
	})
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())

	unchanged, err := testQueries.GetAiTool(context.Background(), aiTool1.ID)
	require.NoError(t, err)
	require.Equal(t, aiTool1.TokenLimit, unchanged.TokenLimit)
}

func TestUpdateAiToolTokensUsed(t *testing.T) {
	aiTool1 := createAiToolForRandomUser(t, util.RandomTool())
	newUsed := util.RandomTokenUsed() + 1
	require.LessOrEqual(t, newUsed, aiTool1.TokenLimit)

	aiTool2, err := testQueries.UpdateAiToolTokensUsed(context.Background(), db.UpdateAiToolTokensUsedParams{
		TokensUsed: newUsed,
		ID:         aiTool1.ID,
		Username:   aiTool1.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, aiTool2)

	require.Equal(t, aiTool1.ID, aiTool2.ID)
	require.Equal(t, aiTool1.Username, aiTool2.Username)
	require.Equal(t, aiTool1.Tool, aiTool2.Tool)
	require.Equal(t, aiTool1.TokenLimit, aiTool2.TokenLimit)
	require.Equal(t, newUsed, aiTool2.TokensUsed)
	require.WithinDuration(t, aiTool1.CreatedAt, aiTool2.CreatedAt, 0)

	aiTool3, err := testQueries.GetAiTool(context.Background(), aiTool1.ID)
	require.NoError(t, err)
	require.Equal(t, newUsed, aiTool3.TokensUsed)
}

func TestUpdateAiToolTokensUsedWrongUser(t *testing.T) {
	aiTool1 := createAiToolForRandomUser(t, util.RandomTool())
	otherUser := createRandomUser(t)

	_, err := testQueries.UpdateAiToolTokensUsed(context.Background(), db.UpdateAiToolTokensUsedParams{
		TokensUsed: 1,
		ID:         aiTool1.ID,
		Username:   otherUser.Username,
	})
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())

	unchanged, err := testQueries.GetAiTool(context.Background(), aiTool1.ID)
	require.NoError(t, err)
	require.Equal(t, aiTool1.TokensUsed, unchanged.TokensUsed)
}

func TestUpdateAiToolTokensUsedExceedsLimit(t *testing.T) {
	aiTool1 := createAiToolForRandomUser(t, util.RandomTool())

	_, err := testQueries.UpdateAiToolTokensUsed(context.Background(), db.UpdateAiToolTokensUsedParams{
		TokensUsed: aiTool1.TokenLimit + 1,
		ID:         aiTool1.ID,
		Username:   aiTool1.Username,
	})
	require.Error(t, err)

	unchanged, err := testQueries.GetAiTool(context.Background(), aiTool1.ID)
	require.NoError(t, err)
	require.Equal(t, aiTool1.TokensUsed, unchanged.TokensUsed)
}

func TestUpdateAiToolThatDoesNotExist(t *testing.T) {
	user := createRandomUser(t)

	_, err := testQueries.UpdateAiToolTokenLimit(context.Background(), db.UpdateAiToolTokenLimitParams{
		TokenLimit: 500,
		ID:         999999999,
		Username:   user.Username,
	})
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestDeleteAiTool(t *testing.T) {
	aiTool1 := createAiToolForRandomUser(t, util.RandomTool())
	err := testQueries.DeleteAiTool(context.Background(), db.DeleteAiToolParams{
		ID:       aiTool1.ID,
		Username: aiTool1.Username,
	})
	require.NoError(t, err)

	aiTool2, err := testQueries.GetAiTool(context.Background(), aiTool1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, aiTool2)
}

func TestDeleteAiToolCascadesTokenTransfers(t *testing.T) {
	from := createAiToolForRandomUser(t, util.RandomTool())
	to := createAiToolForRandomUser(t, util.RandomTool())

	transfer, err := testQueries.CreateTokenTransfer(context.Background(), db.CreateTokenTransferParams{
		FromAiToolID: from.ID,
		ToAiToolID:   to.ID,
		Tokens:       util.RandomTokenUsed() + 1,
	})
	require.NoError(t, err)

	err = testQueries.DeleteAiTool(context.Background(), db.DeleteAiToolParams{
		ID:       from.ID,
		Username: from.Username,
	})
	require.NoError(t, err)

	_, err = testQueries.GetTokenTransfer(context.Background(), transfer.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestListAiTools(t *testing.T) {
	var lastAiTool db.AiTool
	for i := 0; i < 10; i++ {
		lastAiTool = createAiToolForRandomUser(t, util.RandomTool())
	}

	arg := db.ListAiToolsParams{
		Username: lastAiTool.Username,
		Limit:  10,
		Offset: 0,
	}

	aiTools, err := testQueries.ListAiTools(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, aiTools)

	for _, aiTool := range aiTools {
		require.NotEmpty(t, aiTool)
		require.Equal(t, lastAiTool.ID, aiTool.ID)
	}
}
