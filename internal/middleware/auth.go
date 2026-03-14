package middleware

import (
	"artifactor/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

const API_HEADER = "X-Api-Token"

func AuthMiddleware(repo repository.IAuthRepo) gin.HandlerFunc {
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

		admin, err := repo.IsAdmin(authHeader)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Set("admin", admin)
		c.Set("token", authHeader)
	}
}
