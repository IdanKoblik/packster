package types

import "packster/internal/utils"

type Product struct {
	Name      string                      `json:"name"`
	GroupName string                      `json:"group_name"`
	Tokens    map[string]TokenPermissions `json:"tokens"`
	Versions  map[string]Version          `json:"versions"`
}

func (p *Product) HashTokens() {
	hashed := make(map[string]TokenPermissions, len(p.Tokens))

	for token, perms := range p.Tokens {
		hashed[utils.Hash(token)] = perms
	}

	p.Tokens = hashed
}

type Version struct {
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
}

type TokenPermissions struct {
	Maintainer bool `json:"maintainer"`
	Download   bool `json:"download"`
	Upload     bool `json:"upload"`
	Delete     bool `json:"delete"`
}
