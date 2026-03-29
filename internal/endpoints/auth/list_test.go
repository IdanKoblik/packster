package auth

import (
	"packster/pkg/types"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleListTokens_Unauthorized_NoAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/tokens", nil)

	handler := &AuthHandler{Repo: &mockRepo{}}
	handler.HandleListTokens(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleListTokens_Unauthorized_NotAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/tokens", nil)
	c.Set("admin", false)

	handler := &AuthHandler{Repo: &mockRepo{}}
	handler.HandleListTokens(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleListTokens_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/tokens", nil)
	c.Set("admin", true)

	repo := &mockRepo{}
	repo.On("ListTokens").Return(nil, errors.New("db error"))

	handler := &AuthHandler{Repo: repo}
	handler.HandleListTokens(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	repo.AssertExpectations(t)
}

func TestHandleListTokens_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/tokens", nil)
	c.Set("admin", true)

	tokens := []types.ApiToken{{Token: "hash1", Admin: true}, {Token: "hash2", Admin: false}}
	repo := &mockRepo{}
	repo.On("ListTokens").Return(tokens, nil)

	handler := &AuthHandler{Repo: repo}
	handler.HandleListTokens(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "hash1")
	assert.Contains(t, w.Body.String(), "hash2")
	repo.AssertExpectations(t)
}
