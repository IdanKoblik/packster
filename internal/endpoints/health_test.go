package endpoints

import (
	"net/http"
	"net/http/httptest"
	"packster/internal/helpers"
	"packster/internal/repository"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleHealth_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	repo, cleanup := helpers.SetupRepo(t)
	defer cleanup()

	HandleHealth(c, repo.DB, repo.RedisClient)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "mysql")
	assert.Contains(t, w.Body.String(), "redis")
}

func TestHandleHealth_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	repo := repository.AuthRepository{
		DB:          nil,
		RedisClient: nil,
	}

	HandleHealth(c, repo.DB, repo.RedisClient)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Missing mysql client")
	assert.Contains(t, w.Body.String(), "Missing redis client")
}
