package gitlab

import (
	"fmt"
	"bytes"
	"io"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *GitlabHandler) HandleCallback(c *gin.Context) {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	redirectURI := fmt.Sprintf("%s://%s/api/auth/gitlab/callback", scheme, c.Request.Host)
	payload := map[string]string{
		"client_id":     h.Cfg.Gitlab.ApplicationId,
		"client_secret": h.Cfg.Gitlab.Secret,
		"code":          c.Query("code"),
		"grant_type":    "authorization_code",
		"redirect_uri":  redirectURI,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req, err := http.NewRequest("POST", h.Cfg.Gitlab.Host+"/oauth/token", bytes.NewBuffer(body))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var res interface{}
	_ = json.Unmarshal(respBody, &res)
	c.JSON(http.StatusOK, res)
}
