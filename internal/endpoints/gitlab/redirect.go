package gitlab

import (
	"fmt"
	"net/url"
	"net/http"

	"packster/pkg/config"

	"github.com/gin-gonic/gin"
)

func (h *GitlabHandler) HandleRedirect(c *gin.Context) {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	url := redirectUrl(h.Cfg, scheme, c.Request.Host)
	c.Redirect(http.StatusFound, url)
}

func redirectUrl(cfg *config.Config, scheme, host string) string {
	baseURL := fmt.Sprintf("%s/oauth/authorize", cfg.Gitlab.Host)
	redirectURI := fmt.Sprintf("%s://%s/api/auth/gitlab/callback", scheme, host)

	params := url.Values{}
	params.Add("client_id", cfg.Gitlab.ApplicationId)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "read_user")

	return baseURL + "?" + params.Encode()
}
