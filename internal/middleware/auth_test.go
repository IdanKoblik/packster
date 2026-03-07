package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"artifactor/internal/repository"
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

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	mockRepo := new(MockAuthRepo)

	handler := AuthMiddleware(mockRepo)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header missing")
	assert.True(t, c.IsAborted())
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(API_HEADER, "invalid-token")

	mockRepo := new(MockAuthRepo)
	mockRepo.On("TokenExists", "invalid-token").Return(false, nil)

	handler := AuthMiddleware(mockRepo)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid api token")
	assert.True(t, c.IsAborted())
}

func TestAuthMiddleware_TokenExistsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(API_HEADER, "error-token")

	mockRepo := new(MockAuthRepo)
	mockRepo.On("TokenExists", "error-token").Return(false, errors.New("redis error"))

	handler := AuthMiddleware(mockRepo)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "redis error")
	assert.True(t, c.IsAborted())
}

func TestAuthMiddleware_Success_Admin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(API_HEADER, "valid-admin-token")

	mockRepo := new(MockAuthRepo)
	mockRepo.On("TokenExists", "valid-admin-token").Return(true, nil)
	mockRepo.On("IsAdmin", "valid-admin-token").Return(true, nil)

	handler := AuthMiddleware(mockRepo)
	handler(c)

	admin, exists := c.Get("admin")
	assert.True(t, exists)
	assert.True(t, admin.(bool))
	assert.False(t, c.IsAborted())
	mockRepo.AssertExpectations(t)
}

func TestAuthMiddleware_Success_NonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(API_HEADER, "valid-user-token")

	mockRepo := new(MockAuthRepo)
	mockRepo.On("TokenExists", "valid-user-token").Return(true, nil)
	mockRepo.On("IsAdmin", "valid-user-token").Return(false, nil)

	handler := AuthMiddleware(mockRepo)
	handler(c)

	admin, exists := c.Get("admin")
	assert.True(t, exists)
	assert.False(t, admin.(bool))
	assert.False(t, c.IsAborted())
	mockRepo.AssertExpectations(t)
}

func TestAuthMiddleware_IsAdminError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(API_HEADER, "valid-token")

	mockRepo := new(MockAuthRepo)
	mockRepo.On("TokenExists", "valid-token").Return(true, nil)
	mockRepo.On("IsAdmin", "valid-token").Return(false, errors.New("db error"))

	handler := AuthMiddleware(mockRepo)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	assert.True(t, c.IsAborted())
}

var _ repository.AuthRepoInterface = (*MockAuthRepo)(nil)
