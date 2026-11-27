package VueAdmin

import (
	"embed"
	_ "embed"
	"github.com/gin-gonic/gin"
	"net/http"
)

//go:embed index.html
var Html []byte

//go:embed assets/*
var Assets embed.FS

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
	//入口页不缓存:前端产物文件名带 hash,index.html 一旦被缓存就会一直引用旧 JS,
	//出现"改了前端但页面没变化"的情况。这里强制每次校验,静态资源仍可正常缓存。
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.String(200, string(Html))
	return
}
