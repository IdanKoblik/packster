package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
)

func Checksum(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
	}
	defer src.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, src); err != nil {
		return "", fmt.Errorf("error hashing file: %w", err)
	}

	checksum := hex.EncodeToString(hasher.Sum(nil))
	return checksum, nil
}
