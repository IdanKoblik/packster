package gitlab

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"packster/pkg/config"
	"packster/pkg/types"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUserRepo struct {
	createFn     func(username, host string, ssoId int, orgs []int) (*types.User, error)
	existsByIDFn func(id int) (bool, error)
	calls        []createCall
}

type createCall struct {
	Username string
	Host     string
	SsoID    int
	Orgs     []int
}

func (f *fakeUserRepo) CreateUser(username, host string, ssoId int, orgs []int) (*types.User, error) {
	f.calls = append(f.calls, createCall{username, host, ssoId, orgs})
	if f.createFn != nil {
		return f.createFn(username, host, ssoId, orgs)
	}
	return &types.User{ID: 1, Username: username, DisplayName: username, SsoID: ssoId, Host: host, Orgs: orgs}, nil
}

func (f *fakeUserRepo) UserExists(username string, ssoId, host int) (*types.User, error) {
	return nil, nil
}

func (f *fakeUserRepo) UserExistsByID(id int) (bool, error) {
	if f.existsByIDFn != nil {
		return f.existsByIDFn(id)
	}
	return true, nil
}

func (f *fakeUserRepo) PurgeUserData(userID int) ([]string, error) {
	return nil, nil
}

func (f *fakeUserRepo) SearchByName(hostID int, query string, excludeID int) ([]types.User, error) {
	return nil, nil
}

type gitlabServerOpts struct {
	tokenStatus   int
	tokenBody     string
	userStatus    int
	userBody      string
	groupsStatus  int
	groupsBody    string
}

func newGitlabMock(opts gitlabServerOpts) *httptest.Server {
	if opts.tokenStatus == 0 {
		opts.tokenStatus = http.StatusOK
	}
	if opts.userStatus == 0 {
		opts.userStatus = http.StatusOK
	}
	if opts.groupsStatus == 0 {
		opts.groupsStatus = http.StatusOK
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(opts.tokenStatus)
		fmt.Fprint(w, opts.tokenBody)
	})
	mux.HandleFunc("/api/v4/user", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(opts.userStatus)
		fmt.Fprint(w, opts.userBody)
	})
	mux.HandleFunc("/api/v4/groups", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(opts.groupsStatus)
		fmt.Fprint(w, opts.groupsBody)
	})
	return httptest.NewServer(mux)
}

func TestHandleCallback_InvalidState(t *testing.T) {
	defer withHosts(map[string]types.Host{})()
	h := &GitlabHandler{Repo: &fakeUserRepo{}}
	c, w := newCtx(http.MethodGet, "/callback?state=bad")

	h.HandleCallback(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCallback_HostNotFound(t *testing.T) {
	defer withHosts(map[string]types.Host{})()
	h := &GitlabHandler{Repo: &fakeUserRepo{}}
	c, w := newCtx(http.MethodGet, "/callback?state=5")

	h.HandleCallback(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleCallback_WrongHostType(t *testing.T) {
	defer withHosts(map[string]types.Host{
		"https://github.example.com": {Id: 5, Url: "https://github.example.com", Type: types.Github},
	})()
	h := &GitlabHandler{Repo: &fakeUserRepo{}}
	c, w := newCtx(http.MethodGet, "/callback?state=5")

	h.HandleCallback(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleCallback_Success(t *testing.T) {
	srv := newGitlabMock(gitlabServerOpts{
		tokenBody:  `{"access_token": "tok-123"}`,
		userBody:   `{"id": 99, "username": "alice"}`,
		groupsBody: `[{"id": 10, "name": "A"}, {"id": 20, "name": "B"}]`,
	})
	defer srv.Close()

	defer withHosts(map[string]types.Host{
		srv.URL: {
			Id:            5,
			Url:           srv.URL,
			Type:          types.Gitlab,
			ApplicationId: "app",
			Secret:        "sec",
			Orgs:          []int{10, 20},
		},
	})()

	repo := &fakeUserRepo{
		createFn: func(username, host string, ssoId int, orgs []int) (*types.User, error) {
			return &types.User{
				ID:          77,
				Username:    username,
				DisplayName: username,
				SsoID:       ssoId,
				Host:        host,
				Orgs:        orgs,
			}, nil
		},
	}
	h := &GitlabHandler{Cfg: config.Config{Secret: "test-secret"}, Repo: repo}
	c, w := newCtx(http.MethodGet, "/callback?state=5&code=abc")

	h.HandleCallback(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, w.Body.String(), `localStorage.setItem("packster_jwt"`)

	require.Len(t, repo.calls, 1)
	call := repo.calls[0]
	assert.Equal(t, "alice", call.Username)
	assert.Equal(t, srv.URL, call.Host)
	assert.Equal(t, 99, call.SsoID)
	assert.Equal(t, []int{10, 20}, call.Orgs)

	signed := extractJwtFromHTML(t, w.Body.String())
	tok, err := jwt.Parse(signed, func(t *jwt.Token) (any, error) {
		return []byte("test-secret"), nil
	})
	require.NoError(t, err)
	require.True(t, tok.Valid)
	claims := tok.Claims.(jwt.MapClaims)
	assert.Equal(t, "tok-123", claims["token"])
	assert.Equal(t, "alice", claims["name"])
	assert.Equal(t, "77", claims["sub"])
	assert.Equal(t, "gitlab", claims["host"].(map[string]any)["type"])
	gotOrgs := claims["orgs"].([]any)
	assert.Equal(t, []any{float64(10), float64(20)}, gotOrgs)
}

func extractJwtFromHTML(t *testing.T, body string) string {
	t.Helper()
	const prefix = `localStorage.setItem("packster_jwt", "`
	i := indexOf(body, prefix)
	require.GreaterOrEqual(t, i, 0, "jwt setItem not found in body")
	rest := body[i+len(prefix):]
	end := indexOf(rest, `"`)
	require.GreaterOrEqual(t, end, 0)
	return rest[:end]
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestHandleCallback_UserFetchFails(t *testing.T) {
	srv := newGitlabMock(gitlabServerOpts{
		tokenBody: `{"access_token": "tok-123"}`,
		userBody:  `{"id": 0}`,
	})
	defer srv.Close()

	defer withHosts(map[string]types.Host{
		srv.URL: {
			Id:   5,
			Url:  srv.URL,
			Type: types.Gitlab,
		},
	})()

	h := &GitlabHandler{Repo: &fakeUserRepo{}}
	c, w := newCtx(http.MethodGet, "/callback?state=5&code=abc")

	h.HandleCallback(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleCallback_RepoError(t *testing.T) {
	srv := newGitlabMock(gitlabServerOpts{
		tokenBody:  `{"access_token": "tok-123"}`,
		userBody:   `{"id": 99, "username": "alice"}`,
		groupsBody: `[{"id": 10}]`,
	})
	defer srv.Close()

	defer withHosts(map[string]types.Host{
		srv.URL: {Id: 5, Url: srv.URL, Type: types.Gitlab, Orgs: []int{10}},
	})()

	repo := &fakeUserRepo{
		createFn: func(username, host string, ssoId int, orgs []int) (*types.User, error) {
			return nil, fmt.Errorf("denied")
		},
	}
	h := &GitlabHandler{Repo: repo}
	c, w := newCtx(http.MethodGet, "/callback?state=5&code=abc")

	h.HandleCallback(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "denied")
}

func TestHandleCallback_GroupsError(t *testing.T) {
	srv := newGitlabMock(gitlabServerOpts{
		tokenBody:    `{"access_token": "tok-123"}`,
		userBody:     `{"id": 99, "username": "alice"}`,
		groupsStatus: http.StatusInternalServerError,
		groupsBody:   `boom`,
	})
	defer srv.Close()

	defer withHosts(map[string]types.Host{
		srv.URL: {Id: 5, Url: srv.URL, Type: types.Gitlab},
	})()

	h := &GitlabHandler{Repo: &fakeUserRepo{}}
	c, w := newCtx(http.MethodGet, "/callback?state=5&code=abc")

	h.HandleCallback(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
