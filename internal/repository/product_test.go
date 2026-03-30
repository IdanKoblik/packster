package repository_test

import (
	"packster/internal/helpers"
	"packster/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testProductToken = "test-product-token"

func makeProduct(name string) *types.Product {
	return &types.Product{
		Name:      name,
		GroupName: "",
		Tokens: map[string]types.TokenPermissions{
			testProductToken: {
				Maintainer: true,
				Download:   true,
				Upload:     true,
				Delete:     true,
			},
		},
		Versions: map[string]types.Version{},
	}
}

func TestCreateProduct_Success(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	product := makeProduct("test-create-product")
	err := repo.CreateProduct(product)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-create-product", "", testProductToken, true)
}

func TestCreateProduct_Duplicate(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	product := makeProduct("test-dup-product")
	err := repo.CreateProduct(product)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-dup-product", "", testProductToken, true)

	err = repo.CreateProduct(makeProduct("test-dup-product"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product already exists")
}

func TestFetchProduct_Success(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	product := makeProduct("test-fetch-product")
	err := repo.CreateProduct(product)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-fetch-product", "", testProductToken, true)

	fetched, err := repo.FetchProduct("test-fetch-product", "")
	assert.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, "test-fetch-product", fetched.Name)
}

func TestFetchProduct_NotFound(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	fetched, err := repo.FetchProduct("nonexistent-product-xyz-12345", "")
	assert.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestDeleteProduct_Success(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	product := makeProduct("test-delete-product")
	err := repo.CreateProduct(product)
	require.NoError(t, err)

	err = repo.DeleteProduct("test-delete-product", "", testProductToken, true)
	assert.NoError(t, err)

	fetched, err := repo.FetchProduct("test-delete-product", "")
	assert.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestDeleteProduct_MissingPermission(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	product := makeProduct("test-delete-noperm")
	err := repo.CreateProduct(product)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-delete-noperm", "", testProductToken, true)

	err = repo.DeleteProduct("test-delete-noperm", "", "unknown-token", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing delete permission")
}

func TestAddToken_Success(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken, "new-token")
	defer cleanup()

	product := makeProduct("test-addtoken-product")
	err := repo.CreateProduct(product)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-addtoken-product", "", testProductToken, true)

	newPerms := types.TokenPermissions{Download: true}
	err = repo.AddToken("test-addtoken-product", "", testProductToken, "new-token", newPerms, true)
	assert.NoError(t, err)
}

func TestAddToken_ProductNotFound(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	newPerms := types.TokenPermissions{Download: true}
	err := repo.AddToken("nonexistent-product", "", testProductToken, "new-token", newPerms, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product not found")
}

func TestAddToken_MissingMaintainerPermission(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	product := makeProduct("test-addtoken-noperm")
	err := repo.CreateProduct(product)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-addtoken-noperm", "", testProductToken, true)

	newPerms := types.TokenPermissions{Download: true}
	err = repo.AddToken("test-addtoken-noperm", "", "unknown-token", "new-token", newPerms, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing maintainer permission")
}

func TestDeleteToken_Success(t *testing.T) {
	const targetToken = "token-to-delete"
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken, targetToken)
	defer cleanup()

	product := &types.Product{
		Name:      "test-deletetoken-product",
		GroupName: "",
		Tokens: map[string]types.TokenPermissions{
			testProductToken: {Maintainer: true},
			targetToken:      {Download: true},
		},
		Versions: map[string]types.Version{},
	}
	err := repo.CreateProduct(product)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-deletetoken-product", "", testProductToken, true)

	err = repo.DeleteToken("test-deletetoken-product", "", testProductToken, targetToken, true)
	assert.NoError(t, err)
}

func TestDeleteToken_ProductNotFound(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	err := repo.DeleteToken("nonexistent-product", "", testProductToken, "some-token", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product not found")
}

func TestDeleteToken_MissingMaintainerPermission(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	product := makeProduct("test-deletetoken-noperm")
	err := repo.CreateProduct(product)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-deletetoken-noperm", "", testProductToken, true)

	err = repo.DeleteToken("test-deletetoken-noperm", "", "unknown-token", "some-token", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing maintainer permission")
}

func TestAddVersion_Success(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	product := makeProduct("test-addversion-product")
	err := repo.CreateProduct(product)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-addversion-product", "", testProductToken, true)

	v := types.Version{Path: "/some/path/file.zip", Checksum: "abc123"}
	err = repo.AddVersion("test-addversion-product", "", "1.0.0", testProductToken, false, v)
	assert.NoError(t, err)

	fetched, err := repo.FetchProduct("test-addversion-product", "")
	require.NoError(t, err)
	assert.Equal(t, v, fetched.Versions["1.0.0"])
}

func TestAddVersion_ProductNotFound(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	v := types.Version{Path: "/some/path/file.zip", Checksum: "abc123"}
	err := repo.AddVersion("nonexistent-product", "", "1.0.0", testProductToken, false, v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product not found")
}

func TestAddVersion_VersionAlreadyExists(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	product := makeProduct("test-addversion-dup")
	err := repo.CreateProduct(product)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-addversion-dup", "", testProductToken, true)

	v := types.Version{Path: "/some/path/file.zip", Checksum: "abc123"}
	err = repo.AddVersion("test-addversion-dup", "", "1.0.0", testProductToken, false, v)
	require.NoError(t, err)

	err = repo.AddVersion("test-addversion-dup", "", "1.0.0", testProductToken, false, v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "version already exists")
}

func TestAddVersion_MissingUploadPermission(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	product := makeProduct("test-addversion-noperm")
	err := repo.CreateProduct(product)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-addversion-noperm", "", testProductToken, true)

	v := types.Version{Path: "/some/path/file.zip", Checksum: "abc123"}
	err = repo.AddVersion("test-addversion-noperm", "", "1.0.0", "unknown-token", false, v)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing upload permission")
}

func TestCreateProduct_SameNameDifferentGroups(t *testing.T) {
	repo, cleanup := helpers.SetupProductRepo(t, testProductToken)
	defer cleanup()

	p1 := &types.Product{
		Name:      "test-group-product",
		GroupName: "group-a",
		Tokens: map[string]types.TokenPermissions{
			testProductToken: {Maintainer: true, Download: true, Upload: true, Delete: true},
		},
		Versions: map[string]types.Version{},
	}
	p2 := &types.Product{
		Name:      "test-group-product",
		GroupName: "group-b",
		Tokens: map[string]types.TokenPermissions{
			testProductToken: {Maintainer: true, Download: true, Upload: true, Delete: true},
		},
		Versions: map[string]types.Version{},
	}

	err := repo.CreateProduct(p1)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-group-product", "group-a", testProductToken, true)

	err = repo.CreateProduct(p2)
	require.NoError(t, err)
	defer repo.DeleteProduct("test-group-product", "group-b", testProductToken, true)

	fetched1, err := repo.FetchProduct("test-group-product", "group-a")
	assert.NoError(t, err)
	assert.NotNil(t, fetched1)
	assert.Equal(t, "group-a", fetched1.GroupName)

	fetched2, err := repo.FetchProduct("test-group-product", "group-b")
	assert.NoError(t, err)
	assert.NotNil(t, fetched2)
	assert.Equal(t, "group-b", fetched2.GroupName)
}
