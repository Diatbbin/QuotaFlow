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

	router.POST("/ai-tools", server.createAiTool)
	router.GET("/ai-tools/:id", server.getAiTool)
	router.GET("/ai-tools", server.listAiTools)
	router.PUT("/ai-tools/:id", server.updateAiToolTokenLimit)
	router.DELETE("/ai-tools/:id", server.deleteAiTool)

	router.POST("/token-transfers", server.createTokenTransfer)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
