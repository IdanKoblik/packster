package utils

import (
	"fmt"
	"path/filepath"
	"regexp"
)

var safeNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._\-]*$`)

// ValidateName rejects names that contain path-traversal characters or other
// unsafe characters. Valid names start with an alphanumeric character and
// contain only letters, digits, '.', '-', and '_'.
func ValidateName(s string) error {
	if !safeNameRe.MatchString(s) {
		return fmt.Errorf("invalid name %q: must start with alphanumeric and contain only letters, digits, '.', '-', '_'", s)
	}
	return nil
}

// SafeFilename strips any directory components from a filename to prevent
// path traversal attacks. Returns an error if the result is not a usable name.
func SafeFilename(filename string) (string, error) {
	base := filepath.Base(filename)
	if base == "." || base == ".." {
		return "", fmt.Errorf("invalid filename %q", filename)
	}
	return base, nil
}
