package projects

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleDeleteProject_NoAuth(t *testing.T) {
	h := newHandler(&fakeUserRepo{}, &fakeProjectRepo{}, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodDelete, "/projects/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleDeleteProject(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleDeleteProject_InvalidID(t *testing.T) {
	h := newHandler(&fakeUserRepo{}, &fakeProjectRepo{}, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodDelete, "/projects/abc", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "abc"}}

	h.HandleDeleteProject(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleDeleteProject_NotFound(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) { return nil, nil }}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodDelete, "/projects/1", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleDeleteProject(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleDeleteProject_NotOwner(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 99}, nil
	}}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodDelete, "/projects/1", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleDeleteProject(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleDeleteProject_RemovesBlobs(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.bin")
	b := filepath.Join(dir, "b.bin")
	require.NoError(t, os.WriteFile(a, []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(b, []byte("b"), 0o644))

	pr := &fakeProjectRepo{
		getByIDFn: func(id int) (*types.Project, error) {
			return &types.Project{ID: 1, Owner: 7}, nil
		},
		deleteFn: func(id int) ([]string, error) {
			return []string{a, b}, nil
		},
	}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodDelete, "/projects/1", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleDeleteProject(c)
	assert.Equal(t, http.StatusOK, w.Code)
	_, errA := os.Stat(a)
	_, errB := os.Stat(b)
	assert.True(t, os.IsNotExist(errA))
	assert.True(t, os.IsNotExist(errB))
}

func TestHandleDeleteProject_RepoError(t *testing.T) {
	pr := &fakeProjectRepo{
		getByIDFn: func(id int) (*types.Project, error) {
			return &types.Project{ID: 1, Owner: 7}, nil
		},
		deleteFn: func(id int) ([]string, error) { return nil, fmt.Errorf("boom") },
	}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodDelete, "/projects/1", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleDeleteProject(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleListImported_UnknownHost(t *testing.T) {
	defer withHosts(map[string]types.Host{})()
	h := newHandler(&fakeUserRepo{}, &fakeProjectRepo{}, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/user/projects", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))

	h.HandleListImported(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleImport_MissingFields(t *testing.T) {
	hostURL := "https://gitlab.example"
	defer withHosts(map[string]types.Host{hostURL: {Id: 1, Url: hostURL, Type: types.Gitlab}})()

	h := newHandler(&fakeUserRepo{}, &fakeProjectRepo{}, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	body := []byte(`{"source_id":0,"org":0}`)
	c, w := newCtx(t, http.MethodPost, "/user/projects", body)
	setAuthHeader(c, signSession(t, 7, hostURL, []int{10}))

	h.HandleImport(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleImport_OrgNotInSession(t *testing.T) {
	hostURL := "https://gitlab.example"
	defer withHosts(map[string]types.Host{hostURL: {Id: 1, Url: hostURL, Type: types.Gitlab}})()

	h := newHandler(&fakeUserRepo{}, &fakeProjectRepo{}, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	body := []byte(`{"source_id":99,"org":42}`)
	c, w := newCtx(t, http.MethodPost, "/user/projects", body)
	setAuthHeader(c, signSession(t, 7, hostURL, []int{10}))

	h.HandleImport(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleImport_UnknownUser(t *testing.T) {
	hostURL := "https://gitlab.example"
	defer withHosts(map[string]types.Host{hostURL: {Id: 1, Url: hostURL, Type: types.Gitlab}})()

	user := &fakeUserRepo{existsFn: func(int) (bool, error) { return false, nil }}
	h := newHandler(user, &fakeProjectRepo{}, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	body := []byte(`{"source_id":99,"org":10}`)
	c, w := newCtx(t, http.MethodPost, "/user/projects", body)
	setAuthHeader(c, signSession(t, 7, hostURL, []int{10}))

	h.HandleImport(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
