package utils

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret"

func TestSignAndParseToken(t *testing.T) {
	claims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "some-uuid",
		},
		Admin: false,
	}

	token, err := SignToken(claims, testSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	output, err := ParseToken(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, "some-uuid", output.Subject)
	assert.False(t, output.Admin)
}

func TestSignAndParseToken_Admin(t *testing.T) {
	claims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "admin-uuid",
		},
		Admin: true,
	}
	token, err := SignToken(claims, testSecret)
	require.NoError(t, err)

	output, err := ParseToken(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, "admin-uuid", output.Subject)
	assert.True(t, output.Admin)
}

func TestParseToken_WrongSecret(t *testing.T) {
	claims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "some-uuid",
		},
		Admin: false,
	}

	token, err := SignToken(claims, testSecret)
	require.NoError(t, err)

	_, err = ParseToken(token, "wrong-secret")
	assert.Error(t, err)
}

func TestParseToken_InvalidToken(t *testing.T) {
	_, err := ParseToken("not-a-jwt", testSecret)
	assert.Error(t, err)
}

func TestParseToken_TamperedClaims(t *testing.T) {
	claims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "uuid-1",
		},
		Admin: true,
	}
	token, err := SignToken(claims, "other-secret")
	require.NoError(t, err)

	_, err = ParseToken(token, testSecret)
	assert.Error(t, err)
}
