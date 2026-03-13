package endpoints

import (
	"artifactor/internal/helpers"
	"artifactor/internal/repository"
	"net/http"
	"net/http/httptest"
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

	handler := &AuthHandler{
		Repo: repo,
	}

	handler.HandleHealth(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "mongo")
	assert.Contains(t, w.Body.String(), "redis")
}

func TestHandleHealth_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	handler := &AuthHandler{
		Repo: &repository.AuthRepository{
			MongoClient: nil,
			RedisClient: nil,
		},
	}

	handler.HandleHealth(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Missing mongo client")
	assert.Contains(t, w.Body.String(), "Missing redis client")
}
