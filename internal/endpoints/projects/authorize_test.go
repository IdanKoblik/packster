package projects

import (
	"net/http"
	"testing"

	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthorize_Unauthenticated(t *testing.T) {
	h := newHandler(&fakeUserRepo{}, &fakeProjectRepo{}, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/x", nil)

	_, _, _, ok := h.authorize(c, 1, accessRead)
	assert.False(t, ok)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthorize_ProjectNotFound(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(int) (*types.Project, error) { return nil, nil }}
	h := newHandler(&fakeUserRepo{}, pr, &fakePermissionRepo{}, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/x", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))

	_, _, _, ok := h.authorize(c, 1, accessRead)
	assert.False(t, ok)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAuthorize_NoPermissionRow(t *testing.T) {
	pr := &fakeProjectRepo{getByIDFn: func(int) (*types.Project, error) {
		return &types.Project{ID: 1, Owner: 99}, nil
	}}
	perm := &fakePermissionRepo{getFn: func(int, int) (*types.Permission, error) { return nil, nil }}
	h := newHandler(&fakeUserRepo{}, pr, perm, &fakeProductRepo{}, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/x", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))

	_, _, _, ok := h.authorize(c, 1, accessRead)
	assert.False(t, ok)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthorize_AccessGating(t *testing.T) {
	cases := []struct {
		name    string
		perm    types.Permission
		kind    accessKind
		wantOK  bool
		wantStatus int
	}{
		{"read with download only", types.Permission{CanDownload: true}, accessRead, true, 0},
		{"read with upload only", types.Permission{CanUpload: true}, accessRead, true, 0},
		{"read with delete only", types.Permission{CanDelete: true}, accessRead, true, 0},
		{"read with no flags", types.Permission{}, accessRead, false, http.StatusForbidden},
		{"upload without upload flag", types.Permission{CanDownload: true}, accessUpload, false, http.StatusForbidden},
		{"upload with upload flag", types.Permission{CanUpload: true}, accessUpload, true, 0},
		{"delete without delete flag", types.Permission{CanUpload: true}, accessDelete, false, http.StatusForbidden},
		{"delete with delete flag", types.Permission{CanDelete: true}, accessDelete, true, 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			pr := &fakeProjectRepo{getByIDFn: func(int) (*types.Project, error) {
				return &types.Project{ID: 1, Owner: 99}, nil
			}}
			perm := &fakePermissionRepo{getFn: func(int, int) (*types.Permission, error) {
				p := tt.perm
				return &p, nil
			}}
			h := newHandler(&fakeUserRepo{}, pr, perm, &fakeProductRepo{}, &fakeVersionRepo{})
			c, w := newCtx(t, http.MethodGet, "/x", nil)
			setAuthHeader(c, signSession(t, 7, "https://h", nil))

			_, _, _, ok := h.authorize(c, 1, tt.kind)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantStatus != 0 {
				assert.Equal(t, tt.wantStatus, w.Code)
			}
		})
	}
}
