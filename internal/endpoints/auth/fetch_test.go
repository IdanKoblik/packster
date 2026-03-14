package auth

import (
	"artifactor/pkg/types"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleFetch_Unauthorized_NoAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/tokens/sometoken", nil)

	handler := &AuthHandler{Repo: &mockRepo{}}
	handler.HandleFetch(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleFetch_Unauthorized_NotAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/tokens/sometoken", nil)
	c.Set("admin", false)

	handler := &AuthHandler{Repo: &mockRepo{}}
	handler.HandleFetch(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleFetch_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/tokens/", nil)
	c.Set("admin", true)

	handler := &AuthHandler{Repo: &mockRepo{}}
	handler.HandleFetch(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing api token")
}

func TestHandleFetch_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/tokens/sometoken", nil)
	c.Set("admin", true)
	c.Params = gin.Params{{Key: "token", Value: "sometoken"}}

	repo := &mockRepo{}
	repo.On("FetchToken", "sometoken").Return(nil, errors.New("fetch error"))

	handler := &AuthHandler{Repo: repo}
	handler.HandleFetch(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "fetch error")
	repo.AssertExpectations(t)
}

func TestHandleFetch_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/tokens/sometoken", nil)
	c.Set("admin", true)
	c.Params = gin.Params{{Key: "token", Value: "sometoken"}}

	expectedToken := &types.ApiToken{Token: "hashedtoken", Admin: true}
	repo := &mockRepo{}
	repo.On("FetchToken", "sometoken").Return(expectedToken, nil)

	handler := &AuthHandler{Repo: repo}
	handler.HandleFetch(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "hashedtoken")
	repo.AssertExpectations(t)
}
