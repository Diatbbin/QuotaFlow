package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrInsufficientTokens      = errors.New("insufficient spare tokens to transfer")
	ErrTransferSameUser        = errors.New("cannot transfer tokens to the same user")
	ErrTransferDifferentAiTool = errors.New("transfers must be between the same ai tool")
)

func SpareTokens(a AiTool) int64 {
	return a.TokenLimit - a.TokensUsed
}

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
		db:      db,
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

type TransferTokensTxParams struct {
	FromAiToolID int64 `json:"from_ai_tool_id"`
	ToAiToolID   int64 `json:"to_ai_tool_id"`
	Tokens       int64 `json:"tokens"`
}

type TransferTokensTxResult struct {
	Transfer   TokenTransfer `json:"transfer"`
	FromAiTool AiTool        `json:"from_ai_tool"`
	ToAiTool   AiTool        `json:"to_ai_tool"`
}

func (store *Store) TransferTokensTx(ctx context.Context, arg TransferTokensTxParams) (TransferTokensTxResult, error) {
	var result TransferTokensTxResult

	if arg.FromAiToolID == arg.ToAiToolID {
		return result, fmt.Errorf("token sender and recipient cannot be the same")
	}
	if arg.Tokens <= 0 {
		return result, fmt.Errorf("tokens tranferred must be > 0")
	}

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		var from, to AiTool

		if arg.FromAiToolID < arg.ToAiToolID {
			from, to, err = lockAiTools(ctx, q, arg.FromAiToolID, arg.ToAiToolID)
		} else {
			to, from, err = lockAiTools(ctx, q, arg.ToAiToolID, arg.FromAiToolID)
		}
		if err != nil {
			return err
		}

		if from.Username == to.Username {
			return ErrTransferSameUser
		}
		if from.Tool != to.Tool {
			return ErrTransferDifferentAiTool
		}

		if SpareTokens(from) < arg.Tokens {
			return ErrInsufficientTokens
		}

		result.Transfer, err = q.CreateTokenTransfer(ctx, CreateTokenTransferParams{
			FromAiToolID: arg.FromAiToolID,
			ToAiToolID:   arg.ToAiToolID,
			Tokens:       arg.Tokens,
		})
		if err != nil {
			return err
		}

		if arg.FromAiToolID < arg.ToAiToolID {
			result.FromAiTool, result.ToAiTool, err = shiftTokenLimits(
				ctx, q, arg.FromAiToolID, -arg.Tokens, arg.ToAiToolID, arg.Tokens,
			)
		} else {
			result.ToAiTool, result.FromAiTool, err = shiftTokenLimits(
				ctx, q, arg.ToAiToolID, arg.Tokens, arg.FromAiToolID, -arg.Tokens,
			)
		}

		return err
	})

	return result, err
}

func lockAiTools(
	ctx context.Context,
	q *Queries,
	aiToolID1 int64,
	aiToolID2 int64,
) (aiTool1 AiTool, aiTool2 AiTool, err error) {
	aiTool1, err = q.GetAiToolForUpdate(ctx, aiToolID1)
	if err != nil {
		return
	}
	aiTool2, err = q.GetAiToolForUpdate(ctx, aiToolID2)
	return
}

func shiftTokenLimits(
	ctx context.Context,
	q *Queries,
	aiToolID1 int64,
	delta1 int64,
	aiToolID2 int64,
	delta2 int64,
) (aiTool1 AiTool, aiTool2 AiTool, err error) {
	aiTool1, err = q.AddAiToolTokenLimit(ctx, AddAiToolTokenLimitParams{
		ID:     aiToolID1,
		Amount: delta1,
	})
	if err != nil {
		return
	}

	aiTool2, err = q.AddAiToolTokenLimit(ctx, AddAiToolTokenLimitParams{
		ID:     aiToolID2,
		Amount: delta2,
	})
	return
}
