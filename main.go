package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/mariobasic/simplebank/api"
	db "github.com/mariobasic/simplebank/db/sqlc"
	"github.com/mariobasic/simplebank/util"
	"log"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Error loading config:", err)
	}
	conn, err := sql.Open(config.DB.Driver, config.DB.Source)
	if err != nil {
		log.Fatal("cannot connect to db", err)
	}

	server := api.NewServer(db.NewStore(conn))
	err = server.Start(config.Server.Address)
	if err != nil {
		log.Fatal("cannot start server", err)
	}

}
