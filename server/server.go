package server

import (
	"log"

	"github.com/diatbbin/QuotaFlow/util"
	"github.com/gin-gonic/gin"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	token "github.com/diatbbin/QuotaFlow/auth"
)

type Server struct {
	config 		util.Config
	store  		*db.Store
	router 		*gin.Engine
	tokenMaker  *token.PasetoMaker
}

func NewServer(store *db.Store, config util.Config) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		log.Fatal("cannot create token maker:", err)
	}

	server := &Server{
		config: config,
		store: store,
		tokenMaker: tokenMaker,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	router.POST("/ai-tools", server.createAiTool)
	router.GET("/ai-tools/:id", server.getAiTool)
	router.GET("/ai-tools", server.listAiTools)
	router.PUT("/ai-tools/:id", server.updateAiToolTokenLimit)
	router.DELETE("/ai-tools/:id", server.deleteAiTool)

	router.POST("/token-transfers", server.createTokenTransfer)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
