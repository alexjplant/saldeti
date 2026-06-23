package ui

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UIHandler struct {
	client   *GoogleClient
	baseTmpl *template.Template
}

func NewUIHandler(client *GoogleClient, baseTmpl *template.Template) *UIHandler {
	return &UIHandler{
		client:   client,
		baseTmpl: baseTmpl,
	}
}

func (h *UIHandler) render(c *gin.Context, pageFile string, data gin.H) {
	// Prepare common template data
	if data == nil {
		data = gin.H{}
	}

	flash := GetFlash(c)
	data["Flash"] = flash

	if token, ok := c.Get("csrf_token"); ok {
		data["CSRFToken"] = token
	} else {
		data["CSRFToken"] = ""
	}

	if _, ok := data["ActiveNav"]; !ok {
		data["ActiveNav"] = ""
	}

	// Clone base template and parse the page file from embedded FS
	t, err := h.baseTmpl.Clone()
	if err != nil {
		http.Error(c.Writer, "Template clone error", http.StatusInternalServerError)
		return
	}

	t, err = t.ParseFS(templateFS, pageFile)
	if err != nil {
		http.Error(c.Writer, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(c.Writer, "layout", data); err != nil {
		http.Error(c.Writer, "Template execute error: "+err.Error(), http.StatusInternalServerError)
	}
}
