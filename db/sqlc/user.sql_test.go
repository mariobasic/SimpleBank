package db

import (
	"context"
	"github.com/mariobasic/simplebank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) []User {
	var users []User
	tests := []struct {
		name    string
		db      *Queries
		arg     func() CreateUserParams
		wantErr bool
	}{
		{"first", testQueries, func() CreateUserParams {
			password, err := util.HashPassword(util.RandomString(6))
			require.NoError(t, err)
			return CreateUserParams{
				Username:       util.RandomOwner(),
				HashedPassword: password,
				FullName:       util.RandomOwner(),
				Email:          util.RandomEmail()}
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createUserParam := tt.arg()
			user, err := tt.db.CreateUser(context.Background(), createUserParam)
			require.NoError(t, err)
			require.NotEmpty(t, user)

			require.Equal(t, createUserParam.Username, user.Username)
			require.Equal(t, createUserParam.HashedPassword, user.HashedPassword)
			require.Equal(t, createUserParam.FullName, user.FullName)
			require.Equal(t, createUserParam.Email, user.Email)

			require.True(t, user.PasswordChangedAt.IsZero())
			require.NotZero(t, user.CreatedAt)
			users = append(users, user)
		})
	}
	return users
}

func TestQueries_CreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestQueries_GetUser(t *testing.T) {
	tests := []struct {
		name    string
		db      *Queries
		arg     User
		wantErr bool
	}{
		{name: "first", db: testQueries, arg: createRandomUser(t)[0], wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.db.GetUser(context.Background(), tt.arg.Username)
			require.NoError(t, err)
			require.NotEmpty(t, got)

			require.Equal(t, tt.arg.Username, got.Username)
			require.Equal(t, tt.arg.HashedPassword, got.HashedPassword)
			require.Equal(t, tt.arg.FullName, got.FullName)
			require.Equal(t, tt.arg.Email, got.Email)
			require.WithinDuration(t, tt.arg.CreatedAt, got.CreatedAt, time.Second)
			require.WithinDuration(t, tt.arg.PasswordChangedAt, got.PasswordChangedAt, time.Second)
		})
	}
}
