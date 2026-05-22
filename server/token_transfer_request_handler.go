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
	FromAiTool    aiToolResponse     `json:"from_ai_tool"`
	ToAiTool      aiToolResponse     `json:"to_ai_tool"`
}

func toTokenTransferResponse(t db.TokenTransfer, from, to db.AiTool) tokenTransferResponse {
	return tokenTransferResponse{
		TokenTransfer: t,
		FromAiTool:    toAiToolResponse(from),
		ToAiTool:      toAiToolResponse(to),
	}
}

type createTokenTransferRequest struct {
	FromAiToolID int64 `json:"from_ai_tool_id" binding:"required,min=1"`
	ToAiToolID   int64 `json:"to_ai_tool_id" binding:"required,min=1"`
	Tokens       int64 `json:"tokens" binding:"required,min=1"`
}

func (server *Server) createTokenTransfer(ctx *gin.Context) {
	var req createTokenTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if !server.validateTokenTransfer(ctx, req.FromAiToolID) || !server.validateTokenTransfer(ctx, req.ToAiToolID) {
		return
	}

	result, err := server.store.TransferTokensTx(ctx, db.TransferTokensTxParams{
		FromAiToolID: req.FromAiToolID,
		ToAiToolID:   req.ToAiToolID,
		Tokens:       req.Tokens,
	})
	if err != nil {
		switch {
		case errors.Is(err, db.ErrInsufficientTokens),
			errors.Is(err, db.ErrTransferSameUser),
			errors.Is(err, db.ErrTransferDifferentAiTool):
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
		default:
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		}
		return
	}

	ctx.JSON(http.StatusOK, toTokenTransferResponse(result.Transfer, result.FromAiTool, result.ToAiTool))
}

func (server *Server) validateTokenTransfer(ctx *gin.Context, aiToolID int64) bool {
	_, err := server.store.GetAiTool(ctx, aiToolID)
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
