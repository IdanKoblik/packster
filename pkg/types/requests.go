package types

import "mime/multipart"

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

type UploadRequest struct {
	Product string                `form:"product" binding:"required"`
	Version string                `form:"version" binding:"required"`
	File    *multipart.FileHeader `form:"file" binding:"required"`
}
