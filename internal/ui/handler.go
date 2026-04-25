package ui

import (
	"bytes"
	"crypto/tls"
	"html/template"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/gin-gonic/gin"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse // Don't follow redirects
	},
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // InsecureSkipVerify: acceptable for local simulator
		},
	},
}

type UIHandler struct {
  client   *msgraphsdk.GraphServiceClient
  cred     azcore.TokenCredential
  baseURL  string
  baseTmpl *template.Template
}

func NewUIHandler(client *msgraphsdk.GraphServiceClient, cred azcore.TokenCredential, baseURL string, baseTmpl *template.Template) *UIHandler {
	return &UIHandler{
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

// isHtmx checks if the request was initiated by htmx
func isHtmx(c *gin.Context) bool {
	return c.GetHeader("HX-Request") == "true"
}

// renderPartial renders a template without the layout wrapper.
// It sets IsPartial=true so partials can include OOB flash for htmx.
// The caller must set data["Flash"] before calling this method.
func (h *UIHandler) renderPartial(c *gin.Context, templateName string, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["IsPartial"] = true

	t, err := h.baseTmpl.Clone()
	if err != nil {
		http.Error(c.Writer, "Template clone error", http.StatusInternalServerError)
		return
	}

	// Buffer the output so we can catch template errors before writing headers
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, templateName, data); err != nil {
		http.Error(c.Writer, "Template execute error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)
	buf.WriteTo(c.Writer)
}
