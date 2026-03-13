package repository_test

import (
	"testing"

	"artifactor/internal/helpers"
	requests "artifactor/pkg/http"
	"artifactor/pkg/tokens"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateToken_Success(t *testing.T) {
	repo, cleanup := helpers.SetupRepo(t)
	defer cleanup()

	req := &requests.RegisterRequest{Admin: false, Products: []tokens.Product{}}
	token, err := repo.CreateToken(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	repo.PruneToken(token)
}

func TestFetchToken_Success(t *testing.T) {
	repo, cleanup := helpers.SetupRepo(t)
	defer cleanup()

	req := &requests.RegisterRequest{
		Admin:    false,
		Products: []tokens.Product{{Name: "test-product"}},
	}
	token, err := repo.CreateToken(req)
	require.NoError(t, err)
	defer repo.PruneToken(token)

	apiToken, err := repo.FetchToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, apiToken)
	assert.False(t, apiToken.Admin)
	assert.Len(t, apiToken.Products, 1)
	assert.Equal(t, "test-product", apiToken.Products[0].Name)
}

func TestFetchToken_NotFound(t *testing.T) {
	repo, cleanup := helpers.SetupRepo(t)
	defer cleanup()

	_, err := repo.FetchToken("nonexistent-token")
	assert.Error(t, err)
}

func TestTokenExists_Exists(t *testing.T) {
	repo, cleanup := helpers.SetupRepo(t)
	defer cleanup()

	req := &requests.RegisterRequest{}
	token, err := repo.CreateToken(req)
	require.NoError(t, err)
	defer repo.PruneToken(token)

	exists, err := repo.TokenExists(token)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestTokenExists_NotExists(t *testing.T) {
	repo, cleanup := helpers.SetupRepo(t)
	defer cleanup()

	exists, err := repo.TokenExists("nonexistent-token")
	assert.Error(t, err)
	assert.False(t, exists)
}

func TestIsAdmin_True(t *testing.T) {
	repo, cleanup := helpers.SetupRepo(t)
	defer cleanup()

	req := &requests.RegisterRequest{Admin: true}
	token, err := repo.CreateToken(req)
	require.NoError(t, err)
	defer repo.PruneToken(token)

	isAdmin, err := repo.IsAdmin(token)
	assert.NoError(t, err)
	assert.True(t, isAdmin)
}

func TestIsAdmin_False(t *testing.T) {
	repo, cleanup := helpers.SetupRepo(t)
	defer cleanup()

	req := &requests.RegisterRequest{Admin: false}
	token, err := repo.CreateToken(req)
	require.NoError(t, err)
	defer repo.PruneToken(token)

	isAdmin, err := repo.IsAdmin(token)
	assert.NoError(t, err)
	assert.False(t, isAdmin)
}

func TestPruneToken_Success(t *testing.T) {
	repo, cleanup := helpers.SetupRepo(t)
	defer cleanup()

	req := &requests.RegisterRequest{}
	token, err := repo.CreateToken(req)
	require.NoError(t, err)

	err = repo.PruneToken(token)
	assert.NoError(t, err)

	_, err = repo.FetchToken(token)
	assert.Error(t, err)
}

func TestPruneToken_NotFound(t *testing.T) {
	repo, cleanup := helpers.SetupRepo(t)
	defer cleanup()

	err := repo.PruneToken("nonexistent-token")
	assert.Error(t, err)
}
