package product

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"artifactor/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleCreate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString("invalid json"))

	handler := &ProductHandler{Repo: &mockProductRepo{}}
	handler.HandleCreate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreate_MissingName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(types.CreateProductRequest{})
	c.Request = httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))

	handler := &ProductHandler{Repo: &mockProductRepo{}}
	handler.HandleCreate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "name is required")
}

func TestHandleCreate_InvalidName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(types.CreateProductRequest{Name: "../../etc"})
	c.Request = httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))

	handler := &ProductHandler{Repo: &mockProductRepo{}}
	handler.HandleCreate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid name")
}

func TestHandleCreate_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(types.CreateProductRequest{Name: "myproduct"})
	c.Request = httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	c.Set("token", "mytoken")

	repo := &mockProductRepo{}
	repo.On("CreateProduct", mock.Anything).Return(errors.New("db error"))

	handler := &ProductHandler{Repo: repo}
	handler.HandleCreate(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	repo.AssertExpectations(t)
}

func TestHandleCreate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(types.CreateProductRequest{Name: "myproduct"})
	c.Request = httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	c.Set("token", "mytoken")

	repo := &mockProductRepo{}
	repo.On("CreateProduct", mock.Anything).Return(nil)

	handler := &ProductHandler{Repo: repo}
	handler.HandleCreate(c)

	assert.Equal(t, http.StatusCreated, c.Writer.Status())
	repo.AssertExpectations(t)
}
