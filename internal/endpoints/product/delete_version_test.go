package product

import (
	"packster/pkg/types"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDeleteVersionRequest(t *testing.T, product, version string) *http.Request {
	t.Helper()
	return httptest.NewRequest(http.MethodDelete, "/product/"+product+"/"+version, nil)
}

func TestHandleDeleteVersion(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }

	tests := []struct {
		name          string
		product       string
		version       string
		admin         *bool
		token         string
		setupMock     func(*mockProductRepo, string)
		wantStatus    int
		wantBody      string
		needsTempFile bool
	}{
		{
			name:       "MissingProduct",
			product:    "",
			version:    "1.0.0",
			wantStatus: http.StatusBadRequest,
			wantBody:   "product required",
		},
		{
			name:       "MissingVersion",
			product:    "myproduct",
			version:    "",
			wantStatus: http.StatusBadRequest,
			wantBody:   "version required",
		},
		{
			name:    "FetchProductError",
			product: "myproduct",
			version: "1.0.0",
			setupMock: func(repo *mockProductRepo, _ string) {
				repo.On("FetchProduct", "myproduct").Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "db error",
		},
		{
			name:    "ProductNotFound",
			product: "myproduct",
			version: "1.0.0",
			setupMock: func(repo *mockProductRepo, _ string) {
				repo.On("FetchProduct", "myproduct").Return(nil, nil)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Product not found",
		},
		{
			name:    "PermissionDenied_NotAdmin",
			product: "myproduct",
			version: "1.0.0",
			admin:   boolPtr(false),
			token:   "mytoken",
			setupMock: func(repo *mockProductRepo, _ string) {
				repo.On("FetchProduct", "myproduct").Return(
					productWithToken("mytoken", types.TokenPermissions{Delete: false}), nil,
				)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
		{
			name:    "AdminBypass_NoDeletePermission",
			product: "myproduct",
			version: "1.0.0",
			admin:   boolPtr(true),
			token:   "mytoken",
			setupMock: func(repo *mockProductRepo, _ string) {
				repo.On("FetchProduct", "myproduct").Return(
					productWithToken("mytoken", types.TokenPermissions{Delete: false}), nil,
				)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Version not found",
		},
		{
			name:    "VersionNotFound",
			product: "myproduct",
			version: "1.0.0",
			admin:   boolPtr(true),
			token:   "mytoken",
			setupMock: func(repo *mockProductRepo, _ string) {
				repo.On("FetchProduct", "myproduct").Return(
					productWithToken("mytoken", types.TokenPermissions{Delete: true}), nil,
				)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Version not found",
		},
		{
			name:    "PathTraversalBlocked",
			product: "myproduct",
			version: "1.0.0",
			admin:   boolPtr(true),
			token:   "mytoken",
			setupMock: func(repo *mockProductRepo, _ string) {
				p := productWithTokenAndVersion(
					"mytoken",
					types.TokenPermissions{Delete: true},
					"1.0.0",
					types.Version{Path: "/etc/passwd", Checksum: "abc"},
				)
				repo.On("FetchProduct", "myproduct").Return(p, nil)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "invalid file path",
		},
		{
			name:          "RepoError",
			product:       "myproduct",
			version:       "1.0.0",
			admin:         boolPtr(true),
			token:         "mytoken",
			needsTempFile: true,
			setupMock: func(repo *mockProductRepo, filePath string) {
				p := productWithTokenAndVersion(
					"mytoken",
					types.TokenPermissions{Delete: true},
					"1.0.0",
					types.Version{Path: filePath, Checksum: "abc"},
				)
				repo.On("FetchProduct", "myproduct").Return(p, nil)
				repo.On("DeleteVersion", "myproduct", "1.0.0", "mytoken", true).Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "db error",
		},
		{
			name:          "Success",
			product:       "myproduct",
			version:       "1.0.0",
			admin:         boolPtr(true),
			token:         "mytoken",
			needsTempFile: true,
			setupMock: func(repo *mockProductRepo, filePath string) {
				p := productWithTokenAndVersion(
					"mytoken",
					types.TokenPermissions{Delete: true},
					"1.0.0",
					types.Version{Path: filePath, Checksum: "abc"},
				)
				repo.On("FetchProduct", "myproduct").Return(p, nil)
				repo.On("DeleteVersion", "myproduct", "1.0.0", "mytoken", true).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.needsTempFile {
				base := t.TempDir()
				orig, err := os.Getwd()
				require.NoError(t, err)
				require.NoError(t, os.Chdir(base))
				defer os.Chdir(orig)

				fileDir := filepath.Join(productsBaseDir, "myproduct", "1.0.0")
				require.NoError(t, os.MkdirAll(fileDir, 0755))
				filePath = filepath.Join(fileDir, "artifact.zip")
				require.NoError(t, os.WriteFile(filePath, []byte("file content"), 0644))
			}

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = newDeleteVersionRequest(t, tt.product, tt.version)
			c.Params = gin.Params{
				{Key: "product", Value: tt.product},
				{Key: "version", Value: tt.version},
			}
			if tt.admin != nil {
				c.Set("admin", *tt.admin)
			}
			if tt.token != "" {
				c.Set("token", tt.token)
			}

			repo := &mockProductRepo{}
			if tt.setupMock != nil {
				tt.setupMock(repo, filePath)
			}

			handler := &ProductHandler{Repo: repo}
			handler.HandleDeleteVersion(c)

			assert.Equal(t, tt.wantStatus, c.Writer.Status())
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			if tt.name == "Success" && filePath != "" {
				_, err := os.Stat(filePath)
				assert.True(t, os.IsNotExist(err), "file should be deleted from disk")
			}
			repo.AssertExpectations(t)
		})
	}
}
