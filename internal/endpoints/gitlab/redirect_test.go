package gitlab

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"packster/internal"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withHosts(hosts map[string]types.Host) func() {
	prev := internal.Hosts
	internal.Hosts = hosts
	return func() { internal.Hosts = prev }
}

func newCtx(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, nil)
	return c, w
}

func TestHandleRedirect_InvalidId(t *testing.T) {
	defer withHosts(map[string]types.Host{})()
	h := &GitlabHandler{}
	c, w := newCtx(http.MethodGet, "/redirect?id=notanumber")

	h.HandleRedirect(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRedirect_HostNotFound(t *testing.T) {
	defer withHosts(map[string]types.Host{})()
	h := &GitlabHandler{}
	c, w := newCtx(http.MethodGet, "/redirect?id=42")

	h.HandleRedirect(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Contains(t, body["error"], "42")
}

func TestHandleRedirect_WrongType(t *testing.T) {
	defer withHosts(map[string]types.Host{
		"https://github.example.com": {
			Id:   1,
			Url:  "https://github.example.com",
			Type: types.Github,
		},
	})()
	h := &GitlabHandler{}
	c, w := newCtx(http.MethodGet, "/redirect?id=1")

	h.HandleRedirect(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleRedirect_Success(t *testing.T) {
	hostURL := "https://gitlab.example.com"
	defer withHosts(map[string]types.Host{
		hostURL: {
			Id:            7,
			Url:           hostURL,
			Type:          types.Gitlab,
			ApplicationId: "app-id",
			Secret:        "secret",
		},
	})()

	h := &GitlabHandler{}
	c, w := newCtx(http.MethodGet, "/redirect?id=7")
	c.Request.Host = "packster.local"

	h.HandleRedirect(c)

	assert.Equal(t, http.StatusFound, w.Code)
	loc := w.Header().Get("Location")
	require.NotEmpty(t, loc)

	parsed, err := url.Parse(loc)
	require.NoError(t, err)
	assert.Equal(t, "gitlab.example.com", parsed.Host)
	assert.Equal(t, "/oauth/authorize", parsed.Path)

	q := parsed.Query()
	assert.Equal(t, "app-id", q.Get("client_id"))
	assert.Equal(t, "http://packster.local/api/auth/gitlab/callback", q.Get("redirect_uri"))
	assert.Equal(t, "code", q.Get("response_type"))
	assert.Equal(t, "read_api", q.Get("scope"))
	assert.Equal(t, "7", q.Get("state"))
}

func TestBuildRedirectUrl(t *testing.T) {
	host := &types.Host{
		Id:            9,
		Url:           "https://gitlab.example.com",
		ApplicationId: "abc",
	}

	got := buildRedirectUrl(host, "https", "packster.local")
	parsed, err := url.Parse(got)
	require.NoError(t, err)

	assert.Equal(t, "gitlab.example.com", parsed.Host)
	assert.Equal(t, "/oauth/authorize", parsed.Path)

	q := parsed.Query()
	assert.Equal(t, "abc", q.Get("client_id"))
	assert.Equal(t, "https://packster.local/api/auth/gitlab/callback", q.Get("redirect_uri"))
	assert.Equal(t, "code", q.Get("response_type"))
	assert.Equal(t, "read_api", q.Get("scope"))
	assert.Equal(t, "9", q.Get("state"))
}
