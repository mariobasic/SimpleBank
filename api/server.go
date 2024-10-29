package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/mariobasic/simplebank/db/sqlc"
	"github.com/mariobasic/simplebank/token"
	"github.com/mariobasic/simplebank/util"
	"log"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) *Server {
	tokenMaker, err := token.NewPasetoMaker(config.Token.SymmetricKey)
	//tokenMaker, err := token.NewJWTMaker(config.Token.SymmetricKey)
	if err != nil {
		log.Fatal(err)
	}
	server := &Server{config: config, store: store, tokenMaker: tokenMaker}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("currency", validCurrency)
		if err != nil {
			log.Fatalf("error registering validation: %v", err)
		}
	}

	setupRouter(server)
	return server
}

func setupRouter(server *Server) {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	authRouter := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRouter.POST("/accounts", server.createAccount)
	authRouter.GET("/accounts/:id", server.getAccount)
	authRouter.GET("/accounts", server.listAccounts)

	authRouter.POST("/transfers", server.createTransfer)

	server.router = router
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)

}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
