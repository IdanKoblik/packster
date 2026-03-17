package utils

import (
	"fmt"
	"path/filepath"
	"regexp"
)

var safeNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._\-]*$`)

func ValidateName(s string) error {
	if !safeNameRe.MatchString(s) {
		return fmt.Errorf("invalid name %q: must start with alphanumeric and contain only letters, digits, '.', '-', '_'", s)
	}
	return nil
}

func SafeFilename(filename string) (string, error) {
	base := filepath.Base(filename)
	if base == "." || base == ".." {
		return "", fmt.Errorf("invalid filename %q", filename)
	}
	return base, nil
}
