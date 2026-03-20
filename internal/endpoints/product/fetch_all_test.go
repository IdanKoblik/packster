package product

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"artifactor/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleFetchAll_NonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/products", nil)
	c.Set("admin", false)

	repo := &mockProductRepo{}
	handler := &ProductHandler{Repo: repo}
	handler.HandleFetchAll(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "admin access required")
}

func TestHandleFetchAll_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/products", nil)
	c.Set("admin", true)

	repo := &mockProductRepo{}
	repo.On("FetchAllProducts").Return(nil, errors.New("db error"))

	handler := &ProductHandler{Repo: repo}
	handler.HandleFetchAll(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	repo.AssertExpectations(t)
}

func TestHandleFetchAll_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/products", nil)
	c.Set("admin", true)

	expected := []*types.Product{
		{Name: "product1", Tokens: map[string]types.TokenPermissions{}, Versions: map[string]types.Version{}},
		{Name: "product2", Tokens: map[string]types.TokenPermissions{}, Versions: map[string]types.Version{}},
	}
	repo := &mockProductRepo{}
	repo.On("FetchAllProducts").Return(expected, nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleFetchAll(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "product1")
	assert.Contains(t, w.Body.String(), "product2")
	repo.AssertExpectations(t)
}

func TestHandleFetchAll_EmptyList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/products", nil)
	c.Set("admin", true)

	repo := &mockProductRepo{}
	repo.On("FetchAllProducts").Return([]*types.Product{}, nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleFetchAll(c)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}
