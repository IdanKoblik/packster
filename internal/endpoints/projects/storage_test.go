package projects

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeFileHeader(t *testing.T, name string, body []byte) *multipart.FileHeader {
	t.Helper()
	buf := &bytes.Buffer{}
	mw := multipart.NewWriter(buf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+name+`"`)
	w, err := mw.CreatePart(h)
	require.NoError(t, err)
	_, err = w.Write(body)
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	mr := multipart.NewReader(buf, mw.Boundary())
	form, err := mr.ReadForm(int64(len(body)) + 1024)
	require.NoError(t, err)
	require.Len(t, form.File["file"], 1)
	return form.File["file"][0]
}

func TestStoreBlob_Success(t *testing.T) {
	root := t.TempDir()
	contents := []byte("hello world")
	fh := makeFileHeader(t, "asset.bin", contents)

	stored, checksum, err := storeBlob(root, 1, 2, fh)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(stored, filepath.Join(root, "1", "2")))
	assert.True(t, strings.HasSuffix(stored, "-asset.bin"))

	want := sha256.Sum256(contents)
	assert.Equal(t, hex.EncodeToString(want[:]), checksum)

	got, err := os.ReadFile(stored)
	require.NoError(t, err)
	assert.Equal(t, contents, got)
}

func TestStoreBlob_NoStorageRoot(t *testing.T) {
	fh := makeFileHeader(t, "asset.bin", []byte("x"))
	_, _, err := storeBlob("", 1, 2, fh)
	assert.Error(t, err)
}

func TestStoreBlob_InvalidFilename(t *testing.T) {
	root := t.TempDir()
	fh := makeFileHeader(t, ".", []byte("x"))
	_, _, err := storeBlob(root, 1, 2, fh)
	assert.Error(t, err)
}

func TestRemoveBlob_NoOpOnEmpty(t *testing.T) {
	removeBlob("")
}

func TestRemoveBlob_RemovesFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x")
	require.NoError(t, os.WriteFile(p, []byte("x"), 0o644))
	removeBlob(p)
	_, err := os.Stat(p)
	assert.True(t, os.IsNotExist(err))
}

func TestRemoveBlob_MissingFileIgnored(t *testing.T) {
	removeBlob(filepath.Join(t.TempDir(), "does-not-exist"))
}

func TestOriginalFilename_StripsPrefix(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"/var/blobs/1/2/1700000000000000000-foo.tar.gz", "foo.tar.gz"},
		{"1700000000000000000-foo.tar.gz", "foo.tar.gz"},
		{"foo.tar.gz", "foo.tar.gz"},
		{"not-a-timestamp-foo.tar.gz", "not-a-timestamp-foo.tar.gz"},
	}
	for _, tt := range cases {
		assert.Equal(t, tt.want, originalFilename(tt.in))
	}
}

func TestValidateVersionName(t *testing.T) {
	bad := []string{"", "a/b", "..", ".", `a\b`}
	for _, n := range bad {
		assert.Error(t, validateVersionName(n), "expected error for %q", n)
	}
	good := []string{"1.0.0", "v1", "release-2024"}
	for _, n := range good {
		assert.NoError(t, validateVersionName(n), "expected no error for %q", n)
	}
}
