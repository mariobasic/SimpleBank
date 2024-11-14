package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golang/mock/gomock"
	mockdb "github.com/mariobasic/simplebank/db/mock"
	db "github.com/mariobasic/simplebank/db/sqlc"
	"github.com/mariobasic/simplebank/pb"
	"github.com/mariobasic/simplebank/util"
	"github.com/mariobasic/simplebank/worker"
	mockwk "github.com/mariobasic/simplebank/worker/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"testing"
)

type eqCreateUserTxParamsMatcher struct {
	arg      db.CreateUserTxParams
	password string
	user     db.User
}

func (e eqCreateUserTxParamsMatcher) Matches(x any) bool {
	actualArg, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, actualArg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = actualArg.HashedPassword
	if !reflect.DeepEqual(e.arg.CreateUserParams, actualArg.CreateUserParams) {
		return false
	}

	err = actualArg.AfterCreate(e.user)
	return err == nil
}

func (e eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserTxParams(arg db.CreateUserTxParams, password string, user db.User) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{arg, password, user}
}
func TestServer_CreateUser(t *testing.T) {
	user, password := randomUser(t)

	tests := []struct {
		name           string
		body           *pb.CreateUserRequest
		buildStubs     func(store *mockdb.MockStore, tskDist *mockwk.MockTaskDistributor)
		checkResponses func(*testing.T, *pb.CreateUserResponse, error)
	}{
		{
			name: "OK",
			body: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, tskDist *mockwk.MockTaskDistributor) {
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					},
				}

				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(arg, password, user)).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)

				taskPayload := &worker.PayloadSendVerifyEmail{
					Username: user.Username,
				}

				tskDist.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponses: func(t *testing.T, r *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, r)
				createdUsr := r.GetUser()
				require.Equal(t, user.Username, createdUsr.Username)
				require.Equal(t, user.FullName, createdUsr.FullName)
				require.Equal(t, user.Email, createdUsr.Email)
			},
		},
		{
			name: "InternalError",
			body: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, tskDist *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)

				tskDist.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0).
					Return(nil)
			},
			checkResponses: func(t *testing.T, r *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.Internal, status.Code(err))
			},
		},
		{
			name: "DuplicateUsername",
			body: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, tskDist *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, db.ErrUniqueViolation)
				tskDist.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0).
					Return(nil)
			},
			checkResponses: func(t *testing.T, r *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.PermissionDenied, status.Code(err))
			},
		},
		{
			name: "InvalidUsername",
			body: &pb.CreateUserRequest{
				Username: "invalid-user#1",
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, tskDist *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
				tskDist.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponses: func(t *testing.T, r *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.InvalidArgument, status.Code(err))
			},
		},
		{
			name: "InvalidEmail",
			body: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    "invalid-email",
			},
			buildStubs: func(store *mockdb.MockStore, tskDist *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
				tskDist.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponses: func(t *testing.T, r *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.InvalidArgument, status.Code(err))
			},
		},
		{
			name: "TooShortPassword",
			body: &pb.CreateUserRequest{
				Username: user.Username,
				Password: "a",
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, tskDist *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(0)
				tskDist.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponses: func(t *testing.T, r *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				require.Equal(t, codes.InvalidArgument, status.Code(err))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)

			tskCtrl := gomock.NewController(t)
			tskDist := mockwk.NewMockTaskDistributor(tskCtrl)

			tt.buildStubs(store, tskDist)

			server := NewTestServer(t, store, tskDist)

			createUser, err := server.CreateUser(context.Background(), tt.body)
			tt.checkResponses(t, createUser, err)

		})
	}

}

func randomUser(t *testing.T) (db.User, string) {
	password := util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	return db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}, password
}
