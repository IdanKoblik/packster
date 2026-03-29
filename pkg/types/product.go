package types

import "packster/internal/utils"

type Product struct {
	Name     string                      `json:"name" bson:"_id"`
	Tokens   map[string]TokenPermissions `json:"tokens" bson:"tokens"`
	Versions map[string]Version          `json:"versions" bson:"versions"`
}

func (p *Product) HashTokens() {
	hashed := make(map[string]TokenPermissions, len(p.Tokens))

	for token, perms := range p.Tokens {
		hashed[utils.Hash(token)] = perms
	}

	p.Tokens = hashed
}

type Version struct {
	Path     string `json:"path" bson:"path"`
	Checksum string `json:"checksum" bson:"checksum"`
}

type TokenPermissions struct {
	Maintainer bool `json:"maintainer" bson:"maintainer"`
	Download   bool `json:"download" bson:"download"`
	Upload     bool `json:"upload" bson:"upload"`
	Delete     bool `json:"delete" bson:"delete"`
}
