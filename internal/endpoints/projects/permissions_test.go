package projects

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleListPermissions_NoAuth(t *testing.T) {
	h := newHandler(&fakeUserRepo{}, &fakeProjectRepo{}, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/1/permissions", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleListPermissions(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleListPermissions_InvalidID(t *testing.T) {
	h := newHandler(&fakeUserRepo{}, &fakeProjectRepo{}, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/abc/permissions", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "abc"}}

	h.HandleListPermissions(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleListPermissions_ProjectNotFound(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) { return nil, nil }}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/1/permissions", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleListPermissions(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleListPermissions_NotOwner(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 99}, nil
	}}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/1/permissions", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleListPermissions(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleListPermissions_Success(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 7}, nil
	}}
	perm := &fakePermissionRepo{listFn: func(projectID int) ([]types.PermissionEntry, error) {
		return []types.PermissionEntry{
			{Permission: types.Permission{Account: 7, Project: 1, CanDownload: true, CanUpload: true, CanDelete: true}, DisplayName: "Owner"},
			{Permission: types.Permission{Account: 8, Project: 1, CanDownload: true}, DisplayName: "Bob"},
		}, nil
	}}

	h := newHandler(&fakeUserRepo{}, pr, perm, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/1/permissions", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleListPermissions(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var got []permissionEntryDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Len(t, got, 2)
	assert.Equal(t, 7, got[0].UserID)
	assert.True(t, got[0].IsOwner)
	assert.Equal(t, 8, got[1].UserID)
	assert.False(t, got[1].IsOwner)
}

func TestHandleListPermissions_RepoError(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 7}, nil
	}}
	perm := &fakePermissionRepo{listFn: func(int) ([]types.PermissionEntry, error) {
		return nil, fmt.Errorf("boom")
	}}
	h := newHandler(&fakeUserRepo{}, pr, perm, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/1/permissions", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleListPermissions(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleSetPermission_OwnerCannotModifySelf(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 7}, nil
	}}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	body := []byte(`{"user_id":7,"can_download":true}`)
	c, w := newCtx(t, http.MethodPut, "/projects/1/permissions", body)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleSetPermission(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "owner")
}

func TestHandleSetPermission_NonOwnerForbidden(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 99}, nil
	}}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	body := []byte(`{"user_id":8,"can_download":true}`)
	c, w := newCtx(t, http.MethodPut, "/projects/1/permissions", body)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleSetPermission(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleSetPermission_MissingUserID(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 7}, nil
	}}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	body := []byte(`{"user_id":0,"can_download":true}`)
	c, w := newCtx(t, http.MethodPut, "/projects/1/permissions", body)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleSetPermission(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "user_id is required")
}

func TestHandleSetPermission_UserNotFound(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 7}, nil
	}}
	user := &fakeUserRepo{existsFn: func(id int) (bool, error) { return false, nil }}
	h := newHandler(user, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	body := []byte(`{"user_id":42,"can_download":true}`)
	c, w := newCtx(t, http.MethodPut, "/projects/1/permissions", body)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleSetPermission(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleSetPermission_Success(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 7}, nil
	}}
	user := &fakeUserRepo{existsFn: func(id int) (bool, error) { return true, nil }}
	perm := &fakePermissionRepo{}
	h := newHandler(user, pr, perm, &fakeProductRepo{}, &fakeVersionRepo{})

	body := []byte(`{"user_id":42,"can_download":true,"can_upload":false,"can_delete":true}`)
	c, w := newCtx(t, http.MethodPut, "/projects/1/permissions", body)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleSetPermission(c)
	assert.Equal(t, http.StatusOK, w.Code)
	require.Len(t, perm.setCalls, 1)
	assert.Equal(t, 42, perm.setCalls[0].Account)
	assert.Equal(t, 1, perm.setCalls[0].Project)
	assert.True(t, perm.setCalls[0].CanDownload)
	assert.False(t, perm.setCalls[0].CanUpload)
	assert.True(t, perm.setCalls[0].CanDelete)
}

func TestHandleRevokePermission_Success(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 7}, nil
	}}
	perm := &fakePermissionRepo{}
	h := newHandler(&fakeUserRepo{}, pr, perm, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodDelete, "/projects/1/permissions/8", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}, {Key: "user_id", Value: "8"}}

	h.HandleRevokePermission(c)
	assert.Equal(t, http.StatusOK, w.Code)
	require.Len(t, perm.deleteCalls, 1)
	assert.Equal(t, [2]int{8, 1}, perm.deleteCalls[0])
}

func TestHandleRevokePermission_OwnerCannotRevokeSelf(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 7}, nil
	}}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodDelete, "/projects/1/permissions/7", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}, {Key: "user_id", Value: "7"}}

	h.HandleRevokePermission(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRevokePermission_InvalidUserID(t *testing.T) {
	h := newHandler(&fakeUserRepo{}, &fakeProjectRepo{}, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodDelete, "/projects/1/permissions/0", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}, {Key: "user_id", Value: "0"}}

	h.HandleRevokePermission(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleSearchUsers_Success(t *testing.T) {
	hostURL := "https://gitlab.example"
	defer withHosts(map[string]types.Host{hostURL: {Id: 1, Url: hostURL, Type: types.Gitlab}})()

	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 7}, nil
	}}
	user := &fakeUserRepo{searchFn: func(hostID int, q string, exclude int) ([]types.User, error) {
		assert.Equal(t, 1, hostID)
		assert.Equal(t, "Bo", q)
		assert.Equal(t, 7, exclude)
		return []types.User{{ID: 8, DisplayName: "Bob"}}, nil
	}}

	h := newHandler(user, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/1/permissions/candidates?q=Bo", nil)
	setAuthHeader(c, signSession(t, 7, hostURL, nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleSearchUsers(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var got []userCandidateDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, 8, got[0].ID)
	assert.Equal(t, "Bob", got[0].DisplayName)
}

func TestHandleSearchUsers_EmptyQuery(t *testing.T) {
	hostURL := "https://gitlab.example"
	defer withHosts(map[string]types.Host{hostURL: {Id: 1, Url: hostURL, Type: types.Gitlab}})()

	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 7}, nil
	}}
	called := false
	user := &fakeUserRepo{searchFn: func(int, string, int) ([]types.User, error) {
		called = true
		return nil, nil
	}}
	h := newHandler(user, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/1/permissions/candidates?q=", nil)
	setAuthHeader(c, signSession(t, 7, hostURL, nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleSearchUsers(c)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, called, "repo should not be called for empty query")
	assert.Equal(t, "[]", w.Body.String())
}

func TestHandleSearchUsers_NotOwner(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 99}, nil
	}}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/1/permissions/candidates?q=Bo", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleSearchUsers(c)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleSearchUsers_UnknownHost(t *testing.T) {
	defer withHosts(map[string]types.Host{})()

	pr := &fakeProjectRepo{getByIDFn: func(id int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 7}, nil
	}}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/projects/1/permissions/candidates?q=Bo", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	h.HandleSearchUsers(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
