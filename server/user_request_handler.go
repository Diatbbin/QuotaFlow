package server

import (
	"net/http"
	"time"
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/diatbbin/QuotaFlow/db/sqlc"
	"github.com/diatbbin/QuotaFlow/util"
)

type userResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	Message   string `json:"message"`
}

func toCreateUserResponse(u db.User, message string) userResponse {
	return userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		Message:   message,
	}
}

type createUserRequest struct {
	Username string `json:"username" binding:"required,min=3,alphanum"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	user, err := server.store.CreateUser(ctx, db.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, toCreateUserResponse(user, fmt.Sprintf("user %s created successfully", user.Username)))
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,min=3,alphanum"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginUserResponse struct {
	AccessToken string `json:"access_token"`
	User        userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.ValidatePassword(req.Password, user.PasswordHash)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(user.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		AccessToken: accessToken,
		User:        toCreateUserResponse(user, fmt.Sprintf("user %s logged in successfully", user.Username)),
	}

	ctx.JSON(http.StatusOK, rsp)
}
