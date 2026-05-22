package server

import (
	"github.com/gin-gonic/gin"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
)

type Server struct {
	store  *db.Store
	router *gin.Engine
}

func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/workspaces", server.createWorkspace)
	router.GET("/workspaces/:id", server.getWorkspace)
	router.GET("/workspaces", server.listWorkspaces)
	router.PUT("/workspaces/:id", server.updateWorkspaceTokenLimit)
	router.DELETE("/workspaces/:id", server.deleteWorkspace)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
