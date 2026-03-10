package endpoints

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"artifactor/internal/redis"
	"artifactor/internal/sql"
	httprequest "artifactor/pkg/http"
	"artifactor/pkg/tokens"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepo struct {
	mock.Mock
}

func (m *MockAuthRepo) TokenExists(rawToken string) (bool, error) {
	args := m.Called(rawToken)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthRepo) CreateToken(request *httprequest.CreateRequest) (string, error) {
	args := m.Called(request)
	return args.String(0), args.Error(1)
}

func (m *MockAuthRepo) PruneToken(rawToken string) error {
	args := m.Called(rawToken)
	return args.Error(0)
}

func (m *MockAuthRepo) IsAdmin(rawToken string) (bool, error) {
	args := m.Called(rawToken)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthRepo) FetchToken(rawToken string) (*tokens.Token, error) {
	args := m.Called(rawToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tokens.Token), args.Error(1)
}

func setupTestContext(isAdmin bool) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set("admin", isAdmin)
	return c, w
}

func TestHandleRegister_NotAdmin(t *testing.T) {
	c, w := setupTestContext(false)

	handler := &AuthHandler{}
	handler.HandleRegister(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Only admin allowed to register new tokens")
}

func TestHandleRegister_MissingBody(t *testing.T) {
	c, w := setupTestContext(true)
	c.Request = httptest.NewRequest(http.MethodPost, "/register", nil)
	c.Request.Header.Set("Content-Type", "application/json")

	mockRepo := new(MockAuthRepo)
	handler := &AuthHandler{Repo: mockRepo}
	handler.HandleRegister(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRegister_Success(t *testing.T) {
	c, w := setupTestContext(true)

	body := []byte(`{"admin": true, "upload": true, "delete": false}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	mockRepo := new(MockAuthRepo)
	mockRepo.On("CreateToken", mock.AnythingOfType("*http.CreateRequest")).Return("new-token-123", nil)

	handler := &AuthHandler{Repo: mockRepo}
	handler.HandleRegister(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "new-token-123", w.Body.String())
	mockRepo.AssertExpectations(t)
}

func TestHandleRegister_CreateTokenError(t *testing.T) {
	c, w := setupTestContext(true)

	body := []byte(`{"admin": true, "upload": true, "delete": false}`)
	c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	mockRepo := new(MockAuthRepo)
	mockRepo.On("CreateToken", mock.AnythingOfType("*http.CreateRequest")).Return("", errors.New("db error"))

	handler := &AuthHandler{Repo: mockRepo}
	handler.HandleRegister(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
}

func TestHandleFetch_NotAdmin(t *testing.T) {
	c, w := setupTestContext(false)

	handler := &AuthHandler{}
	handler.HandleFetch(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Only admin allowed to prune an api token")
}

func TestHandleFetch_MissingTokenParam(t *testing.T) {
	c, w := setupTestContext(true)
	c.Params = gin.Params{}

	handler := &AuthHandler{}
	handler.HandleFetch(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing api token")
}

func TestHandleFetch_Success(t *testing.T) {
	c, w := setupTestContext(true)
	c.Params = gin.Params{{Key: "token", Value: "test-token"}}

	mockToken := &tokens.Token{
		Data: "hashed-token",
		Permissions: tokens.TokenPermissions{
			Admin:  true,
			Upload: true,
			Delete: false,
		},
	}

	mockRepo := new(MockAuthRepo)
	mockRepo.On("FetchToken", "test-token").Return(mockToken, nil)

	handler := &AuthHandler{Repo: mockRepo}
	handler.HandleFetch(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response tokens.Token
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Permissions.Admin)
	assert.True(t, response.Permissions.Upload)
	assert.False(t, response.Permissions.Delete)
	mockRepo.AssertExpectations(t)
}

func TestHandleFetch_TokenNotFound(t *testing.T) {
	c, w := setupTestContext(true)
	c.Params = gin.Params{{Key: "token", Value: "nonexistent-token"}}

	mockRepo := new(MockAuthRepo)
	mockRepo.On("FetchToken", "nonexistent-token").Return(nil, errors.New("token not found"))

	handler := &AuthHandler{Repo: mockRepo}
	handler.HandleFetch(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "token not found")
}

func TestHandlePrune_NotAdmin(t *testing.T) {
	c, w := setupTestContext(false)

	handler := &AuthHandler{}
	handler.HandlePrune(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Only admin allowed to prune an api token")
}

func TestHandlePrune_MissingTokenParam(t *testing.T) {
	c, w := setupTestContext(true)
	c.Params = gin.Params{}

	handler := &AuthHandler{}
	handler.HandlePrune(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing api token")
}

func TestHandlePrune_Success(t *testing.T) {
	c, w := setupTestContext(true)
	c.Params = gin.Params{{Key: "token", Value: "test-token"}}

	mockRepo := new(MockAuthRepo)
	mockRepo.On("PruneToken", "test-token").Return(nil)

	handler := &AuthHandler{Repo: mockRepo}
	handler.HandlePrune(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestHandlePrune_Error(t *testing.T) {
	c, w := setupTestContext(true)
	c.Params = gin.Params{{Key: "token", Value: "test-token"}}

	mockRepo := new(MockAuthRepo)
	mockRepo.On("PruneToken", "test-token").Return(errors.New("failed to delete"))

	handler := &AuthHandler{Repo: mockRepo}
	handler.HandlePrune(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to delete")
}

func TestHandleHealth_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	if redis.Client == nil || sql.Conn == nil {
		t.Skip("skipping test: redis or sql not connected")
	}

	handler := &AuthHandler{}
	handler.HandleHealth(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "sql")
	assert.Contains(t, w.Body.String(), "redis")
}

func TestNewAuthHandler(t *testing.T) {
	handler := NewAuthHandler(nil)
	assert.NotNil(t, handler)
	assert.Nil(t, handler.Repo)
}
