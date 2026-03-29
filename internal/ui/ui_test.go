package ui

import (
	"packster/pkg/types"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuthRepo struct {
	mock.Mock
}

func (m *mockAuthRepo) FetchToken(rawToken string) (*types.ApiToken, error) {
	args := m.Called(rawToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ApiToken), args.Error(1)
}

func (m *mockAuthRepo) IsAdmin(rawToken string) (bool, error) {
	args := m.Called(rawToken)
	return args.Bool(0), args.Error(1)
}

func TestNewUIHandler(t *testing.T) {
	repo := &mockAuthRepo{}
	h := NewUIHandler(repo)
	assert.NotNil(t, h)
	assert.Equal(t, repo, h.repo)
}

func TestHandleIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/ui", nil)

	h := NewUIHandler(&mockAuthRepo{})
	h.HandleIndex(c)

	// static/index.html exists in the embedded FS
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
}

func TestHandleStatic_EmptyPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/ui/", nil)
	c.Params = gin.Params{{Key: "filepath", Value: "/"}}

	h := NewUIHandler(&mockAuthRepo{})
	h.HandleStatic(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleStatic_ExistingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/ui/index.html", nil)
	c.Params = gin.Params{{Key: "filepath", Value: "/index.html"}}

	h := NewUIHandler(&mockAuthRepo{})
	h.HandleStatic(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleStatic_NotFound_FallsBackToIndex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/ui/nonexistent.js", nil)
	c.Params = gin.Params{{Key: "filepath", Value: "/nonexistent.js"}}

	h := NewUIHandler(&mockAuthRepo{})
	h.HandleStatic(c)

	// Falls back to index.html
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleLogin_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(map[string]string{})
	c.Request = httptest.NewRequest(http.MethodPost, "/ui/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := NewUIHandler(&mockAuthRepo{})
	h.HandleLogin(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "token is required")
}

func TestHandleLogin_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(map[string]string{"token": "badtoken"})
	c.Request = httptest.NewRequest(http.MethodPost, "/ui/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	repo := &mockAuthRepo{}
	repo.On("FetchToken", "badtoken").Return(nil, errors.New("not found"))

	h := NewUIHandler(repo)
	h.HandleLogin(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
	repo.AssertExpectations(t)
}

func TestHandleLogin_NonAdmin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(map[string]string{"token": "usertoken"})
	c.Request = httptest.NewRequest(http.MethodPost, "/ui/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	repo := &mockAuthRepo{}
	repo.On("FetchToken", "usertoken").Return(&types.ApiToken{Token: "hashed", Admin: false}, nil)

	h := NewUIHandler(repo)
	h.HandleLogin(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"admin":false`)
	repo.AssertExpectations(t)
}

func TestHandleLogin_Admin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(map[string]string{"token": "admintoken"})
	c.Request = httptest.NewRequest(http.MethodPost, "/ui/login", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	repo := &mockAuthRepo{}
	repo.On("FetchToken", "admintoken").Return(&types.ApiToken{Token: "hashed", Admin: true}, nil)

	h := NewUIHandler(repo)
	h.HandleLogin(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"admin":true`)
	repo.AssertExpectations(t)
}
