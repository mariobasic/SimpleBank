package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mariobasic/simplebank/util"
	"log"
	"os"
	"testing"
)

var testStore Store

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	pool, err := pgxpool.New(context.Background(), config.DB.Source)
	if err != nil {
		log.Fatal("cannot connect to database:", err)
	}
	testStore = NewStore(pool)
	os.Exit(m.Run())
}
