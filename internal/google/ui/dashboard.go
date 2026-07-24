package ui

import (
	"github.com/gin-gonic/gin"
)

func DashboardHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		var (
			totalUsers           int
			totalGroups          int
			totalOrgUnits        int
			totalChromeOSDevices int
			totalMobileDevices   int
			totalRoles           int
			totalDomains         int
			err                  error
		)

		users, err := h.client.GetUsers(ctx)
		if err != nil {
			totalUsers = 0
		} else {
			totalUsers = len(users)
		}

		groups, err := h.client.GetGroups(ctx)
		if err != nil {
			totalGroups = 0
		} else {
			totalGroups = len(groups)
		}

		orgUnits, err := h.client.GetOrgUnits(ctx, "my_customer")
		if err != nil {
			totalOrgUnits = 0
		} else {
			totalOrgUnits = len(orgUnits)
		}

		chromeOSDevices, err := h.client.GetChromeOSDevices(ctx, "my_customer")
		if err != nil {
			totalChromeOSDevices = 0
		} else {
			totalChromeOSDevices = len(chromeOSDevices)
		}

		mobileDevices, err := h.client.GetMobileDevices(ctx, "my_customer")
		if err != nil {
			totalMobileDevices = 0
		} else {
			totalMobileDevices = len(mobileDevices)
		}

		roles, err := h.client.GetRoles(ctx, "my_customer")
		if err != nil {
			totalRoles = 0
		} else {
			totalRoles = len(roles)
		}

		domains, err := h.client.GetDomains(ctx, "my_customer")
		if err != nil {
			totalDomains = 0
		} else {
			totalDomains = len(domains)
		}

		h.render(c, "templates/dashboard.html", gin.H{
			"ActiveNav":            "dashboard",
			"TotalUsers":           totalUsers,
			"TotalGroups":          totalGroups,
			"TotalOrgUnits":        totalOrgUnits,
			"TotalChromeOSDevices": totalChromeOSDevices,
			"TotalMobileDevices":   totalMobileDevices,
			"TotalRoles":           totalRoles,
			"TotalDomains":         totalDomains,
		})
	}
}
