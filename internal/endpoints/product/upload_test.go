package product

import (
	"artifactor/internal/utils"
	"artifactor/pkg/types"
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newUploadRequest(t *testing.T, product, version, filename, content string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if product != "" {
		require.NoError(t, writer.WriteField("product", product))
	}
	if version != "" {
		require.NoError(t, writer.WriteField("version", version))
	}
	if filename != "" {
		part, err := writer.CreateFormFile("file", filename)
		require.NoError(t, err)
		_, err = part.Write([]byte(content))
		require.NoError(t, err)
	}

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func productWithToken(token string, perms types.TokenPermissions) *types.Product {
	return &types.Product{
		Name:     "myproduct",
		Tokens:   map[string]types.TokenPermissions{utils.Hash(token): perms},
		Versions: map[string]types.Version{},
	}
}

func TestHandleUpload(t *testing.T) {
	boolPtr := func(b bool) *bool { return &b }

	tests := []struct {
		name         string
		product      string
		version      string
		filename     string
		content      string
		admin        *bool
		token        string
		setupMock    func(*mockProductRepo)
		wantStatus   int
		wantBody     string
		needsTempDir bool
	}{
		{
			name:       "InvalidForm",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "FetchProductError",
			product:  "myproduct",
			version:  "1.0.0",
			filename: "artifact.zip",
			content:  "data",
			setupMock: func(repo *mockProductRepo) {
				repo.On("FetchProduct", "myproduct").Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "db error",
		},
		{
			name:     "ProductNotFound",
			product:  "myproduct",
			version:  "1.0.0",
			filename: "artifact.zip",
			content:  "data",
			setupMock: func(repo *mockProductRepo) {
				repo.On("FetchProduct", "myproduct").Return(nil, nil)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "Product not found",
		},
		{
			name:     "PermissionDenied_NotAdmin",
			product:  "myproduct",
			version:  "1.0.0",
			filename: "artifact.zip",
			content:  "data",
			admin:    boolPtr(false),
			token:    "mytoken",
			setupMock: func(repo *mockProductRepo) {
				repo.On("FetchProduct", "myproduct").Return(
					productWithToken("mytoken", types.TokenPermissions{Upload: true}), nil,
				)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
		{
			name:     "PermissionDenied_NoUploadPermission",
			product:  "myproduct",
			version:  "1.0.0",
			filename: "artifact.zip",
			content:  "data",
			admin:    boolPtr(true),
			token:    "mytoken",
			setupMock: func(repo *mockProductRepo) {
				repo.On("FetchProduct", "myproduct").Return(
					productWithToken("mytoken", types.TokenPermissions{Upload: false}), nil,
				)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "permission denied",
		},
		{
			name:     "VersionAlreadyExists",
			product:  "myproduct",
			version:  "1.0.0",
			filename: "artifact.zip",
			content:  "data",
			admin:    boolPtr(true),
			token:    "mytoken",
			setupMock: func(repo *mockProductRepo) {
				p := productWithToken("mytoken", types.TokenPermissions{Upload: true})
				p.Versions["1.0.0"] = types.Version{Path: "/some/path", Checksum: "abc"}
				repo.On("FetchProduct", "myproduct").Return(p, nil)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "already uploaded",
		},
		{
			name:         "AddVersionError",
			product:      "myproduct",
			version:      "1.0.0",
			filename:     "artifact.zip",
			content:      "data",
			admin:        boolPtr(true),
			token:        "mytoken",
			needsTempDir: true,
			setupMock: func(repo *mockProductRepo) {
				repo.On("FetchProduct", "myproduct").Return(
					productWithToken("mytoken", types.TokenPermissions{Upload: true}), nil,
				)
				repo.On("AddVersion", "myproduct", "1.0.0", "mytoken", true, mock.Anything).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "db error",
		},
		{
			name:         "Success",
			product:      "myproduct",
			version:      "1.0.0",
			filename:     "artifact.zip",
			content:      "data",
			admin:        boolPtr(true),
			token:        "mytoken",
			needsTempDir: true,
			setupMock: func(repo *mockProductRepo) {
				repo.On("FetchProduct", "myproduct").Return(
					productWithToken("mytoken", types.TokenPermissions{Upload: true}), nil,
				)
				repo.On("AddVersion", "myproduct", "1.0.0", "mytoken", true, mock.Anything).Return(nil)
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.needsTempDir {
				dir := t.TempDir()
				orig, err := os.Getwd()
				require.NoError(t, err)
				require.NoError(t, os.Chdir(dir))
				defer os.Chdir(orig)
			}

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = newUploadRequest(t, tt.product, tt.version, tt.filename, tt.content)
			if tt.admin != nil {
				c.Set("admin", *tt.admin)
			}
			if tt.token != "" {
				c.Set("token", tt.token)
			}

			repo := &mockProductRepo{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			handler := &ProductHandler{Repo: repo}
			handler.HandleUpload(c)

			assert.Equal(t, tt.wantStatus, c.Writer.Status())
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
			repo.AssertExpectations(t)
		})
	}
}
