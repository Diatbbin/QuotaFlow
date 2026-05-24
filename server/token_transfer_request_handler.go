package server

import (
	"database/sql"
	"errors"
	"net/http"
	"fmt"
	
	token "github.com/diatbbin/QuotaFlow/auth"
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
		FromAiTool:    toAiToolResponse(from, fmt.Sprintf("tokens %d transferred successfully from ai tool %s to ai tool %s", t.Tokens, from.Tool, to.Tool)),
		ToAiTool:      toAiToolResponse(to, fmt.Sprintf("tokens %d received successfully from ai tool %s to ai tool %s", t.Tokens, from.Tool, to.Tool)),
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

	fromAiTool, fromValid := server.validateTokenTransfer(ctx, req.FromAiToolID)
	if !fromValid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAiTool.Username != authPayload.Username {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("Sender's ai tool does not belong to this user")))
		return
	}

	_, toValid := server.validateTokenTransfer(ctx, req.ToAiToolID)
	if !toValid {
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

func (server *Server) validateTokenTransfer(ctx *gin.Context, aiToolID int64) (db.AiTool, bool) {
	account, err := server.store.GetAiTool(ctx, aiToolID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}

	return account, true
}
