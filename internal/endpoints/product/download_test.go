package product

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"packster/internal/metrics"
	"packster/internal/utils"
	"packster/pkg/types"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDownloadRequest(t *testing.T) *http.Request {
	t.Helper()
	return httptest.NewRequest(http.MethodGet, "/download/myproduct/1.0.0", nil)
}

func productWithTokenAndVersion(token string, perms types.TokenPermissions, version string, v types.Version) *types.Product {
	return &types.Product{
		Name:     "myproduct",
		Tokens:   map[string]types.TokenPermissions{utils.Hash(token): perms},
		Versions: map[string]types.Version{version: v},
	}
}

func TestHandleDownload(t *testing.T) {
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
			name:    "FetchProductError",
			product: "myproduct",
			version: "1.0.0",
			setupMock: func(repo *mockProductRepo, _ string) {
				repo.On("FetchProduct", "myproduct", "").Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "db error",
		},
		{
			name:    "ProductNotFound",
			product: "myproduct",
			version: "1.0.0",
			setupMock: func(repo *mockProductRepo, _ string) {
				repo.On("FetchProduct", "myproduct", "").Return(nil, nil)
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
				repo.On("FetchProduct", "myproduct", "").Return(
					productWithToken("mytoken", types.TokenPermissions{Download: false}), nil,
				)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
		{
			name:    "AdminBypass_NoDownloadPermission",
			product: "myproduct",
			version: "1.0.0",
			admin:   boolPtr(true),
			token:   "mytoken",
			setupMock: func(repo *mockProductRepo, _ string) {
				repo.On("FetchProduct", "myproduct", "").Return(
					productWithToken("mytoken", types.TokenPermissions{Download: false}), nil,
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
				repo.On("FetchProduct", "myproduct", "").Return(
					productWithToken("mytoken", types.TokenPermissions{Download: true}), nil,
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
					types.TokenPermissions{Download: true},
					"1.0.0",
					types.Version{Path: "/etc/passwd", Checksum: "abc"},
				)
				repo.On("FetchProduct", "myproduct", "").Return(p, nil)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "invalid file path",
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
					types.TokenPermissions{Download: true},
					"1.0.0",
					types.Version{Path: filePath, Checksum: "abc"},
				)
				repo.On("FetchProduct", "myproduct", "").Return(p, nil)
			},
			wantStatus: http.StatusOK,
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
			c.Request = newDownloadRequest(t)
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
			handler.HandleDownload(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestHandleDownload_SuccessIncrementsMetrics(t *testing.T) {
	base := t.TempDir()
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(base))
	defer os.Chdir(orig)

	fileDir := filepath.Join(productsBaseDir, "myproduct", "1.0.0")
	require.NoError(t, os.MkdirAll(fileDir, 0755))
	filePath := filepath.Join(fileDir, "artifact.zip")
	require.NoError(t, os.WriteFile(filePath, []byte("data"), 0644))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = newDownloadRequest(t)
	c.Params = gin.Params{
		{Key: "product", Value: "myproduct"},
		{Key: "version", Value: "1.0.0"},
	}
	c.Set("admin", true)
	c.Set("token", "mytoken")

	repo := &mockProductRepo{}
	repo.On("FetchProduct", "myproduct", "").Return(
		productWithTokenAndVersion("mytoken", types.TokenPermissions{Download: true}, "1.0.0", types.Version{Path: filePath, Checksum: "abc"}), nil,
	)

	before := testutil.ToFloat64(metrics.ArtifactDownloadsTotal.WithLabelValues("myproduct"))

	handler := &ProductHandler{Repo: repo}
	handler.HandleDownload(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, float64(1), testutil.ToFloat64(metrics.ArtifactDownloadsTotal.WithLabelValues("myproduct"))-before)
}

func TestHandleDownload_WithGroup(t *testing.T) {
	base := t.TempDir()
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(base))
	defer os.Chdir(orig)

	fileDir := filepath.Join(productsBaseDir, "staging", "myproduct", "1.0.0")
	require.NoError(t, os.MkdirAll(fileDir, 0755))
	filePath := filepath.Join(fileDir, "artifact.zip")
	require.NoError(t, os.WriteFile(filePath, []byte("data"), 0644))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/download/myproduct/1.0.0?group=staging", nil)
	c.Params = gin.Params{
		{Key: "product", Value: "myproduct"},
		{Key: "version", Value: "1.0.0"},
	}
	c.Set("admin", true)
	c.Set("token", "mytoken")

	repo := &mockProductRepo{}
	repo.On("FetchProduct", "myproduct", "staging").Return(
		productWithTokenAndVersion("mytoken", types.TokenPermissions{Download: true}, "1.0.0", types.Version{Path: filePath, Checksum: "abc"}), nil,
	)

	handler := &ProductHandler{Repo: repo}
	handler.HandleDownload(c)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}
