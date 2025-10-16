package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/joekings2k/logistics-eta/db/sqlc"
	"github.com/joekings2k/logistics-eta/token"
	"github.com/joekings2k/logistics-eta/util"
)


type Server struct {
	config util.Config
	store db.Store
	tokenMaker token.Maker
	router *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err )
	}
	server := &Server{
		config:config,
		store: store,
		tokenMaker: tokenMaker,
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate);ok{
		v.RegisterValidation("roles", ValidRoles)
	}

	server.setupRouter()

	return  server, nil
}

func (server *Server)setupRouter() {
	router := gin.Default()
	router.GET("/",server.checkHealth)

	// user routes 
	userRoute := router.Group("/users")
	userRoute.POST("/login", server.LoginUser)
	userRoute.POST("/register", server.CreateUser)
	
	server.router = router
	
}

func (server *Server) Start(addres string)error{
	return server.router.Run(addres)
}

func errorResponse(err error) gin.H{
	return  gin.H{"error":err.Error()}
}
