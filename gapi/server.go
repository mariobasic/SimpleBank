package gapi

import (
	db "github.com/mariobasic/simplebank/db/sqlc"
	"github.com/mariobasic/simplebank/pb"
	"github.com/mariobasic/simplebank/token"
	"github.com/mariobasic/simplebank/util"
	"github.com/mariobasic/simplebank/worker"
	"log"
)

type Server struct {
	pb.UnimplementedSimpleBankServer
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

func NewServer(config util.Config, store db.Store, distributor worker.TaskDistributor) *Server {
	tokenMaker, err := token.NewPasetoMaker(config.Token.SymmetricKey)
	if err != nil {
		log.Fatal(err)
	}

	return &Server{config: config, store: store, tokenMaker: tokenMaker, taskDistributor: distributor}
}
