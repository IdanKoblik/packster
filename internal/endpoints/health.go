package endpoints

import (
	"artifactor/internal/mongo"
	"artifactor/internal/redis"
	"artifactor/internal/repository"
	responses "artifactor/pkg/http"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) HandleHealth(c *gin.Context) {
	response := responses.HealthResponse{
		MongoStatus: "Mongo is fine",
		RedisStatus: "Redis is fine",
	}

	authRepository, ok := h.Repo.(*repository.AuthRepository)
	if !ok {
		c.String(http.StatusInternalServerError, "NOT GOOD PLEASE CONTACT PROJECT MAINTAINER")
		return
	}

	status := http.StatusOK
	err := mongo.CheckHealth(authRepository.MongoClient)
	if err != nil {
		response.MongoStatus = err.Error()
		status = http.StatusInternalServerError
	}

	err = redis.CheckHealth(authRepository.RedisClient)
	if err != nil {
		response.RedisStatus = err.Error()
		status = http.StatusInternalServerError
	}

	c.JSON(status, response)
}
