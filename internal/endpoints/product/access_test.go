package product

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newAccessContext(token string) (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/product/access", nil)
	if token != "" {
		c.Set("token", token)
	}
	return w, c
}

func runAccess(c *gin.Context, repo *mockProductRepo) {
	(&ProductHandler{Repo: repo}).HandleAccess(c)
}

func TestHandleAccess_ListProductsError(t *testing.T) {
	w, c := newAccessContext("")

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return(nil, errors.New("db error"))

	runAccess(c, repo)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	repo.AssertExpectations(t)
}

func TestHandleAccess_NoProducts(t *testing.T) {
	w, c := newAccessContext("mytoken")

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return([]types.Product{}, nil)

	runAccess(c, repo)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "null")
	repo.AssertExpectations(t)
}

func TestHandleAccess_TokenHasAccess(t *testing.T) {
	w, c := newAccessContext("mytoken")

	product := productWithToken("mytoken", types.TokenPermissions{Download: true})
	product.Name = "productA"

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return([]types.Product{{Name: "productA", GroupName: ""}}, nil)
	repo.On("FetchProduct", "productA", "").Return(product, nil)

	runAccess(c, repo)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "productA")
	repo.AssertExpectations(t)
}

func TestHandleAccess_TokenNoAccess(t *testing.T) {
	w, c := newAccessContext("othertoken")

	product := productWithToken("mytoken", types.TokenPermissions{Download: true})
	product.Name = "productA"

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return([]types.Product{{Name: "productA", GroupName: ""}}, nil)
	repo.On("FetchProduct", "productA", "").Return(product, nil)

	runAccess(c, repo)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "productA")
	repo.AssertExpectations(t)
}

func TestHandleAccess_PartialAccess(t *testing.T) {
	w, c := newAccessContext("mytoken")

	productA := productWithToken("mytoken", types.TokenPermissions{Download: true})
	productA.Name = "productA"

	productB := &types.Product{
		Name:     "productB",
		Tokens:   map[string]types.TokenPermissions{},
		Versions: map[string]types.Version{},
	}

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return([]types.Product{
		{Name: "productA", GroupName: ""},
		{Name: "productB", GroupName: ""},
	}, nil)
	repo.On("FetchProduct", "productA", "").Return(productA, nil)
	repo.On("FetchProduct", "productB", "").Return(productB, nil)

	runAccess(c, repo)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "productA")
	assert.NotContains(t, w.Body.String(), "productB")
	repo.AssertExpectations(t)
}

func TestHandleAccess_FetchProductError_Skipped(t *testing.T) {
	w, c := newAccessContext("mytoken")

	productB := productWithToken("mytoken", types.TokenPermissions{Download: true})
	productB.Name = "productB"

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return([]types.Product{
		{Name: "productA", GroupName: ""},
		{Name: "productB", GroupName: ""},
	}, nil)
	repo.On("FetchProduct", "productA", "").Return(nil, errors.New("db error"))
	repo.On("FetchProduct", "productB", "").Return(productB, nil)

	runAccess(c, repo)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "productA")
	assert.Contains(t, w.Body.String(), "productB")
	repo.AssertExpectations(t)
}

func TestHandleAccess_GroupedProducts(t *testing.T) {
	w, c := newAccessContext("mytoken")

	productStaging := productWithToken("mytoken", types.TokenPermissions{Download: true})
	productStaging.Name = "myapp"
	productStaging.GroupName = "staging"

	productProd := &types.Product{
		Name:      "myapp",
		GroupName: "production",
		Tokens:    map[string]types.TokenPermissions{},
		Versions:  map[string]types.Version{},
	}

	repo := &mockProductRepo{}
	repo.On("ListProducts").Return([]types.Product{
		{Name: "myapp", GroupName: "staging"},
		{Name: "myapp", GroupName: "production"},
	}, nil)
	repo.On("FetchProduct", "myapp", "staging").Return(productStaging, nil)
	repo.On("FetchProduct", "myapp", "production").Return(productProd, nil)

	runAccess(c, repo)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "staging")
	assert.NotContains(t, w.Body.String(), "production")
	repo.AssertExpectations(t)
}
