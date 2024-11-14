package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mariobasic/simplebank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) User {
	password, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)
	createUserParam := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: password,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testStore.CreateUser(context.Background(), createUserParam)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, createUserParam.Username, user.Username)
	require.Equal(t, createUserParam.HashedPassword, user.HashedPassword)
	require.Equal(t, createUserParam.FullName, user.FullName)
	require.Equal(t, createUserParam.Email, user.Email)

	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestQueries_CreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestQueries_GetUser(t *testing.T) {
	tests := []struct {
		name    string
		db      Store
		arg     User
		wantErr bool
	}{
		{name: "first", db: testStore, arg: createRandomUser(t), wantErr: false},
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

func TestQueries_UpdateUser(t *testing.T) {
	tests := []struct {
		name        string
		db          Store
		arg         User
		update      func(u User) UpdateUserParams
		updateCheck func(old User, got User, p UpdateUserParams)
		wantErr     bool
	}{
		{
			name: "first",
			db:   testStore,
			arg:  createRandomUser(t),
			update: func(u User) UpdateUserParams {
				return UpdateUserParams{
					Username: u.Username,
					FullName: pgtype.Text{String: util.RandomOwner(), Valid: true}}
			},
			updateCheck: func(old User, got User, up UpdateUserParams) {
				require.Equal(t, old.Username, got.Username)
				require.Equal(t, up.FullName.String, got.FullName)
				require.Equal(t, old.Email, got.Email)
			},
			wantErr: false,
		},
		{
			name: "Only Email Update",
			db:   testStore,
			arg:  createRandomUser(t),
			update: func(u User) UpdateUserParams {
				return UpdateUserParams{
					Username: u.Username,
					Email:    pgtype.Text{String: util.RandomEmail(), Valid: true},
				}
			},
			updateCheck: func(old User, got User, up UpdateUserParams) {
				require.Equal(t, old.Username, got.Username)
				require.Equal(t, old.FullName, got.FullName)
				require.Equal(t, up.Email.String, got.Email)
			},
			wantErr: false,
		},
		{
			name: "Only Password",
			db:   testStore,
			arg:  createRandomUser(t),
			update: func(u User) UpdateUserParams {
				password, err := util.HashPassword(util.RandomString(6))
				require.NoError(t, err)

				return UpdateUserParams{
					Username:       u.Username,
					HashedPassword: pgtype.Text{String: password, Valid: true},
				}
			},
			updateCheck: func(old User, got User, up UpdateUserParams) {
				require.Equal(t, old.Username, got.Username)
				require.Equal(t, old.FullName, got.FullName)
				require.Equal(t, old.Email, got.Email)
				require.Equal(t, up.HashedPassword.String, got.HashedPassword)
			},
			wantErr: false,
		},
		{
			name: "All fields",
			db:   testStore,
			arg:  createRandomUser(t),
			update: func(u User) UpdateUserParams {
				password, err := util.HashPassword(util.RandomString(6))
				require.NoError(t, err)

				return UpdateUserParams{
					Username:       u.Username,
					FullName:       pgtype.Text{String: util.RandomOwner(), Valid: true},
					Email:          pgtype.Text{String: util.RandomEmail(), Valid: true},
					HashedPassword: pgtype.Text{String: password, Valid: true},
				}
			},
			updateCheck: func(old User, got User, up UpdateUserParams) {
				require.Equal(t, old.Username, got.Username)
				require.Equal(t, up.FullName.String, got.FullName)
				require.Equal(t, up.Email.String, got.Email)
				require.Equal(t, up.HashedPassword.String, got.HashedPassword)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateUserParams := tt.update(tt.arg)

			got, err := tt.db.UpdateUser(context.Background(), updateUserParams)

			require.NoError(t, err)
			require.NotEmpty(t, got)

			tt.updateCheck(tt.arg, got, updateUserParams)

			require.WithinDuration(t, tt.arg.CreatedAt, got.CreatedAt, time.Second)
			require.WithinDuration(t, tt.arg.PasswordChangedAt, got.PasswordChangedAt, time.Second)

		})
	}
}
