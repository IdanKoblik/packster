package middleware

import (
	"net/http"
	"artifactor/internal/repository"

	"github.com/gin-gonic/gin"
)

const API_HEADER = "X-Api-Token"

func AuthMiddleware(repo *repository.AuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(API_HEADER)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header missing",
			})
			return
		}

		token, err := repo.FetchToken(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		if token == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid api token",
			})
			return
		}

		c.Set("admin", token.Permissions.Admin)
	}
}
