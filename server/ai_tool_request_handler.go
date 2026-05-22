package server

import (
	"database/sql"
	"net/http"

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
	UserID     int64  `json:"user_id" binding:"required,min=1"`
	Tool       string `json:"tool" binding:"required"`
	TokenLimit int64  `json:"token_limit" binding:"required,min=1"`
}

func (server *Server) createAiTool(ctx *gin.Context) {
	var req createAiToolRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	aiTool, err := server.store.CreateAiTool(ctx, db.CreateAiToolParams{
		UserID:     req.UserID,
		Tool:       req.Tool,
		TokenLimit: req.TokenLimit,
	})
	if err != nil {
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

	aiTools, err := server.store.ListAiTools(ctx, db.ListAiToolsParams{
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

	aiTool, err := server.store.UpdateAiToolTokenLimit(ctx, db.UpdateAiToolTokenLimitParams{
		TokenLimit: req.TokenLimit,
		ID:         uriReq.ID,
	})
	if err != nil {
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

	err := server.store.DeleteAiTool(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "ai tool deleted successfully",
	})
}
