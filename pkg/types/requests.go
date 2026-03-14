package types

type RegisterRequest struct {
	Admin bool `json:"admin,omitempty"`
}

type CreateProductRequest struct {
	Name   string                      `json:"name"`
	Tokens map[string]TokenPermissions `json:"tokens,omitempty" bson:"tokens"`
}

type DeleteProductTokenRequest struct {
	Product string `json:"product"`
	Token   string `json:"token"`
}

type AddProductTokenRequest struct {
	Product     string           `json:"product"`
	Token       string           `json:"token"`
	Permissions TokenPermissions `json:"permissions"`
}

type DeleteVersionRequest struct {
	Product string `json:"product"`
	Version string `json:"version"`
}
