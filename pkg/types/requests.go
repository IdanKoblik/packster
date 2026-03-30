package types

import "mime/multipart"

type RegisterRequest struct {
	Admin bool `json:"admin,omitempty"`
}

type CreateProductRequest struct {
	Name      string                      `json:"name"`
	GroupName string                      `json:"group_name"`
	Tokens    map[string]TokenPermissions `json:"tokens,omitempty"`
}

type DeleteProductTokenRequest struct {
	Product   string `json:"product"`
	GroupName string `json:"group_name"`
	Token     string `json:"token"`
}

type AddProductTokenRequest struct {
	Product     string           `json:"product"`
	GroupName   string           `json:"group_name"`
	Token       string           `json:"token"`
	Permissions TokenPermissions `json:"permissions"`
}

type DeleteVersionRequest struct {
	Product   string `json:"product"`
	GroupName string `json:"group_name"`
	Version   string `json:"version"`
}

type UploadRequest struct {
	Product   string                `form:"product" binding:"required"`
	GroupName string                `form:"group_name"`
	Version   string                `form:"version" binding:"required"`
	File      *multipart.FileHeader `form:"file" binding:"required"`
}
