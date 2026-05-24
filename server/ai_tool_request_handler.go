package server

import (
	"database/sql"
	"net/http"
	"errors"

	token "github.com/diatbbin/QuotaFlow/auth"
	"github.com/lib/pq"
	"github.com/gin-gonic/gin"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
)

type aiToolResponse struct {
	AiTool      db.AiTool `json:"ai_tool"`
	SpareTokens int64     `json:"spare_tokens"`
}

func toAiToolResponse(a db.AiTool) aiToolResponse {
	return aiToolResponse{
		AiTool:      a,
		SpareTokens: db.SpareTokens(a),
	}
}

type createAiToolRequest struct {
	Username   string `json:"username" 		binding:"required"`
	Tool       string `json:"tool"     		binding:"required"`
	TokenLimit int64  `json:"token_limit" 	binding:"required,min=1"`
}

func (server *Server) createAiTool(ctx *gin.Context) {
	var req createAiToolRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	aiTool, err := server.store.CreateAiTool(ctx, db.CreateAiToolParams{
		Username:   authPayload.Username,
		Tool:       req.Tool,
		TokenLimit: req.TokenLimit,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation", "foreign_key_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, toAiToolResponse(aiTool))
}

type getAiToolRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAiTool(ctx *gin.Context) {
	var req getAiToolRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	aiTool, err := server.store.GetAiTool(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if aiTool.Username != authPayload.Username {
		ctx.JSON(http.StatusUnauthorized, errorResponse(errors.New("ai tool does not belong to this user")))
		return
	}

	ctx.JSON(http.StatusOK, toAiToolResponse(aiTool))
}

type listAiToolsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAiTools(ctx *gin.Context) {
	var req listAiToolsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	aiTools, err := server.store.ListAiTools(ctx, db.ListAiToolsParams{
		Username: authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := make([]aiToolResponse, len(aiTools))
	for i, a := range aiTools {
		resp[i] = toAiToolResponse(a)
	}

	ctx.JSON(http.StatusOK, resp)
}

type updateAiToolURIRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateAiToolTokenLimitRequest struct {
	TokenLimit int64 `json:"token_limit" binding:"required,min=0"`
}

func (server *Server) updateAiToolTokenLimit(ctx *gin.Context) {
	var uriReq updateAiToolURIRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updateAiToolTokenLimitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	aiTool, err := server.store.UpdateAiToolTokenLimit(ctx, db.UpdateAiToolTokenLimitParams{
		TokenLimit: req.TokenLimit,
		ID:         uriReq.ID,
		Username:   authPayload.Username,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, toAiToolResponse(aiTool))
}

type updateAiToolTokensUsedRequest struct {
	TokensUsed int64 `json:"tokens_used" binding:"required,min=0"`
}

func (server *Server) updateAiToolTokensUsed(ctx *gin.Context) {
	var uriReq updateAiToolURIRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updateAiToolTokensUsedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	aiTool, err := server.store.UpdateAiToolTokensUsed(ctx, db.UpdateAiToolTokensUsedParams{
		TokensUsed: req.TokensUsed,
		ID:         uriReq.ID,
		Username:   authPayload.Username,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "check_violation":
				ctx.JSON(http.StatusBadRequest, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, toAiToolResponse(aiTool))
}

type deleteAiToolURIRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteAiTool(ctx *gin.Context) {
	var req deleteAiToolURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	err := server.store.DeleteAiTool(ctx, db.DeleteAiToolParams{
		ID:       req.ID,
		Username: authPayload.Username,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ai tool deleted successfully",
	})
}
