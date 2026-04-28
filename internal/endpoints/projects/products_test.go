package projects

import (
	"encoding/json"
	"net/http"
	"testing"

	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ownerProject(id int) *fakeProjectRepo {
	return &fakeProjectRepo{getByIDFn: func(int) (*types.Project, error) {
		return &types.Project{ID: id, Owner: 7}, nil
	}}
}

func fullPerm() *fakePermissionRepo {
	return &fakePermissionRepo{getFn: func(int, int) (*types.Permission, error) {
		return &types.Permission{CanDownload: true, CanUpload: true, CanDelete: true}, nil
	}}
}

func TestHandleListProducts_InvalidID(t *testing.T) {
	h := newHandler(&fakeUserRepo{}, &fakeProjectRepo{}, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/0/products", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "0"}}

	h.HandleListProducts(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleListProducts_Success(t *testing.T) {
	prod := &fakeProductRepo{listFn: func(projectID int) ([]types.Product, error) {
		return []types.Product{{ID: 1, Name: "spigot", Project: projectID}}, nil
	}}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), prod, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/1/products", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleListProducts(c)
	assert.Equal(t, http.StatusOK, w.Code)
	var out []productDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	assert.Len(t, out, 1)
	assert.Equal(t, "spigot", out[0].Name)
}

func TestHandleCreateProduct_BlankName(t *testing.T) {
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), &fakeProductRepo{}, &fakeVersionRepo{})
	body := []byte(`{"name":"   "}`)
	c, w := newCtx(t, http.MethodPost, "/projects/1/products", body)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleCreateProduct(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateProduct_PermissionDenied(t *testing.T) {
	noUpload := &fakePermissionRepo{getFn: func(int, int) (*types.Permission, error) {
		return &types.Permission{CanDownload: true}, nil
	}}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), noUpload, &fakeProductRepo{}, &fakeVersionRepo{})
	body := []byte(`{"name":"spigot"}`)
	c, w := newCtx(t, http.MethodPost, "/projects/1/products", body)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleCreateProduct(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleCreateProduct_Success(t *testing.T) {
	prod := &fakeProductRepo{createFn: func(projectID int, name string) (*types.Product, error) {
		return &types.Product{ID: 99, Name: name, Project: projectID}, nil
	}}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), prod, &fakeVersionRepo{})
	body := []byte(`{"name":"spigot"}`)
	c, w := newCtx(t, http.MethodPost, "/projects/1/products", body)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleCreateProduct(c)
	assert.Equal(t, http.StatusCreated, w.Code)
	var got productDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, 99, got.ID)
}

func TestHandleDeleteProduct_NotFound(t *testing.T) {
	prod := &fakeProductRepo{getByIDFn: func(int) (*types.Product, error) { return nil, nil }}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), prod, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodDelete, "/projects/1/products/9", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}, {Key: "product_id", Value: "9"}}

	h.HandleDeleteProduct(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleDeleteProduct_PermissionDenied(t *testing.T) {
	prod := &fakeProductRepo{getByIDFn: func(int) (*types.Product, error) {
		return &types.Product{ID: 9, Project: 1}, nil
	}}
	noDelete := &fakePermissionRepo{getFn: func(int, int) (*types.Permission, error) {
		return &types.Permission{CanDownload: true, CanUpload: true}, nil
	}}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), noDelete, prod, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodDelete, "/projects/1/products/9", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}, {Key: "product_id", Value: "9"}}

	h.HandleDeleteProduct(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleDeleteProduct_Success(t *testing.T) {
	prod := &fakeProductRepo{
		getByIDFn: func(int) (*types.Product, error) {
			return &types.Product{ID: 9, Project: 1}, nil
		},
		deleteFn: func(int) error { return nil },
	}
	versionsCalled := false
	ver := &fakeVersionRepo{
		listFn: func(int) ([]types.Version, error) {
			versionsCalled = true
			return []types.Version{}, nil
		},
		deleteFn: func(int) error { return nil },
	}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), prod, ver)
	c, w := newCtx(t, http.MethodDelete, "/projects/1/products/9", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}, {Key: "product_id", Value: "9"}}

	h.HandleDeleteProduct(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, versionsCalled)
}
