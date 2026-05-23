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

func TestDeleteAiTool(t *testing.T) {
	aiTool1 := createAiToolForRandomUser(t, util.RandomTool())
	err := testQueries.DeleteAiTool(context.Background(), aiTool1.ID)
	require.NoError(t, err)

	aiTool2, err := testQueries.GetAiTool(context.Background(), aiTool1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, aiTool2)
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
