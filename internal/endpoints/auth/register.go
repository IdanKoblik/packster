package auth

import (
	requests "packster/pkg/types"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleRegister godoc
// @Summary      Register a new API token
// @Description  Creates and returns a new API token. Requires admin privileges.
// @Tags         auth
// @Accept       json
// @Produce      plain
// @Param        request  body      requests.RegisterRequest  true  "Registration options"
// @Success      201  {string}  string  "The newly created token string"
// @Failure      400  {object}  map[string]string  "Missing or invalid request body"
// @Failure      401  {string}  string  "Admin privileges required"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /register [put]
func (h *AuthHandler) HandleRegister(c *gin.Context) {
	admin, exists := c.Get("admin")
	if !exists || !admin.(bool) {
		c.String(http.StatusUnauthorized, "Only admin allowed to register new tokens")
		return
	}

	var request requests.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		msg := err.Error()
		if errors.Is(err, io.EOF) {
			msg = "Missing request body"
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": msg,
		})
		return
	}

	token, err := h.Repo.CreateToken(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.String(http.StatusCreated, token)
}
