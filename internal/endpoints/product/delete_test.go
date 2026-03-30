package product

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleDelete_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/products/myproduct", nil)
	c.Params = gin.Params{{Key: "product", Value: "myproduct"}}
	c.Set("token", "mytoken")
	c.Set("admin", false)

	repo := &mockProductRepo{}
	repo.On("DeleteProduct", "myproduct", "", "mytoken", false).Return(errors.New("db error"))

	handler := &ProductHandler{Repo: repo}
	handler.HandleDelete(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	repo.AssertExpectations(t)
}

func TestHandleDelete_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/products/myproduct", nil)
	c.Params = gin.Params{{Key: "product", Value: "myproduct"}}
	c.Set("token", "mytoken")
	c.Set("admin", true)

	repo := &mockProductRepo{}
	repo.On("DeleteProduct", "myproduct", "", "mytoken", true).Return(nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleDelete(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	repo.AssertExpectations(t)
}

func TestHandleDelete_WithGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/products/myproduct?group=staging", nil)
	c.Params = gin.Params{{Key: "product", Value: "myproduct"}}
	c.Set("token", "mytoken")
	c.Set("admin", true)

	repo := &mockProductRepo{}
	repo.On("DeleteProduct", "myproduct", "staging", "mytoken", true).Return(nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleDelete(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	repo.AssertExpectations(t)
}
