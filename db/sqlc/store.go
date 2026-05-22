package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrInsufficientTokens = errors.New("insufficient spare tokens to transfer")

func SpareTokens(w Workspace) int64 {
	return w.TokenLimit - w.TokensUsed
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
	FromWorkspaceID int64 `json:"from_workspace_id"`
	ToWorkspaceID   int64 `json:"to_workspace_id"`
	Tokens          int64 `json:"tokens"`
}

type TransferTokensTxResult struct {
	Transfer      TokenTransfer `json:"transfer"`
	FromWorkspace Workspace     `json:"from_workspace"`
	ToWorkspace   Workspace     `json:"to_workspace"`
}

func (store *Store) TransferTokensTx(ctx context.Context, arg TransferTokensTxParams) (TransferTokensTxResult, error) {
	var result TransferTokensTxResult

	if arg.FromWorkspaceID == arg.ToWorkspaceID {
		return result, fmt.Errorf("token sender and recipient cannot be the same")
	}
	if arg.Tokens <= 0 {
		return result, fmt.Errorf("tokens tranferred must be > 0")
	}

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		var from Workspace

		if arg.FromWorkspaceID < arg.ToWorkspaceID {
			from, _, err = lockWorkspaces(ctx, q, arg.FromWorkspaceID, arg.ToWorkspaceID)
		} else {
			_, from, err = lockWorkspaces(ctx, q, arg.ToWorkspaceID, arg.FromWorkspaceID)
		}
		if err != nil {
			return err
		}
		if SpareTokens(from) < arg.Tokens {
			return ErrInsufficientTokens
		}

		result.Transfer, err = q.CreateTokenTransfer(ctx, CreateTokenTransferParams{
			FromWorkspaceID: arg.FromWorkspaceID,
			ToWorkspaceID:   arg.ToWorkspaceID,
			Tokens:          arg.Tokens,
		})
		if err != nil {
			return err
		}

		if arg.FromWorkspaceID < arg.ToWorkspaceID {
			result.FromWorkspace, result.ToWorkspace, err = shiftTokenLimits(
				ctx, q, arg.FromWorkspaceID, -arg.Tokens, arg.ToWorkspaceID, arg.Tokens,
			)
		} else {
			result.ToWorkspace, result.FromWorkspace, err = shiftTokenLimits(
				ctx, q, arg.ToWorkspaceID, arg.Tokens, arg.FromWorkspaceID, -arg.Tokens,
			)
		}

		return err
	})

	return result, err
}

func lockWorkspaces(
	ctx context.Context,
	q *Queries,
	workspaceID1 int64,
	workspaceID2 int64,
) (workspace1 Workspace, workspace2 Workspace, err error) {
	workspace1, err = q.GetWorkspaceForUpdate(ctx, workspaceID1)
	if err != nil {
		return
	}
	workspace2, err = q.GetWorkspaceForUpdate(ctx, workspaceID2)
	return
}

func shiftTokenLimits(
	ctx context.Context,
	q *Queries,
	workspaceID1 int64,
	delta1 int64,
	workspaceID2 int64,
	delta2 int64,
) (workspace1 Workspace, workspace2 Workspace, err error) {
	workspace1, err = q.AddWorkspaceTokenLimit(ctx, AddWorkspaceTokenLimitParams{
		ID:     workspaceID1,
		Amount: delta1,
	})
	if err != nil {
		return
	}

	workspace2, err = q.AddWorkspaceTokenLimit(ctx, AddWorkspaceTokenLimitParams{
		ID:     workspaceID2,
		Amount: delta2,
	})
	return
}
