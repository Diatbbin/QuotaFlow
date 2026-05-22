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

	tool := util.RandomTool()
	aiTool1 := createRandomAiToolForUser(t, createRandomUser(t).ID, tool)
	aiTool2 := createRandomAiToolForUser(t, createRandomUser(t).ID, tool)

	fmt.Printf(">> Initial spare tokens: t1: %v, t2: %v\n", db.SpareTokens(aiTool1), db.SpareTokens(aiTool2))

	n := 10
	tokens := int64(10)

	errs := make(chan error)
	results := make(chan db.TransferTokensTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTokensTx(context.Background(), db.TransferTokensTxParams{
				FromAiToolID: aiTool1.ID,
				ToAiToolID:   aiTool2.ID,
				Tokens:       tokens,
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
		require.Equal(t, aiTool1.ID, transfer.FromAiToolID)
		require.Equal(t, aiTool2.ID, transfer.ToAiToolID)
		require.Equal(t, tokens, transfer.Tokens)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTokenTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		from := result.FromAiTool
		require.NotEmpty(t, from)
		require.Equal(t, aiTool1.ID, from.ID)

		to := result.ToAiTool
		require.NotEmpty(t, to)
		require.Equal(t, aiTool2.ID, to.ID)

		fmt.Printf(">> after tx spare tokens: t1: %v, t2: %v\n", db.SpareTokens(from), db.SpareTokens(to))
		diff1 := db.SpareTokens(aiTool1) - db.SpareTokens(from)
		diff2 := db.SpareTokens(to) - db.SpareTokens(aiTool2)
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%tokens == 0)

		k := int(diff1 / tokens)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	updated1, err := store.GetAiTool(context.Background(), aiTool1.ID)
	require.NoError(t, err)

	updated2, err := store.GetAiTool(context.Background(), aiTool2.ID)
	require.NoError(t, err)

	fmt.Printf(">> Final spare tokens: t1: %v, t2: %v\n", db.SpareTokens(updated1), db.SpareTokens(updated2))
	require.Equal(t, db.SpareTokens(aiTool1)-int64(n)*tokens, db.SpareTokens(updated1))
	require.Equal(t, db.SpareTokens(aiTool2)+int64(n)*tokens, db.SpareTokens(updated2))
}

func TestTransferTokensTxDeadlock(t *testing.T) {
	store := db.NewStore(testDB)

	tool := util.RandomTool()
	aiTool1 := createRandomAiToolForUser(t, createRandomUser(t).ID, tool)
	aiTool2 := createRandomAiToolForUser(t, createRandomUser(t).ID, tool)

	fmt.Printf(">> Initial spare tokens: t1: %v, t2: %v\n", db.SpareTokens(aiTool1), db.SpareTokens(aiTool2))

	n := 20
	tokens := int64(10)

	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromID := aiTool1.ID
		toID := aiTool2.ID

		if i%2 == 1 {
			fromID = aiTool2.ID
			toID = aiTool1.ID
		}

		go func() {
			_, err := store.TransferTokensTx(context.Background(), db.TransferTokensTxParams{
				FromAiToolID: fromID,
				ToAiToolID:   toID,
				Tokens:       tokens,
			})
			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	updated1, err := store.GetAiTool(context.Background(), aiTool1.ID)
	require.NoError(t, err)

	updated2, err := store.GetAiTool(context.Background(), aiTool2.ID)
	require.NoError(t, err)

	fmt.Printf(">> Final spare tokens: t1: %v, t2: %v\n", db.SpareTokens(updated1), db.SpareTokens(updated2))
	require.Equal(t, db.SpareTokens(aiTool1), db.SpareTokens(updated1))
	require.Equal(t, db.SpareTokens(aiTool2), db.SpareTokens(updated2))
}

func TestTransferTokensTxInsufficientSpare(t *testing.T) {
	store := db.NewStore(testDB)

	tool := util.RandomTool()
	sender := createRandomAiToolForUser(t, createRandomUser(t).ID, tool)
	recipient := createRandomAiToolForUser(t, createRandomUser(t).ID, tool)

	_, err := testDB.ExecContext(context.Background(),
		`UPDATE ai_tools SET tokens_used = token_limit WHERE id = $1`, sender.ID)
	require.NoError(t, err)

	_, err = store.TransferTokensTx(context.Background(), db.TransferTokensTxParams{
		FromAiToolID: sender.ID,
		ToAiToolID:   recipient.ID,
		Tokens:       1,
	})
	require.ErrorIs(t, err, db.ErrInsufficientTokens)
}

func TestTransferTokensTxSameUser(t *testing.T) {
	store := db.NewStore(testDB)

	user := createRandomUser(t)
	from := createRandomAiToolForUser(t, user.ID, "cursor")
	to := createRandomAiToolForUser(t, user.ID, "copilot")

	_, err := store.TransferTokensTx(context.Background(), db.TransferTokensTxParams{
		FromAiToolID: from.ID,
		ToAiToolID:   to.ID,
		Tokens:       1,
	})
	require.ErrorIs(t, err, db.ErrTransferSameUser)
}

func TestTransferTokensTxDifferentAiTool(t *testing.T) {
	store := db.NewStore(testDB)

	from := createRandomAiToolForUser(t, createRandomUser(t).ID, "cursor")
	to := createRandomAiToolForUser(t, createRandomUser(t).ID, "copilot")

	_, err := store.TransferTokensTx(context.Background(), db.TransferTokensTxParams{
		FromAiToolID: from.ID,
		ToAiToolID:   to.ID,
		Tokens:       1,
	})
	require.ErrorIs(t, err, db.ErrTransferDifferentAiTool)
}
