package gapi

import (
	"context"
	"database/sql"
	"github.com/golang/mock/gomock"
	mockdb "github.com/mariobasic/simplebank/db/mock"
	db "github.com/mariobasic/simplebank/db/sqlc"
	"github.com/mariobasic/simplebank/pb"
	"github.com/mariobasic/simplebank/token"
	"github.com/mariobasic/simplebank/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

func TestServer_UpdateUser(t *testing.T) {
	user, _ := randomUser(t)
	newName := util.RandomOwner()
	newEmail := util.RandomEmail()
	tests := []struct {
		name           string
		body           *pb.UpdateUserRequest
		buildStubs     func(store *mockdb.MockStore)
		buildContext   func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponses func(*testing.T, *pb.UpdateUserResponse, error)
	}{
		{
			name: "OK",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					Username: user.Username,
					FullName: sql.NullString{String: newName, Valid: true},
					Email:    sql.NullString{String: newEmail, Valid: true},
				}

				store.EXPECT().
					UpdateUser(gomock.Any(), arg).
					Times(1).
					Return(db.User{
						Username:          user.Username,
						HashedPassword:    user.HashedPassword,
						Email:             newEmail,
						FullName:          newName,
						PasswordChangedAt: user.PasswordChangedAt,
						CreatedAt:         user.CreatedAt,
						IsEmailVerified:   user.IsEmailVerified,
					}, nil)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponses: func(t *testing.T, r *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, r)
				uptUsr := r.GetUser()
				require.Equal(t, user.Username, uptUsr.Username)
				require.Equal(t, newName, uptUsr.FullName)
				require.Equal(t, newEmail, uptUsr.Email)
			},
		},
		{
			name: "UserNotFound",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponses: func(t *testing.T, r *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.NotFound, status.Code(err))
			},
		},
		{
			name: "InvalidEmail",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    util.ToPtr("invalid-email"),
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0).
					Return(db.User{}, sql.ErrNoRows)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, time.Minute)
			},
			checkResponses: func(t *testing.T, r *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.InvalidArgument, status.Code(err))
			},
		},
		{
			name: "ExpiredToken",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0).
					Return(db.User{}, sql.ErrNoRows)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(t, tokenMaker, user.Username, -time.Minute)
			},
			checkResponses: func(t *testing.T, r *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.Unauthenticated, status.Code(err))
			},
		},
		{
			name: "NoAuthorization",
			body: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0).
					Return(db.User{}, sql.ErrNoRows)

			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},
			checkResponses: func(t *testing.T, r *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.Unauthenticated, status.Code(err))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)

			tt.buildStubs(store)

			server := NewTestServer(t, store, nil)
			ctx := tt.buildContext(t, server.tokenMaker)
			createUser, err := server.UpdateUser(ctx, tt.body)
			tt.checkResponses(t, createUser, err)

		})
	}
}
