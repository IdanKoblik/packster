package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret"

func TestSignAndParseToken(t *testing.T) {
	token, err := SignToken("some-uuid", false, testSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := ParseToken(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, "some-uuid", claims.Subject)
	assert.False(t, claims.Admin)
}

func TestSignAndParseToken_Admin(t *testing.T) {
	token, err := SignToken("admin-uuid", true, testSecret)
	require.NoError(t, err)

	claims, err := ParseToken(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, "admin-uuid", claims.Subject)
	assert.True(t, claims.Admin)
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, err := SignToken("some-uuid", false, testSecret)
	require.NoError(t, err)

	_, err = ParseToken(token, "wrong-secret")
	assert.Error(t, err)
}

func TestParseToken_InvalidToken(t *testing.T) {
	_, err := ParseToken("not-a-jwt", testSecret)
	assert.Error(t, err)
}

func TestParseToken_TamperedClaims(t *testing.T) {
	// A token signed with a different secret should be rejected
	token, err := SignToken("uuid-1", true, "other-secret")
	require.NoError(t, err)

	_, err = ParseToken(token, testSecret)
	assert.Error(t, err)
}
