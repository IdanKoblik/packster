package auth

import (
	pkghttp "packster/pkg/types"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleRegister_Unauthorized_NoAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/tokens", nil)

	handler := &AuthHandler{Repo: &mockRepo{}}
	handler.HandleRegister(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleRegister_Unauthorized_NotAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/tokens", nil)
	c.Set("admin", false)

	handler := &AuthHandler{Repo: &mockRepo{}}
	handler.HandleRegister(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleRegister_MissingBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/tokens", nil)
	c.Set("admin", true)

	handler := &AuthHandler{Repo: &mockRepo{}}
	handler.HandleRegister(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing request body")
}

func TestHandleRegister_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()

	request := pkghttp.RegisterRequest{Admin: false}
	body, _ := json.Marshal(request)
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("admin", true)

	repo := &mockRepo{}
	repo.On("CreateToken", &request).Return("", errors.New("create error"))

	handler := &AuthHandler{Repo: repo}
	handler.HandleRegister(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "create error")
	repo.AssertExpectations(t)
}

func TestHandleRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()

	request := pkghttp.RegisterRequest{Admin: true}
	body, _ := json.Marshal(request)
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("admin", true)

	repo := &mockRepo{}
	repo.On("CreateToken", &request).Return("new-api-token", nil)

	handler := &AuthHandler{Repo: repo}
	handler.HandleRegister(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "new-api-token")
	repo.AssertExpectations(t)
}
