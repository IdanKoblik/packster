package http

type CreateRequest struct {
	Admin bool `json"admin"`
	Upload bool `json:"upload"`
	Delete bool `json:"delete"`
}
