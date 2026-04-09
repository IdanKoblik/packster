package ui

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFS embed.FS

func RegisterRoutes(router *gin.Engine) {
	sub, _ := fs.Sub(staticFS, "static")
	router.StaticFS("/ui", http.FS(sub))
}
