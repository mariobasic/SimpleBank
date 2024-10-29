package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/mariobasic/simplebank/db/sqlc"
	"github.com/mariobasic/simplebank/util"
	"os"
	"testing"
	"time"
)

func NewTestServer(_ *testing.T, store db.Store) *Server {
	config := util.Config{Token: struct {
		SymmetricKey   string        `mapstructure:"symmetric_key"`
		AccessDuration time.Duration `mapstructure:"access_duration"`
	}{SymmetricKey: util.RandomString(32), AccessDuration: time.Minute}}

	return NewServer(config, store)
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
