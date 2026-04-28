package projects

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"packster/pkg/config"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleListVersions_ProductNotFound(t *testing.T) {
	prod := &fakeProductRepo{getByIDFn: func(int) (*types.Product, error) { return nil, nil }}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), prod, &fakeVersionRepo{})
	c, w := newCtx(t, http.MethodGet, "/products/9/versions", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "product_id", Value: "9"}}

	h.HandleListVersions(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleListVersions_Success(t *testing.T) {
	prod := &fakeProductRepo{getByIDFn: func(int) (*types.Product, error) {
		return &types.Product{ID: 9, Project: 1}, nil
	}}
	ver := &fakeVersionRepo{listFn: func(int) ([]types.Version, error) {
		return []types.Version{
			{ID: 1, Name: "1.0.0", Path: "/blobs/9/1700000000000000000-asset.bin", Checksum: "abc", Product: 9},
		}, nil
	}}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), prod, ver)
	c, w := newCtx(t, http.MethodGet, "/products/9/versions", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "product_id", Value: "9"}}

	h.HandleListVersions(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var got []versionDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "asset.bin", got[0].Path, "should strip the timestamp prefix")
}

func TestHandleUploadVersion_InvalidName(t *testing.T) {
	prod := &fakeProductRepo{getByIDFn: func(int) (*types.Product, error) {
		return &types.Product{ID: 9, Project: 1}, nil
	}}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), prod, &fakeVersionRepo{})
	c, w := uploadCtx(t, "9", "../bad", "asset.bin", []byte("x"))
	setAuthHeader(c, signSession(t, 7, "https://h", nil))

	h.HandleUploadVersion(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleUploadVersion_DuplicateName(t *testing.T) {
	prod := &fakeProductRepo{getByIDFn: func(int) (*types.Product, error) {
		return &types.Product{ID: 9, Project: 1}, nil
	}}
	ver := &fakeVersionRepo{getByNameFn: func(int, string) (*types.Version, error) {
		return &types.Version{ID: 11, Name: "1.0.0"}, nil
	}}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), prod, ver)
	c, w := uploadCtx(t, "9", "1.0.0", "asset.bin", []byte("x"))
	setAuthHeader(c, signSession(t, 7, "https://h", nil))

	h.HandleUploadVersion(c)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandleUploadVersion_Success(t *testing.T) {
	root := t.TempDir()
	prod := &fakeProductRepo{getByIDFn: func(int) (*types.Product, error) {
		return &types.Product{ID: 9, Project: 1}, nil
	}}
	ver := &fakeVersionRepo{
		getByNameFn: func(int, string) (*types.Version, error) { return nil, nil },
		createFn: func(productID int, name, path, checksum string) (*types.Version, error) {
			return &types.Version{ID: 50, Name: name, Path: path, Checksum: checksum, Product: productID}, nil
		},
	}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), prod, ver)
	h.Cfg = config.Config{Secret: testSecret, Storage: config.StorageConfig{Path: root}}
	c, w := uploadCtx(t, "9", "1.0.0", "asset.bin", []byte("hello"))
	setAuthHeader(c, signSession(t, 7, "https://h", nil))

	h.HandleUploadVersion(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	var got versionDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, "1.0.0", got.Name)
	assert.Equal(t, "asset.bin", got.Path)
	assert.NotEmpty(t, got.Checksum)
}

func TestHandleUploadVersion_ExceedsLimit(t *testing.T) {
	prod := &fakeProductRepo{getByIDFn: func(int) (*types.Product, error) {
		return &types.Product{ID: 9, Project: 1}, nil
	}}
	ver := &fakeVersionRepo{getByNameFn: func(int, string) (*types.Version, error) { return nil, nil }}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), prod, ver)
	h.Cfg = config.Config{Secret: testSecret, FileUploadLimit: 0} // 0 disables the check
	// re-set with 1 MB limit but body smaller — sanity that successful path works
	h.Cfg.FileUploadLimit = 1
	big := bytes.Repeat([]byte("x"), 2*1024*1024)
	c, w := uploadCtx(t, "9", "1.0.0", "asset.bin", big)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))

	h.HandleUploadVersion(c)
	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
}

func TestHandleDeleteVersion_NotFound(t *testing.T) {
	ver := &fakeVersionRepo{getByIDFn: func(int) (*types.Version, error) { return nil, nil }}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), &fakeProductRepo{}, ver)
	c, w := newCtx(t, http.MethodDelete, "/versions/50", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "version_id", Value: "50"}}

	h.HandleDeleteVersion(c)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleDeleteVersion_Success(t *testing.T) {
	dir := t.TempDir()
	blobPath := filepath.Join(dir, "blob.bin")
	require.NoError(t, os.WriteFile(blobPath, []byte("x"), 0o644))

	prod := &fakeProductRepo{getByIDFn: func(int) (*types.Product, error) {
		return &types.Product{ID: 9, Project: 1}, nil
	}}
	ver := &fakeVersionRepo{
		getByIDFn: func(int) (*types.Version, error) {
			return &types.Version{ID: 50, Path: blobPath, Product: 9}, nil
		},
		deleteFn: func(int) error { return nil },
	}
	h := newHandler(&fakeUserRepo{}, ownerProject(1), fullPerm(), prod, ver)
	c, w := newCtx(t, http.MethodDelete, "/versions/50", nil)
	setAuthHeader(c, signSession(t, 7, "https://h", nil))
	c.Params = []gin.Param{{Key: "version_id", Value: "50"}}

	h.HandleDeleteVersion(c)
	assert.Equal(t, http.StatusOK, w.Code)
	_, err := os.Stat(blobPath)
	assert.True(t, os.IsNotExist(err))
}

func uploadCtx(t *testing.T, productID, versionName, filename string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	require.NoError(t, mw.WriteField("name", versionName))
	fw, err := mw.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = fw.Write(body)
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	c.Request = httptest.NewRequest(http.MethodPost, "/products/"+productID+"/versions", buf)
	c.Request.Header.Set("Content-Type", mw.FormDataContentType())
	c.Params = []gin.Param{{Key: "product_id", Value: productID}}
	return c, w
}
