package gapi

import (
	"context"
	"fmt"
	db "github.com/mariobasic/simplebank/db/sqlc"
	"github.com/mariobasic/simplebank/token"
	"github.com/mariobasic/simplebank/util"
	"github.com/mariobasic/simplebank/worker"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"testing"
	"time"
)

func NewTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{Token: struct {
		SymmetricKey    string        `mapstructure:"symmetric_key"`
		AccessDuration  time.Duration `mapstructure:"access_duration"`
		RefreshDuration time.Duration `mapstructure:"refresh_duration"`
	}{SymmetricKey: util.RandomString(32), AccessDuration: time.Minute, RefreshDuration: time.Hour}}

	return NewServer(config, store, taskDistributor)
}

func newContextWithBearerToken(t *testing.T, tokenMaker token.Maker, username string, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)

	return metadata.NewIncomingContext(context.Background(), metadata.MD{
		authorizationHeader: []string{bearerToken},
	})
}
