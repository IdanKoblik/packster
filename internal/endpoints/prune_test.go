package endpoints

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandlePrune_Unauthorized_NoAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/tokens/sometoken", nil)

	handler := &AuthHandler{Repo: &mockRepo{}}
	handler.HandlePrune(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandlePrune_Unauthorized_NotAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/tokens/sometoken", nil)
	c.Set("admin", false)

	handler := &AuthHandler{Repo: &mockRepo{}}
	handler.HandlePrune(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandlePrune_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/tokens/", nil)
	c.Set("admin", true)

	handler := &AuthHandler{Repo: &mockRepo{}}
	handler.HandlePrune(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing api token")
}

func TestHandlePrune_RepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/tokens/sometoken", nil)
	c.Set("admin", true)
	c.Params = gin.Params{{Key: "token", Value: "sometoken"}}

	repo := &mockRepo{}
	repo.On("PruneToken", "sometoken").Return(errors.New("prune error"))

	handler := &AuthHandler{Repo: repo}
	handler.HandlePrune(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "prune error")
	repo.AssertExpectations(t)
}

func TestHandlePrune_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/tokens/sometoken", nil)
	c.Set("admin", true)
	c.Params = gin.Params{{Key: "token", Value: "sometoken"}}

	repo := &mockRepo{}
	repo.On("PruneToken", "sometoken").Return(nil)

	handler := &AuthHandler{Repo: repo}
	handler.HandlePrune(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertExpectations(t)
}
