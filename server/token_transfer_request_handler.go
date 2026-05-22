package server

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
)

type tokenTransferResponse struct {
	TokenTransfer db.TokenTransfer `json:"token_transfer"`
	FromWorkspace workspaceResponse  `json:"from_workspace"`
	ToWorkspace   workspaceResponse  `json:"to_workspace"`
}

func toTokenTransferResponse(t db.TokenTransfer, from, to db.Workspace) tokenTransferResponse {
	return tokenTransferResponse{
		TokenTransfer: t,
		FromWorkspace: toWorkspaceResponse(from),
		ToWorkspace:   toWorkspaceResponse(to),
	}
}

type createTokenTransferRequest struct {
	FromWorkspaceID int64 `json:"from_workspace_id" binding:"required,min=1"`
	ToWorkspaceID   int64 `json:"to_workspace_id" binding:"required,min=1"`
	Tokens          int64 `json:"tokens" binding:"required,min=1"`
}

func (server *Server) createTokenTransfer(ctx *gin.Context) {
	var req createTokenTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if (!server.validateTokenTransfer(ctx, req.FromWorkspaceID) || !server.validateTokenTransfer(ctx, req.ToWorkspaceID)) {
		return
	}

	result, err := server.store.TransferTokensTx(ctx, db.TransferTokensTxParams{
		FromWorkspaceID: req.FromWorkspaceID,
		ToWorkspaceID:   req.ToWorkspaceID,
		Tokens:          req.Tokens,
	})
	if err != nil {
		if errors.Is(err, db.ErrInsufficientTokens) {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, toTokenTransferResponse(result.Transfer, result.FromWorkspace, result.ToWorkspace))
}

func (server *Server) validateTokenTransfer(ctx *gin.Context, workspaceID int64) bool {
	_, err := server.store.GetWorkspace(ctx, workspaceID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return false
	}

	return true
}
