package ui

import (
	"context"
	"fmt"
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/store"
)

func RegisterUIRoutes(engine *gin.Engine, st store.Store, port int) {
	baseTmpl := parseBaseTemplates()

	baseURL := fmt.Sprintf("http://localhost:%d", port)

	// Read the first registered client from the store for credentials
	ctx := context.Background()
	clients, err := st.ListClients(ctx)
	if err != nil || len(clients) == 0 {
		panic("UI requires at least one registered client in the store")
	}
	first := clients[0]
	cred := NewSimulatorCredential(baseURL, first.TenantID, first.ClientID, first.ClientSecret)

	client, err := newGraphClient(baseURL, cred)
	if err != nil {
		panic("Failed to create Graph SDK client: " + err.Error())
	}

	handler := NewUIHandler(st, client, cred, baseURL, baseTmpl)

	uiGroup := engine.Group("/ui")

	// Unprotected routes
	uiGroup.GET("/login", LoginHandler(handler))
	uiGroup.POST("/login", LoginHandler(handler))
	uiGroup.GET("/logout", LogoutHandler(handler))

	// Protected routes
	uiGroup.Use(AuthMiddleware())
	{
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
