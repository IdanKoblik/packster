package endpoints

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newHostsContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/hosts", nil)
	return c, w
}

func TestHandleHosts_Empty(t *testing.T) {
	c, w := newHostsContext()

	HandleHosts(c, map[string]types.Host{})

	assert.Equal(t, http.StatusOK, w.Code)

	var body []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Empty(t, body)
}

func TestHandleHosts_Nil(t *testing.T) {
	c, w := newHostsContext()

	HandleHosts(c, nil)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

func TestHandleHosts_SingleGitlabHost(t *testing.T) {
	c, w := newHostsContext()

	url := "https://gitlab.example.com"
	hosts := map[string]types.Host{
		url: {
			Id:            42,
			Url:           url,
			Type:          types.Gitlab,
			ApplicationId: "app-id",
			Secret:        "secret",
			Orgs:          []int{1, 2, 3},
		},
	}

	HandleHosts(c, hosts)

	assert.Equal(t, http.StatusOK, w.Code)

	var body []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	require.Len(t, body, 1)

	got := body[0]
	assert.Equal(t, float64(42), got["id"])
	assert.Equal(t, url, got["url"])
	assert.Equal(t, "gitlab", got["type"])
	// Secret must never leak over the wire
	assert.NotContains(t, got, "secret")
	assert.NotContains(t, got, "ApplicationId")
}

func TestHandleHosts_MultipleHosts(t *testing.T) {
	c, w := newHostsContext()

	hosts := map[string]types.Host{
		"https://gitlab.example.com": {
			Id:            1,
			Url:           "https://gitlab.example.com",
			Type:          types.Gitlab,
			ApplicationId: "gitlab-app",
			Secret:        "gitlab-secret",
			Orgs:          []int{1},
		},
		"https://github.example.com": {
			Id:            2,
			Url:           "https://github.example.com",
			Type:          types.Github,
			ApplicationId: "github-app",
			Secret:        "github-secret",
			Orgs:          []int{2, 3},
		},
	}

	HandleHosts(c, hosts)

	assert.Equal(t, http.StatusOK, w.Code)

	var body []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	require.Len(t, body, 2)

	byURL := map[string]map[string]any{}
	for _, item := range body {
		byURL[item["url"].(string)] = item
	}

	require.Contains(t, byURL, "https://gitlab.example.com")
	require.Contains(t, byURL, "https://github.example.com")
	assert.Equal(t, "gitlab", byURL["https://gitlab.example.com"]["type"])
	assert.Equal(t, "github", byURL["https://github.example.com"]["type"])
	assert.Equal(t, float64(1), byURL["https://gitlab.example.com"]["id"])
	assert.Equal(t, float64(2), byURL["https://github.example.com"]["id"])
}

func TestHandleHosts_ContentTypeJSON(t *testing.T) {
	c, w := newHostsContext()

	HandleHosts(c, map[string]types.Host{})

	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}
