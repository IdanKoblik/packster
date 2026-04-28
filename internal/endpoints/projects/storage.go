package projects

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func storeBlob(storageRoot string, projectID, productID int, fh *multipart.FileHeader) (string, string, error) {
	if storageRoot == "" {
		return "", "", fmt.Errorf("storage path is not configured")
	}

	src, err := fh.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	dir := filepath.Join(storageRoot, fmt.Sprintf("%d", projectID), fmt.Sprintf("%d", productID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}

	name := filepath.Base(fh.Filename)
	if name == "" || name == "." || name == "/" {
		return "", "", fmt.Errorf("invalid filename")
	}
	dst := filepath.Join(dir, strconv.FormatInt(time.Now().UnixNano(), 10)+"-"+name)

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", "", err
	}

	hasher := sha256.New()
	if _, err := io.Copy(io.MultiWriter(out, hasher), src); err != nil {
		out.Close()
		os.Remove(dst)
		return "", "", err
	}
	if err := out.Close(); err != nil {
		os.Remove(dst)
		return "", "", err
	}

	return dst, hex.EncodeToString(hasher.Sum(nil)), nil
}

func removeBlob(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}

func originalFilename(storedPath string) string {
	base := filepath.Base(storedPath)
	for i := 0; i < len(base); i++ {
		if base[i] == '-' {
			if _, err := strconv.ParseInt(base[:i], 10, 64); err == nil {
				return base[i+1:]
			}
			break
		}
	}
	return base
}
