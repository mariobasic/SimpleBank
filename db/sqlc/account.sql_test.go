package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/mariobasic/simplebank/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomAccount(t *testing.T) []Account {
	user := createRandomUser(t)
	var accounts []Account
	tests := []struct {
		name    string
		db      Store
		arg     CreateAccountParams
		wantErr bool
	}{
		{"first", testStore, CreateAccountParams{Owner: user.Username, Balance: util.RandomMoney(), Currency: util.RandomCurrency()}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := tt.db.CreateAccount(context.Background(), tt.arg)
			require.NoError(t, err)
			require.NotEmpty(t, account)

			require.Equal(t, tt.arg.Owner, account.Owner)
			require.Equal(t, tt.arg.Balance, account.Balance)
			require.Equal(t, tt.arg.Currency, account.Currency)

			require.NotZero(t, account.ID)
			require.NotZero(t, account.CreatedAt)
			accounts = append(accounts, account)
		})
	}
	return accounts
}

func TestQueries_CreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestQueries_GetAccount(t *testing.T) {

	tests := []struct {
		name    string
		db      Store
		arg     Account
		wantErr bool
	}{
		{"GetAccount1", testStore, createRandomAccount(t)[0], false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.GetAccount(context.Background(), tt.arg.ID)
			require.NoError(t, err)
			require.NotEmpty(t, got)
			require.Equal(t, tt.arg.ID, got.ID)
			require.Equal(t, tt.arg.Owner, got.Owner)
			require.Equal(t, tt.arg.Balance, got.Balance)
			require.Equal(t, tt.arg.Currency, got.Currency)
			require.WithinDuration(t, tt.arg.CreatedAt, got.CreatedAt, time.Second)
		})
	}
}

func TestQueries_UpdateAccount(t *testing.T) {
	account := createRandomAccount(t)[0]
	tests := []struct {
		name    string
		account Account
		db      Store
		arg     UpdateAccountParams
		wantErr bool
	}{
		{"UpdateAccount", account, testStore, UpdateAccountParams{ID: account.ID, Balance: util.RandomMoney()}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.db.UpdateAccount(context.Background(), tt.arg)
			require.NoError(t, err)
			require.NotEmpty(t, got)
			require.Equal(t, tt.account.ID, got.ID)
			require.Equal(t, tt.arg.Balance, got.Balance)
			require.Equal(t, tt.account.Currency, got.Currency)
			require.Equal(t, tt.account.Owner, got.Owner)
			require.WithinDuration(t, tt.account.CreatedAt, got.CreatedAt, time.Second)
		})
	}
}

func TestQueries_DeleteAccount(t *testing.T) {
	account := createRandomAccount(t)[0]
	tests := []struct {
		name    string
		db      Store
		arg     int64
		wantErr bool
	}{
		{"DeleteAccount", testStore, account.ID, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.db.DeleteAccount(context.Background(), tt.arg)
			require.NoError(t, err)

			got, err := tt.db.GetAccount(context.Background(), tt.arg)
			require.Error(t, err)
			require.EqualError(t, err, pgx.ErrNoRows.Error())
			require.Empty(t, got)
		})
	}
}

func TestQueries_ListAccounts(t *testing.T) {
	var lastAccount Account

	for range 10 {
		lastAccount = createRandomAccount(t)[0]
	}

	tests := []struct {
		name    string
		db      Store
		arg     ListAccountsParams
		want    int
		wantErr bool
	}{
		{
			"ListAccounts",
			testStore,
			ListAccountsParams{Owner: lastAccount.Owner, Limit: 5, Offset: 0},
			5,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.db.ListAccounts(context.Background(), tt.arg)
			require.NoError(t, err)
			require.NotEmpty(t, got)
			for _, account := range got {
				require.NotEmpty(t, account)
				require.Equal(t, lastAccount.Owner, account.Owner)
			}
		})
	}
}
