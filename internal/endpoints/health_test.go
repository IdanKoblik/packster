package endpoints

import (
	"path/filepath"
	"net/http"
	"net/http/httptest"
	"testing"

	"packster/internal/sql"
	"packster/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleHealth_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	cfg, err := config.ParseConfig(filepath.Join("..", "..", "fixtures", "example.yml"))
	assert.NoError(t, err)

	pgsql, err := sql.OpenPgsqlConnection(&cfg.Sql)
	assert.NoError(t, err)

	HandleHealth(c, pgsql)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pgsql")
}

func TestHandleHealth_Nil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	HandleHealth(c, nil)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Pgsql instance was not found")
}

func TestHandleHealth_Failure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	cfg, err := config.ParseConfig(filepath.Join("..", "..", "fixtures", "example.yml"))
	assert.NoError(t, err)

	pgsql, err := sql.OpenPgsqlConnection(&cfg.Sql)
	assert.NoError(t, err)

	err = pgsql.Close()
	assert.NoError(t, err)

	HandleHealth(c, pgsql)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
