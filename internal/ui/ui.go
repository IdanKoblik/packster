package ui

import (
	"embed"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"packster/internal/logging"
	"packster/pkg/types"

	"github.com/gin-gonic/gin"
)

//go:embed all:static
var staticFiles embed.FS

type authRepo interface {
	FetchToken(rawToken string) (*types.ApiToken, error)
	IsAdmin(rawToken string) (bool, error)
}

type UIHandler struct {
	repo authRepo
}

func NewUIHandler(repo authRepo) *UIHandler {
	return &UIHandler{repo: repo}
}

func SetupUI(repo authRepo, router *gin.Engine) {
	handler := NewUIHandler(repo)
	logging.Log.Info("Web UI enabled, available at /ui")

	router.POST("/ui/login", handler.HandleLogin)
	router.GET("/ui", handler.HandleIndex)
	router.GET("/ui/*filepath", handler.HandleStatic)
}

func (h *UIHandler) HandleIndex(c *gin.Context) {
	h.serveIndex(c)
}

func (h *UIHandler) HandleStatic(c *gin.Context) {
	fp := strings.TrimPrefix(c.Param("filepath"), "/")

	if fp == "" {
		h.serveIndex(c)
		return
	}

	data, err := staticFiles.ReadFile("static/" + fp)
	if err != nil {
		h.serveIndex(c)
		return
	}

	ext := filepath.Ext(fp)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	c.Data(http.StatusOK, mimeType, data)
}

func (h *UIHandler) serveIndex(c *gin.Context) {
	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		c.String(http.StatusServiceUnavailable, "UI not built. Run: make ui-build")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", data)
}

type loginRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *UIHandler) HandleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	token, err := h.repo.FetchToken(req.Token)
	if err != nil || token == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "admin": token.Admin})
}
