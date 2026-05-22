package sqlc_test

import (
	"context"
	"database/sql"
	"testing"

	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	"github.com/diatbbin/QuotaFlow/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) db.User {
	user, err := testQueries.CreateUser(context.Background(), db.CreateUserParams{
		Email:        util.RandomString(10) + "@test.com",
		Username:     util.RandomString(8),
		PasswordHash: util.RandomString(32),
	})
	require.NoError(t, err)
	require.NotEmpty(t, user)
	return user
}

func createRandomAiToolForUser(t *testing.T, userID int64, tool string) db.AiTool {
	arg := db.CreateAiToolParams{
		UserID:     userID,
		Tool:       tool,
		TokenLimit: util.RandomTokenLimit(),
	}

	aiTool, err := testQueries.CreateAiTool(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, aiTool)

	require.Equal(t, arg.UserID, aiTool.UserID)
	require.Equal(t, arg.Tool, aiTool.Tool)
	require.Equal(t, arg.TokenLimit, aiTool.TokenLimit)
	require.Equal(t, int64(0), aiTool.TokensUsed)

	require.NotZero(t, aiTool.ID)
	require.NotZero(t, aiTool.CreatedAt)

	return aiTool
}

func createRandomAiTool(t *testing.T) db.AiTool {
	return createRandomAiToolForUser(t, createRandomUser(t).ID, util.RandomTool())
}

func TestCreateAiTool(t *testing.T) {
	createRandomAiTool(t)
}

func TestGetAiTool(t *testing.T) {
	aiTool1 := createRandomAiTool(t)

	aiTool2, err := testQueries.GetAiTool(context.Background(), aiTool1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, aiTool2)

	require.Equal(t, aiTool1.ID, aiTool2.ID)
	require.Equal(t, aiTool1.UserID, aiTool2.UserID)
	require.Equal(t, aiTool1.Tool, aiTool2.Tool)
	require.Equal(t, aiTool1.TokenLimit, aiTool2.TokenLimit)
	require.Equal(t, aiTool1.TokensUsed, aiTool2.TokensUsed)
	require.WithinDuration(t, aiTool1.CreatedAt, aiTool2.CreatedAt, 0)
}

func TestDeleteAiTool(t *testing.T) {
	aiTool1 := createRandomAiTool(t)
	err := testQueries.DeleteAiTool(context.Background(), aiTool1.ID)
	require.NoError(t, err)

	aiTool2, err := testQueries.GetAiTool(context.Background(), aiTool1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, aiTool2)
}

func TestListAiTools(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAiTool(t)
	}

	arg := db.ListAiToolsParams{
		Limit:  10,
		Offset: 5,
	}

	aiTools, err := testQueries.ListAiTools(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, aiTools, 10)

	for _, aiTool := range aiTools {
		require.NotEmpty(t, aiTool)
	}
}
