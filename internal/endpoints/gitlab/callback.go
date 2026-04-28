package gitlab

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"packster/internal"
	"packster/internal/requests"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const jwtStorageKey = "packster_jwt"

const callbackHTML = `<!doctype html>
<meta charset="utf-8">
<title>Signing in…</title>
<script>
  try { localStorage.setItem(%q, %q); } catch (e) {}
  location.replace("/");
</script>
`

// HandleCallback completes the OAuth flow: exchanges the code for a token,
// fetches the GitLab user and groups, upserts the user record, and renders an
// HTML page that stores a signed JWT in localStorage before redirecting to the
// SPA root.
func (h *GitlabHandler) HandleCallback(c *gin.Context) {
	stateParam := c.Query("state")
	state, err := strconv.Atoi(stateParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	var host *types.Host
	for _, v := range internal.Hosts {
		if v.Id == state && v.Type == types.Gitlab {
			host = &v
			break;
		}
	}

	if host == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("host %d not found", state)})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	redirectURI := fmt.Sprintf("%s://%s/api/auth/gitlab/callback", scheme, c.Request.Host)
	payload := map[string]string{
		"client_id":     host.ApplicationId,
		"client_secret": host.Secret,
		"code":          c.Query("code"),
		"grant_type":    "authorization_code",
		"redirect_uri":  redirectURI,
	}

	client := &http.Client{}
	res, err := requests.GitlabOauthToken(client, payload, host.Url)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	gitlabUser, err := requests.FetchGitlabUser(client, res.Token, host.Url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	orgs := make([]int, 0, len(gitlabUser.Groups))
	for _, g := range gitlabUser.Groups {
		orgs = append(orgs, g.ID)
	}

	user, err := h.Repo.CreateUser(gitlabUser.Username, host.Url, gitlabUser.ID, orgs)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	signed, err := h.signJwt(user, host, res.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, callbackHTML, jwtStorageKey, signed)
}

func (h *GitlabHandler) signJwt(user *types.User, host *types.Host, providerToken string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"token": providerToken,
		"host":  map[string]string{"type": host.Type.String(), "url": host.Url},
		"orgs":  user.Orgs,
		"sub":   strconv.Itoa(user.ID),
		"name":  user.DisplayName,
		"iat":   now.Unix(),
		"exp":   now.Add(24 * time.Hour).Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(h.Cfg.Secret))
}
