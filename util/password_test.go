package util

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

func Test_checkPassword(t *testing.T) {

	tests := []struct {
		name     string
		password string
	}{
		{name: "first", password: RandomString(6)},
		{name: "first", password: RandomString(6)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			hp, err := HashPassword(tt.password)
			require.NoError(t, err)
			require.NotEmpty(t, hp)

			err = CheckPassword(tt.password, hp)
			require.NoError(t, err)

			err = CheckPassword(RandomString(6), hp)
			require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

			hp2, err := HashPassword(tt.password)
			require.NoError(t, err)
			require.NotEmpty(t, hp2)
			require.NotEqual(t, hp, hp2)
		})
	}
}
