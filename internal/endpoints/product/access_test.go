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

func TestHandleAccess_ListProductsError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/product/access", nil)

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return(nil, errors.New("db error"))

	handler := &ProductHandler{Repo: repo}
	handler.HandleAccess(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	repo.AssertExpectations(t)
}

func TestHandleAccess_NoProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/product/access", nil)
	c.Set("token", "mytoken")

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return([]string{}, nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleAccess(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "null")
	repo.AssertExpectations(t)
}

func TestHandleAccess_TokenHasAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/product/access", nil)
	c.Set("token", "mytoken")

	product := productWithToken("mytoken", types.TokenPermissions{Download: true})
	product.Name = "productA"

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return([]string{"productA"}, nil)
	repo.On("FetchProduct", "productA").Return(product, nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleAccess(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "productA")
	repo.AssertExpectations(t)
}

func TestHandleAccess_TokenNoAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/product/access", nil)
	c.Set("token", "othertoken")

	product := productWithToken("mytoken", types.TokenPermissions{Download: true})
	product.Name = "productA"

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return([]string{"productA"}, nil)
	repo.On("FetchProduct", "productA").Return(product, nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleAccess(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "productA")
	repo.AssertExpectations(t)
}

func TestHandleAccess_PartialAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/product/access", nil)
	c.Set("token", "mytoken")

	productA := productWithToken("mytoken", types.TokenPermissions{Download: true})
	productA.Name = "productA"

	productB := &types.Product{
		Name:     "productB",
		Tokens:   map[string]types.TokenPermissions{},
		Versions: map[string]types.Version{},
	}

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return([]string{"productA", "productB"}, nil)
	repo.On("FetchProduct", "productA").Return(productA, nil)
	repo.On("FetchProduct", "productB").Return(productB, nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleAccess(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "productA")
	assert.NotContains(t, w.Body.String(), "productB")
	repo.AssertExpectations(t)
}

func TestHandleAccess_FetchProductError_Skipped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/product/access", nil)
	c.Set("token", "mytoken")

	productB := productWithToken("mytoken", types.TokenPermissions{Download: true})
	productB.Name = "productB"

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return([]string{"productA", "productB"}, nil)
	repo.On("FetchProduct", "productA").Return(nil, errors.New("db error"))
	repo.On("FetchProduct", "productB").Return(productB, nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleAccess(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "productA")
	assert.Contains(t, w.Body.String(), "productB")
	repo.AssertExpectations(t)
}
