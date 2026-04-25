package ui

import (
	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/applications"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

func DashboardHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// List users via SDK
		usersResult, err := h.client.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
				Top: ptrInt32(999),
			},
		})
		if err != nil {
			h.render(c, "templates/dashboard.html", gin.H{
				"ActiveNav": "dashboard",
				"Error":     "Failed to load data",
			})
			return
		}
		sdkUsers := usersResult.GetValue()

		totalUsers := len(sdkUsers)
		enabledUsers := 0
		disabledUsers := 0
		for _, u := range sdkUsers {
			if u.GetAccountEnabled() != nil && *u.GetAccountEnabled() {
				enabledUsers++
			} else {
				disabledUsers++
			}
		}

		// List groups via SDK
		groupsResult, err := h.client.Groups().Get(ctx, &groups.GroupsRequestBuilderGetRequestConfiguration{
			QueryParameters: &groups.GroupsRequestBuilderGetQueryParameters{
				Top: ptrInt32(999),
			},
		})
		if err != nil {
			h.render(c, "templates/dashboard.html", gin.H{
				"ActiveNav": "dashboard",
				"Error":     "Failed to load data",
			})
			return
		}
		sdkGroups := groupsResult.GetValue()

		totalGroups := len(sdkGroups)
		securityGroups := 0
		unifiedGroups := 0

		for _, group := range sdkGroups {
			if group.GetSecurityEnabled() != nil && *group.GetSecurityEnabled() {
				securityGroups++
			}
			gt := group.GetGroupTypes()
			if gt != nil && len(gt) > 0 {
				unifiedGroups++
			}
		}

		// List applications via SDK
		totalApplications := 0
		appsResult, err := h.client.Applications().Get(ctx, &applications.ApplicationsRequestBuilderGetRequestConfiguration{
			QueryParameters: &applications.ApplicationsRequestBuilderGetQueryParameters{
				Top: ptrInt32(999),
			},
		})
		if err == nil && appsResult != nil {
			totalApplications = len(appsResult.GetValue())
		}

		// List service principals via SDK
		totalSPs := 0
		spsResult, err := h.client.ServicePrincipals().Get(ctx, &serviceprincipals.ServicePrincipalsRequestBuilderGetRequestConfiguration{
			QueryParameters: &serviceprincipals.ServicePrincipalsRequestBuilderGetQueryParameters{
				Top: ptrInt32(999),
			},
		})
		if err == nil && spsResult != nil {
			totalSPs = len(spsResult.GetValue())
		}

		h.render(c, "templates/dashboard.html", gin.H{
			"ActiveNav":              "dashboard",
			"TotalUsers":             totalUsers,
			"EnabledUsers":           enabledUsers,
			"DisabledUsers":          disabledUsers,
			"TotalGroups":            totalGroups,
			"SecurityGroups":         securityGroups,
			"UnifiedGroups":          unifiedGroups,
			"TotalApplications":      totalApplications,
			"TotalServicePrincipals": totalSPs,
		})
	}
}
