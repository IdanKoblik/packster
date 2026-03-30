package types

type ApiToken struct {
	Token string `json:"token"`
	Admin bool   `json:"admin"`
}
