package sqlc_test

import (
	"context"
	"fmt"
	"testing"

	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	"github.com/diatbbin/QuotaFlow/util"
	"github.com/stretchr/testify/require"
)

func TestTransferTokensTx(t *testing.T) {
	store := db.NewStore(testDB)

	workspace1, err := testQueries.CreateWorkspace(context.Background(), db.CreateWorkspaceParams{
		Name:       util.RandomWorkspaceName(),
		TokenLimit: util.RandomTokenLimit(),
	})
	require.NoError(t, err)

	workspace2, err := testQueries.CreateWorkspace(context.Background(), db.CreateWorkspaceParams{
		Name:       util.RandomWorkspaceName(),
		TokenLimit: util.RandomTokenLimit(),
	})
	require.NoError(t, err)

	fmt.Printf(">> Initial spare tokens: w1: %v, w2: %v\n", db.SpareTokens(workspace1), db.SpareTokens(workspace2))

	n := 10
	tokens := int64(10)

	errs := make(chan error)
	results := make(chan db.TransferTokensTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTokensTx(context.Background(), db.TransferTokensTxParams{
				FromWorkspaceID: workspace1.ID,
				ToWorkspaceID:   workspace2.ID,
				Tokens:          tokens,
			})

			errs <- err
			results <- result
		}()
	}

	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)

		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, workspace1.ID, transfer.FromWorkspaceID)
		require.Equal(t, workspace2.ID, transfer.ToWorkspaceID)
		require.Equal(t, tokens, transfer.Tokens)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTokenTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		from := result.FromWorkspace
		require.NotEmpty(t, from)
		require.Equal(t, workspace1.ID, from.ID)

		to := result.ToWorkspace
		require.NotEmpty(t, to)
		require.Equal(t, workspace2.ID, to.ID)

		fmt.Printf(">> after tx spare tokens: w1: %v, w2: %v\n", db.SpareTokens(from), db.SpareTokens(to))
		diff1 := db.SpareTokens(workspace1) - db.SpareTokens(from)
		diff2 := db.SpareTokens(to) - db.SpareTokens(workspace2)
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%tokens == 0)

		k := int(diff1 / tokens)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	updated1, err := store.GetWorkspace(context.Background(), workspace1.ID)
	require.NoError(t, err)

	updated2, err := store.GetWorkspace(context.Background(), workspace2.ID)
	require.NoError(t, err)

	fmt.Printf(">> Final spare tokens: w1: %v, w2: %v\n", db.SpareTokens(updated1), db.SpareTokens(updated2))
	require.Equal(t, db.SpareTokens(workspace1)-int64(n)*tokens, db.SpareTokens(updated1))
	require.Equal(t, db.SpareTokens(workspace2)+int64(n)*tokens, db.SpareTokens(updated2))
}

func TestTransferTokensTxDeadlock(t *testing.T) {
	store := db.NewStore(testDB)

	workspace1 := createRandomWorkspace(t)
	workspace2 := createRandomWorkspace(t)

	fmt.Printf(">> Initial spare tokens: w1: %v, w2: %v\n", db.SpareTokens(workspace1), db.SpareTokens(workspace2))

	n := 20
	tokens := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromID := workspace1.ID
		toID := workspace2.ID

		if i%2 == 1 {
			fromID = workspace2.ID
			toID = workspace1.ID
		}

		go func() {
			_, err := store.TransferTokensTx(context.Background(), db.TransferTokensTxParams{
				FromWorkspaceID: fromID,
				ToWorkspaceID:   toID,
				Tokens:          tokens,
			})
			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	updated1, err := store.GetWorkspace(context.Background(), workspace1.ID)
	require.NoError(t, err)

	updated2, err := store.GetWorkspace(context.Background(), workspace2.ID)
	require.NoError(t, err)

	fmt.Printf(">> Final spare tokens: w1: %v, w2: %v\n", db.SpareTokens(updated1), db.SpareTokens(updated2))
	require.Equal(t, db.SpareTokens(workspace1), db.SpareTokens(updated1))
	require.Equal(t, db.SpareTokens(workspace2), db.SpareTokens(updated2))
}

func TestTransferTokensTxInsufficientSpare(t *testing.T) {
	store := db.NewStore(testDB)

	sender := createRandomWorkspace(t)
	recipient := createRandomWorkspace(t)

	_, err := testDB.ExecContext(context.Background(),
		`UPDATE workspaces SET tokens_used = token_limit WHERE id = $1`, sender.ID)
	require.NoError(t, err)

	_, err = store.TransferTokensTx(context.Background(), db.TransferTokensTxParams{
		FromWorkspaceID: sender.ID,
		ToWorkspaceID:   recipient.ID,
		Tokens:          1,
	})
	require.ErrorIs(t, err, db.ErrInsufficientTokens)
}
