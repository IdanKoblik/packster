package http

import (
	"artifactor/pkg/tokens"
)

type RegisterRequest struct {
	Admin    bool             `json:"admin"`
	Products []tokens.Product `json:"products"`
}
