package types

import "time"

type Project struct {
	ID         int
	Host       int
	Repository int
	Owner      int
	CreatedAt  time.Time
}

type Product struct {
	ID        int
	Name      string
	Project   int
	CreatedAt time.Time
}

type Version struct {
	ID       int
	Name     string
	Path     string
	Checksum string
	Product  int
}

type Permission struct {
	Account     int
	Project     int
	CanDownload bool
	CanUpload   bool
	CanDelete   bool
}

type PermissionEntry struct {
	Permission
	DisplayName string
}
