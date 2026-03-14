package endpoints

import (
	internalmongo "artifactor/internal/mongo"
	internalredis "artifactor/internal/redis"
	responses "artifactor/pkg/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

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
