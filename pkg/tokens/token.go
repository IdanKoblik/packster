package tokens

type Token struct {
	Data string `json:"data"`
	Permissions TokenPermissions `json:"permissions"`
}

type TokenPermissions struct {
	Admin bool `json"admin"`
	Upload bool `json:"upload"`
	Delete bool `json:"delete"`
}
