package middleware

import (
	"packster/internal/metrics"
	"packster/pkg/types"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"packster/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
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

func (m *MockAuthRepo) CreateToken(request *types.RegisterRequest) (string, error) {
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

func (m *MockAuthRepo) FetchToken(rawToken string) (*types.ApiToken, error) {
	args := m.Called(rawToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ApiToken), args.Error(1)
}

func (m *MockAuthRepo) ListTokens() ([]types.ApiToken, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.ApiToken), args.Error(1)
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
	mockRepo.On("FetchToken", "invalid-token").Return(nil, nil)

	handler := AuthMiddleware(mockRepo)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid api token")
	assert.True(t, c.IsAborted())
	mockRepo.AssertExpectations(t)
}

func TestAuthMiddleware_FetchTokenError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(API_HEADER, "error-token")

	mockRepo := new(MockAuthRepo)
	mockRepo.On("FetchToken", "error-token").Return(nil, errors.New("db error"))

	handler := AuthMiddleware(mockRepo)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	assert.True(t, c.IsAborted())
	mockRepo.AssertExpectations(t)
}

func TestAuthMiddleware_Success_Admin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(API_HEADER, "valid-admin-token")

	mockRepo := new(MockAuthRepo)
	mockRepo.On("FetchToken", "valid-admin-token").Return(&types.ApiToken{Token: "hashed", Admin: true}, nil)
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
	mockRepo.On("FetchToken", "valid-user-token").Return(&types.ApiToken{Token: "hashed", Admin: false}, nil)
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
	mockRepo.On("FetchToken", "valid-token").Return(&types.ApiToken{Token: "hashed"}, nil)
	mockRepo.On("IsAdmin", "valid-token").Return(false, errors.New("db error"))

	handler := AuthMiddleware(mockRepo)
	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	assert.True(t, c.IsAborted())
	mockRepo.AssertExpectations(t)
}

var _ repository.IAuthRepo = (*MockAuthRepo)(nil)

func TestAuthMiddleware_MissingHeader_IncrementsCounter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	before := testutil.ToFloat64(metrics.AuthFailures.WithLabelValues("missing_header"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	AuthMiddleware(new(MockAuthRepo))(c)

	after := testutil.ToFloat64(metrics.AuthFailures.WithLabelValues("missing_header"))
	assert.Equal(t, float64(1), after-before)
}

func TestAuthMiddleware_InvalidToken_IncrementsCounter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	before := testutil.ToFloat64(metrics.AuthFailures.WithLabelValues("invalid_token"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(API_HEADER, "invalid-token")

	mockRepo := new(MockAuthRepo)
	mockRepo.On("FetchToken", "invalid-token").Return(nil, nil)

	AuthMiddleware(mockRepo)(c)

	after := testutil.ToFloat64(metrics.AuthFailures.WithLabelValues("invalid_token"))
	assert.Equal(t, float64(1), after-before)
}

func TestAuthMiddleware_FetchTokenError_IncrementsCounter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	before := testutil.ToFloat64(metrics.AuthFailures.WithLabelValues("fetch_error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request.Header.Set(API_HEADER, "bad-token")

	mockRepo := new(MockAuthRepo)
	mockRepo.On("FetchToken", "bad-token").Return(nil, errors.New("db error"))

	AuthMiddleware(mockRepo)(c)

	after := testutil.ToFloat64(metrics.AuthFailures.WithLabelValues("fetch_error"))
	assert.Equal(t, float64(1), after-before)
}
