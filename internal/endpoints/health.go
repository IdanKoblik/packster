package endpoints

import (
	internalmongo "packster/internal/mongo"
	internalredis "packster/internal/redis"
	responses "packster/pkg/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// HandleHealth godoc
// @Summary      Health check
// @Description  Returns the health status of MongoDB and Redis connections.
// @Tags         system
// @Produce      json
// @Success      200  {object}  types.HealthResponse  "All services healthy"
// @Failure      500  {object}  types.HealthResponse  "One or more services unhealthy"
// @Security     ApiKeyAuth
// @Router       /health [get]
func HandleHealth(c *gin.Context, mongo *mongo.Client, redis *redis.Client) {
	response := responses.HealthResponse{
		MongoStatus: "Mongo is fine",
		RedisStatus: "Redis is fine",
	}

	status := http.StatusOK
	err := internalmongo.CheckHealth(mongo)
	if err != nil {
		response.MongoStatus = err.Error()
		status = http.StatusInternalServerError
	}

	err = internalredis.CheckHealth(redis)
	if err != nil {
		response.RedisStatus = err.Error()
		status = http.StatusInternalServerError
	}

	c.JSON(status, response)
}
