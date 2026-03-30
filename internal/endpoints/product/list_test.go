package product

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleListProducts_Admin_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/product/list", nil)
	c.Set("admin", true)

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return(nil, errors.New("db error"))

	handler := &ProductHandler{Repo: repo}
	handler.HandleListProducts(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	repo.AssertExpectations(t)
}

func TestHandleListProducts_Admin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/product/list", nil)
	c.Set("admin", true)

	products := []types.Product{{Name: "productA"}, {Name: "productB"}}
	repo := &mockProductRepo{}
	repo.On("ListProducts").Return(products, nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleListProducts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "productA")
	assert.Contains(t, w.Body.String(), "productB")
	repo.AssertExpectations(t)
}

func TestHandleListProducts_NonAdmin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/product/list", nil)
	c.Set("admin", false)
	c.Set("token", "mytoken")

	products := []types.Product{{Name: "productA"}}
	repo := &mockProductRepo{}
	repo.On("ListProductsByToken", mock.AnythingOfType("string")).
		Return(products, nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleListProducts(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "productA")
	repo.AssertExpectations(t)
}

func TestHandleListProducts_NonAdmin_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/product/list", nil)
	c.Set("admin", false)
	c.Set("token", "mytoken")

	repo := &mockProductRepo{}
	repo.On("ListProductsByToken", mock.AnythingOfType("string")).
		Return(nil, errors.New("db error"))

	handler := &ProductHandler{Repo: repo}
	handler.HandleListProducts(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	repo.AssertExpectations(t)
}
