package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/joekings2k/logistics-eta/db/sqlc"
	"github.com/joekings2k/logistics-eta/util"
	"github.com/lib/pq"

)


type CreateUserRequest struct {
	Name string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role string `json:"role" binding:"required,roles"`
}

type UserResponse struct {
	ID uuid.UUID `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
	Role string `json:"role"`
}

func newUserResponse(user db.User) UserResponse {
	return UserResponse{
		ID: user.ID,
		Name: user.Name,
		Email: user.Email,
		Role: user.Role,
	}
}


func (server *Server) CreateUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	// check if the req role is valid
	// role := util.Role(req.Role)
	// if !role.IsValid(){
	// 	ctx.JSON(http.StatusBadRequest, gin.H{"error":"invalid role"})
	// 	return
	// }

	arg := db.CreateUserParams{
		ID: uuid.New(),
		Name: req.Name,
		Email: req.Email,
		PasswordHash: hashedPassword,
		Role: req.Role,
	}
	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error);ok{
			switch pqErr.Code.Name(){
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, newUserResponse(user))
}

type LoginUserRequest struct {
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role string `json:"role" binding:"required,roles"`
}

type LoginUserResponse struct {
	AccessToken string `json:"access_token"`
	User UserResponse `json:"user"`

}

func (server *Server) LoginUser(ctx *gin.Context) {
	var req LoginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	user, err := server.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err  == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	err = util.CheckPassword(req.Password, user.PasswordHash)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}
	if req.Role != user.Role {
		msg := fmt.Sprintf("This user is not a %s", req.Role)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error":msg})
		return
	}
	accessToken, err := server.tokenMaker.CreateToken(user.ID, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
	}

	response := LoginUserResponse{
		AccessToken: accessToken,
		User: newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, response)
}