package requests

import (
	"artifactor/pkg/users"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Name string `json:"name"`
	Mail string `json:"mail"`
	Password string `json:"password"`
	Admin bool `json:"admin,omitempty"`
	Permissions users.UserPermissions `json:"permissions"`
}
