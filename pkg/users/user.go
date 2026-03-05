package users

type User struct {
	Username string `json:"username"`
	Name string `json:"name"`
	Mail string `json:"mail,omitempty"`
	Password string `json:"password"`
	Permissions UserPermissions `json:"permissions"`
}

type UserPermissions struct {
	Upload bool `json:"upload"`
	Delete bool `json:"delete"`
}
