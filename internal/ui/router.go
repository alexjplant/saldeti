package ui

import (
	"html/template"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/gin-gonic/gin"
)

// httpTransport wraps an http.Client to implement policy.Transporter
type httpTransport struct {
	client *http.Client
}

func (t *httpTransport) Do(req *http.Request) (*http.Response, error) {
	return t.client.Do(req)
}

func RegisterUIRoutes(engine *gin.Engine, baseURL, adminClientID, adminClientSecret, adminTenantID string) {
	baseTmpl := parseBaseTemplates()

	insecureClient := newInsecureHTTPClient()

	// Use admin client credentials for simulator authentication
	cred, err := azidentity.NewClientSecretCredential(
		adminTenantID, adminClientID, adminClientSecret,
		&azidentity.ClientSecretCredentialOptions{
			ClientOptions: azcore.ClientOptions{
				Cloud: cloud.Configuration{
					ActiveDirectoryAuthorityHost: baseURL,
				},
				Transport: &httpTransport{client: insecureClient},
			},
			DisableInstanceDiscovery: true,
		},
	)
	if err != nil {
		panic("Failed to create admin credential: " + err.Error())
	}

	client, err := newGraphClient(baseURL, cred)
	if err != nil {
		panic("Failed to create Graph SDK client: " + err.Error())
	}

	handler := NewUIHandler(client, cred, baseURL, baseTmpl)

	uiGroup := engine.Group("/ui")

	// Routes
	uiGroup.GET("", DashboardHandler(handler))
	uiGroup.GET("/", DashboardHandler(handler))

	// User routes
	uiGroup.GET("/users", UserListHandler(handler))
	uiGroup.GET("/users/new", UserCreateHandler(handler))
	uiGroup.POST("/users/new", UserCreateHandler(handler))
	uiGroup.GET("/users/:id", UserDetailHandler(handler))
	uiGroup.GET("/users/:id/edit", UserEditHandler(handler))
	uiGroup.POST("/users/:id/edit", UserEditHandler(handler))
	uiGroup.POST("/users/:id/delete", UserDeleteHandler(handler))

	// Group routes
	uiGroup.GET("/groups", GroupListHandler(handler))
	uiGroup.GET("/groups/new", GroupCreateHandler(handler))
	uiGroup.POST("/groups/new", GroupCreateHandler(handler))
	uiGroup.GET("/groups/:id", GroupDetailHandler(handler))
	uiGroup.GET("/groups/:id/edit", GroupEditHandler(handler))
	uiGroup.POST("/groups/:id/edit", GroupEditHandler(handler))
	uiGroup.POST("/groups/:id/delete", GroupDeleteHandler(handler))
	uiGroup.POST("/groups/:id/members/add", GroupAddMemberHandler(handler))
	uiGroup.POST("/groups/:id/members/:memberId/remove", GroupRemoveMemberHandler(handler))
	uiGroup.POST("/groups/:id/owners/add", GroupAddOwnerHandler(handler))
	uiGroup.POST("/groups/:id/owners/:ownerId/remove", GroupRemoveOwnerHandler(handler))
}

func parseBaseTemplates() *template.Template {
	t := template.New("").Funcs(funcMap())
	// Parse all embedded templates at once from the embedded filesystem
	tmpl, err := t.ParseFS(templateFS,
		"templates/partials/*.html",
		"templates/layout.html",
	)
	if err != nil {
		panic("Failed to parse embedded base templates: " + err.Error())
	}
	return tmpl
}
