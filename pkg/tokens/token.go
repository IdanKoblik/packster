package tokens

type ApiToken struct {
	Token    string    `json:"token" bson:"_id"`
	Admin    bool      `json:"admin" bson:"admin"`
	Products []Product `json:"products" bson:"products"`
}

type Product struct {
	Name        string             `json:"name" bson:"name"`
	Versions    map[string]Version `json:"versions" bson:"versions"`
	Permissions ProductPermissions `json:"permissions" bson:"permissions"`
}

type Version struct {
	Path     string `json:"path" bson:"path"`
	Checksum string `json:"checksum" bson:"checksum"`
}

type ProductPermissions struct {
	Download bool `json:"download" bson:"download"`
	Upload   bool `json:"upload" bson:"upload"`
	Delete   bool `json:"delete" bson:"delete"`
}
