package endpoints

import (
	"net/http"

	"packster/pkg/types"

	"github.com/gin-gonic/gin"
)

// HandleHosts returns the configured hosts as a JSON list of {id, type, url}
// objects. Used by the login screen to render the available providers.
func HandleHosts(c *gin.Context, hosts map[string]types.Host) {
	list := make([]gin.H, 0, len(hosts))
	for url, data := range hosts {
		list = append(list, gin.H{
			"id":   data.Id,
			"type": data.Type.String(),
			"url":  url,
		})
	}
	c.JSON(http.StatusOK, list)
}
