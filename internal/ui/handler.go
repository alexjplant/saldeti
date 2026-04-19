package ui

import (
	"html/template"
	"net/http"
	"time"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/store"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

type UIHandler struct {
  store    store.Store // kept for login validation only
  client   *msgraphsdk.GraphServiceClient
  cred     *SimulatorCredential
  baseURL  string
  baseTmpl *template.Template
}

func NewUIHandler(st store.Store, client *msgraphsdk.GraphServiceClient, cred *SimulatorCredential, baseURL string, baseTmpl *template.Template) *UIHandler {
	return &UIHandler{
		store:    st,
		client:   client,
		cred:     cred,
		baseURL:  baseURL,
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

	user := currentUser(c)
	data["LoggedIn"] = user != ""
	data["CurrentUser"] = user

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

func currentUser(c *gin.Context) string {
	user, exists := c.Get("ui_user")
	if !exists {
		return ""
	}
	return user.(string)
}
