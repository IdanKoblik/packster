package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeFileHeader(t *testing.T, filename, content string) *multipart.FileHeader {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write([]byte(content))
	require.NoError(t, err)
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, "/", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(32 << 20)
	require.NoError(t, err)

	fh := req.MultipartForm.File["file"][0]
	return fh
}

func TestChecksum(t *testing.T) {
	t.Run("known content produces correct SHA256", func(t *testing.T) {
		content := "hello world"
		fh := makeFileHeader(t, "test.txt", content)

		expected := sha256.Sum256([]byte(content))
		expectedHex := hex.EncodeToString(expected[:])

		result, err := Checksum(fh)
		require.NoError(t, err)
		assert.Equal(t, expectedHex, result)
	})

	t.Run("empty file produces correct SHA256", func(t *testing.T) {
		fh := makeFileHeader(t, "empty.txt", "")

		expected := sha256.Sum256([]byte(""))
		expectedHex := hex.EncodeToString(expected[:])

		result, err := Checksum(fh)
		require.NoError(t, err)
		assert.Equal(t, expectedHex, result)
	})

	t.Run("different content produces different checksums", func(t *testing.T) {
		fh1 := makeFileHeader(t, "a.txt", "content A")
		fh2 := makeFileHeader(t, "b.txt", "content B")

		sum1, err := Checksum(fh1)
		require.NoError(t, err)
		sum2, err := Checksum(fh2)
		require.NoError(t, err)

		assert.NotEqual(t, sum1, sum2)
	})

	t.Run("checksum is idempotent", func(t *testing.T) {
		fh := makeFileHeader(t, "test.txt", "some data")

		sum1, err := Checksum(fh)
		require.NoError(t, err)
		sum2, err := Checksum(fh)
		require.NoError(t, err)

		assert.Equal(t, sum1, sum2)
	})
}
