package ui

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterUIRoutes(engine *gin.Engine, baseURL, adminClientID, adminClientSecret string) error {
	client := NewGoogleClient(baseURL, adminClientID, adminClientSecret)
	baseTmpl, err := parseBaseTemplates()
	if err != nil {
		return fmt.Errorf("failed to parse base templates: %w", err)
	}
	handler := NewUIHandler(client, baseTmpl)

	uiGroup := engine.Group("/google-ui")
	uiGroup.Use(csrfMiddleware())

	// Serve embedded static files
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return fmt.Errorf("failed to get static sub-filesystem: %w", err)
	}
	uiGroup.StaticFS("/static", http.FS(staticSub))

	// Dashboard routes
	uiGroup.GET("/", DashboardHandler(handler))
	uiGroup.GET("", DashboardHandler(handler))

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

	// Org Unit routes
	uiGroup.GET("/orgunits", OrgUnitListHandler(handler))
	uiGroup.GET("/orgunits/new", OrgUnitCreateHandler(handler))
	uiGroup.POST("/orgunits/new", OrgUnitCreateHandler(handler))

	// Device routes
	uiGroup.GET("/devices", DeviceListHandler(handler))

	// Role routes
	uiGroup.GET("/roles", RoleListHandler(handler))

	// Domain routes
	uiGroup.GET("/domains", DomainListHandler(handler))

	return nil
}

func parseBaseTemplates() (*template.Template, error) {
	t := template.New("").Funcs(funcMap())
	tmpl, err := t.ParseFS(templateFS,
		"templates/partials/*.html",
		"templates/layout.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded base templates: %w", err)
	}
	return tmpl, nil
}
