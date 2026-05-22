package sqlc_test

import (
	"context"
	"database/sql"
	"testing"

	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	"github.com/diatbbin/QuotaFlow/util"
	"github.com/stretchr/testify/require"
)

func createRandomWorkspace(t *testing.T) db.Workspace {
	arg := db.CreateWorkspaceParams{
		Name:       util.RandomWorkspaceName(),
		TokenLimit: util.RandomTokenLimit(),
	}

	workspace, err := testQueries.CreateWorkspace(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, workspace)

	require.Equal(t, arg.Name, workspace.Name)
	require.Equal(t, arg.TokenLimit, workspace.TokenLimit)
	require.Equal(t, int64(0), workspace.TokensUsed)

	require.NotZero(t, workspace.ID)
	require.NotZero(t, workspace.CreatedAt)

	return workspace
}

func TestCreateWorkspace(t *testing.T) {
	createRandomWorkspace(t)
}

func TestGetWorkspace(t *testing.T) {
	workspace1 := createRandomWorkspace(t)

	workspace2, err := testQueries.GetWorkspace(context.Background(), workspace1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, workspace2)

	require.Equal(t, workspace1.ID, workspace2.ID)
	require.Equal(t, workspace1.Name, workspace2.Name)
	require.Equal(t, workspace1.TokenLimit, workspace2.TokenLimit)
	require.Equal(t, workspace1.TokensUsed, workspace2.TokensUsed)
	require.WithinDuration(t, workspace1.CreatedAt, workspace2.CreatedAt, 0)
}

func TestDeleteWorkspace(t *testing.T) {
	workspace1 := createRandomWorkspace(t)
	err := testQueries.DeleteWorkspace(context.Background(), workspace1.ID)
	require.NoError(t, err)

	workspace2, err := testQueries.GetWorkspace(context.Background(), workspace1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, workspace2)
}

func TestListWorkspaces(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomWorkspace(t)
	}

	arg := db.ListWorkspacesParams{
		Limit:  10,
		Offset: 5,
	}

	workspaces, err := testQueries.ListWorkspaces(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, workspaces, 10)

	for _, workspace := range workspaces {
		require.NotEmpty(t, workspace)
	}
}
