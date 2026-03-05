package endpoints

import (
	"net/http"
	"artifactor/pkg/requests"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gin-gonic/gin"
)

func (h *AuthHandler) HandleRegister(c *gin.Context, signingKey string) {
	var request requests.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.String(http.StatusBadRequest, "Missing request body")
		return
	}

	if request.Username == "" {
		c.String(http.StatusBadRequest, "Missing username field")
		return
	}

	if request.Password == "" {
		c.String(http.StatusBadRequest, "Missing password field")
		return
	}

	if request.Name == "" {
		c.String(http.StatusBadRequest, "Missing name field")
		return
	}

	exists, err := h.Repo.UserExists(request.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	if exists {
		c.String(http.StatusConflict, "This user already exists")
		return
	}

	err = h.Repo.CreateUser(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	key, err := generateJWT(request.Username, request.Password, signingKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.String(http.StatusCreated, key)
}

func generateJWT(username, password, signingKey string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"password": password,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
