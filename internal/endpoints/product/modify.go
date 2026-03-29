package product

import (
	"packster/pkg/types"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type action int

const (
	deleteVersion action = iota
	deleteToken
	addToken
	unknown
)

var actions = map[string]action{
	"deleteVersion": deleteVersion,
	"addToken":      addToken,
	"deleteToken":   deleteToken,
}

// HandleModify godoc
// @Summary      Modify a product
// @Description  Dispatches to one of three sub-actions based on the {action} path parameter.
// @Description
// @Description  **deleteVersion** (DELETE): Remove a version from a product.
// @Description  Body: `{"product": "...", "version": "..."}`
// @Description
// @Description  **deleteToken** (DELETE): Revoke a token's access to a product.
// @Description  Body: `{"product": "...", "token": "..."}`
// @Description
// @Description  **addToken** (PUT): Grant a token access to a product with specified permissions.
// @Description  Body: `{"product": "...", "token": "...", "permissions": {...}}`
// @Tags         products
// @Accept       json
// @Param        action  path  string  true  "Action to perform"  Enums(deleteVersion, deleteToken, addToken)
// @Param        body    body  object  true  "Action-specific payload (see description)"
// @Success      200  "Version deleted (deleteVersion)"
// @Success      201  "Token added (addToken)"
// @Success      204  "Token deleted (deleteToken)"
// @Failure      400  {object}  map[string]string  "Invalid action or request body"
// @Failure      405  {string}  string  "Method not allowed"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /product/modify/{action} [delete]
// @Router       /product/modify/{action} [put]
func (h *ProductHandler) HandleModify(c *gin.Context) {
	actionStr := c.Param("action")
	action, err := parseAction(actionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	switch action {
	case deleteVersion:
		handleDeleteVersion(c, h)
	case deleteToken:
		handleDeleteToken(c, h)
	case addToken:
		handleAddToken(c, h)
	default:
		c.String(http.StatusBadRequest, "Invalid action")
	}
}

func handleDeleteToken(c *gin.Context, h *ProductHandler) {
	var request types.DeleteProductTokenRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	err = h.Repo.DeleteToken(request.Product, c.GetString("token"), request.Token, c.GetBool("admin"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.Status(http.StatusNoContent)
}

func handleAddToken(c *gin.Context, h *ProductHandler) {
	if c.Request.Method != http.MethodPut {
		c.String(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	var request types.AddProductTokenRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	err = h.Repo.AddToken(request.Product, c.GetString("token"), request.Token, request.Permissions, c.GetBool("admin"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.Status(http.StatusCreated)
}

func handleDeleteVersion(c *gin.Context, h *ProductHandler) {
	var request types.DeleteVersionRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	err = h.Repo.DeleteVersion(
		request.Product,
		request.Version,
		c.GetString("token"),
		c.GetBool("admin"),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.Status(http.StatusOK)
}

func parseAction(action string) (action, error) {
	v, ok := actions[action]
	if !ok {
		return unknown, fmt.Errorf("unknown action: %s", action)
	}

	return v, nil
}
