package gitlab

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"packster/internal"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
)

// HandleRedirect issues a 302 to the GitLab OAuth authorize endpoint for the
// host identified by the `id` query parameter. The host id is round-tripped
// back as the OAuth `state` so the callback can resolve which host the code
// was issued for.
func (h *GitlabHandler) HandleRedirect(c *gin.Context) {
	idParam := c.Query("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	var host *types.Host
	for _, v := range internal.Hosts {
		if v.Id == id && v.Type == types.Gitlab {
			host = &v
			break;
		}
	}

	if host == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("host %d not found", id)})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	redirectURL := buildRedirectUrl(host, scheme, c.Request.Host)
	c.Redirect(http.StatusFound, redirectURL)
}

func buildRedirectUrl(host *types.Host, scheme, reqHost string) string {
	baseURL := fmt.Sprintf("%s/oauth/authorize", host.Url)
	redirectURI := fmt.Sprintf("%s://%s/api/auth/gitlab/callback", scheme, reqHost)

	params := url.Values{}
	params.Add("client_id", host.ApplicationId)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "read_api")
	params.Add("state", strconv.Itoa(host.Id))

	return baseURL + "?" + params.Encode()
}
