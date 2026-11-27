package webUser

import (
	"embed"
	"github.com/gin-gonic/gin"
	"net/http"
)

//go:embed index.html
var Html []byte

//go:embed assets/*
var Assets embed.FS

//go:embed static/*
var Static embed.FS

type HtmlHandler struct{}

func NewHtmlHandler() *HtmlHandler {
	return &HtmlHandler{}
}

func (h *HtmlHandler) RedirectIndex(c *gin.Context) {
	c.Redirect(http.StatusFound, "/index.html")
	return
}

func (h *HtmlHandler) Index(c *gin.Context) {
	c.Header("content-type", "text/html;charset=utf-8")
	//入口页不缓存,理由同 VueAdmin:避免 index.html 被缓存后继续引用旧 JS
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.String(200, string(Html))
	return
}
