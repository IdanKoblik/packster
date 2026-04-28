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

	assetsSub, _ := fs.Sub(sub, "assets")
	router.StaticFS("/assets", http.FS(assetsSub))

	indexHTML, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		panic(err)
	}

	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}
