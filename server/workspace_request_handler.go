package server

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
)

type workspaceResponse struct {
	Workspace   db.Workspace `json:"workspace"`
	SpareTokens int64        `json:"spare_tokens"`
}

func toWorkspaceResponse(w db.Workspace) workspaceResponse {
	return workspaceResponse{
		Workspace:   w,
		SpareTokens: db.SpareTokens(w),
	}
}

type createWorkspaceRequest struct {
	Name       string `json:"name" binding:"required"`
	TokenLimit int64  `json:"token_limit" binding:"required,min=1"`
}

func (server *Server) createWorkspace(ctx *gin.Context) {
	var req createWorkspaceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	workspace, err := server.store.CreateWorkspace(ctx, db.CreateWorkspaceParams{
		Name:       req.Name,
		TokenLimit: req.TokenLimit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, toWorkspaceResponse(workspace))
}

type getWorkspaceRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getWorkspace(ctx *gin.Context) {
	var req getWorkspaceRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	workspace, err := server.store.GetWorkspace(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, toWorkspaceResponse(workspace))
}

type listWorkspacesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listWorkspaces(ctx *gin.Context) {
	var req listWorkspacesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	workspaces, err := server.store.ListWorkspaces(ctx, db.ListWorkspacesParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := make([]workspaceResponse, len(workspaces))
	for i, w := range workspaces {
		resp[i] = toWorkspaceResponse(w)
	}

	ctx.JSON(http.StatusOK, resp)
}

type updateWorkspaceURIRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateWorkspaceTokenLimitRequest struct {
	TokenLimit int64 `json:"token_limit" binding:"required,min=0"`
}

func (server *Server) updateWorkspaceTokenLimit(ctx *gin.Context) {
	var uriReq updateWorkspaceURIRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var req updateWorkspaceTokenLimitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	workspace, err := server.store.UpdateWorkspaceTokenLimit(ctx, db.UpdateWorkspaceTokenLimitParams{
		TokenLimit: req.TokenLimit,
		ID:         uriReq.ID,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, toWorkspaceResponse(workspace))
}

type deleteWorkspaceURIRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteWorkspace(ctx *gin.Context) {
	var req deleteWorkspaceURIRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	err := server.store.DeleteWorkspace(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "workspace deleted successfully",
	})
}
