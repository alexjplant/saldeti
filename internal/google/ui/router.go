package ui

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterUIRoutes(engine *gin.Engine, baseURL, adminClientID, adminClientSecret string) {
	client := NewGoogleClient(baseURL, adminClientID, adminClientSecret)
	baseTmpl := parseBaseTemplates()
	handler := NewUIHandler(client, baseTmpl)

	uiGroup := engine.Group("/google-ui")
	uiGroup.Use(csrfMiddleware())

	// Serve embedded static files
	staticSub, _ := fs.Sub(staticFS, "static")
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
}

func parseBaseTemplates() *template.Template {
	t := template.New("").Funcs(funcMap())
	tmpl, err := t.ParseFS(templateFS,
		"templates/partials/*.html",
		"templates/layout.html",
	)
	if err != nil {
		panic("Failed to parse embedded base templates: " + err.Error())
	}
	return tmpl
}