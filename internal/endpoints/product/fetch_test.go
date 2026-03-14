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

func TestHandleFetch_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/products/myproduct", nil)
	c.Params = gin.Params{{Key: "product", Value: "myproduct"}}

	repo := &mockProductRepo{}
	repo.On("FetchProduct", "myproduct").Return(nil, errors.New("db error"))

	handler := &ProductHandler{Repo: repo}
	handler.HandleFetch(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	repo.AssertExpectations(t)
}

func TestHandleFetch_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/products/myproduct", nil)
	c.Params = gin.Params{{Key: "product", Value: "myproduct"}}

	expected := &types.Product{Name: "myproduct", Tokens: map[string]types.TokenPermissions{}, Versions: map[string]types.Version{}}
	repo := &mockProductRepo{}
	repo.On("FetchProduct", "myproduct").Return(expected, nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleFetch(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "myproduct")
	repo.AssertExpectations(t)
}
