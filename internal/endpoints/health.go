package endpoints

import (
	"database/sql"
	"net/http"
	internalmysql "packster/internal/mysql"
	internalredis "packster/internal/redis"
	responses "packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HandleHealth godoc
// @Summary      Health check
// @Description  Returns the health status of MySQL and Redis connections.
// @Tags         system
// @Produce      json
// @Success      200  {object}  types.HealthResponse  "All services healthy"
// @Failure      500  {object}  types.HealthResponse  "One or more services unhealthy"
// @Security     ApiKeyAuth
// @Router       /health [get]
func HandleHealth(c *gin.Context, db *sql.DB, redis *redis.Client) {
	response := responses.HealthResponse{
		MySQLStatus: "MySQL is fine",
		RedisStatus: "Redis is fine",
	}

	status := http.StatusOK
	err := internalmysql.CheckHealth(db)
	if err != nil {
		response.MySQLStatus = err.Error()
		status = http.StatusInternalServerError
	}

	err = internalredis.CheckHealth(redis)
	if err != nil {
		response.RedisStatus = err.Error()
		status = http.StatusInternalServerError
	}

	c.JSON(status, response)
}
