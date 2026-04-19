package ui

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/saldeti/saldeti/internal/model"
)

func UserListHandler(h *UIHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
		search := c.Query("search")
		if page < 1 {
			page = 1
		}
		if perPage < 1 {
			perPage = 20
		}

		ctx := c.Request.Context()

		// Calculate server-side pagination
		// Since SDK doesn't support $skip directly, fetch enough records for the requested page
		skip := (page - 1) * perPage
		top := skip + perPage + 1 // Fetch extra to determine if there's a next page

		// Use server-side $top to limit records fetched (but need more than one page for pagination)
		result, err := h.client.Users().Get(ctx, &users.UsersRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
				Top: ptrInt32(int32(top)),
			},
		})
		if err != nil {
			h.render(c, "templates/users/list.html", gin.H{
				"ActiveNav": "users",
				"Error":     "Failed to load users",
			})
			return
		}

		sdkUsers := result.GetValue()
		// Convert to model types
		usersList := make([]model.User, 0, len(sdkUsers))
		for _, u := range sdkUsers {
			usersList = append(usersList, sdkUserToModel(u))
		}

		// Apply search filter if provided
		if search != "" {
			var filtered []model.User
			for _, u := range usersList {
				if containsIgnoreCase(u.DisplayName, search) || containsIgnoreCase(u.UserPrincipalName, search) {
					filtered = append(filtered, u)
				}
			}
			usersList = filtered
		}

		total := len(usersList)

		// Apply client-side pagination for the specific page
		start := skip
		if start >= total {
			usersList = []model.User{}
		} else {
			end := start + perPage
			if end > total {
				end = total
			}
			usersList = usersList[start:end]
		}

		hasPagination := total > perPage || skip > 0
		prevPage, nextPage := 0, 0
		if page > 1 {
			prevPage = page - 1
		}
		if skip+perPage < total {
			nextPage = page + 1
		}

		h.render(c, "templates/users/list.html", gin.H{
			"ActiveNav":     "users",
			"Users":         usersList,
			"TotalCount":    total,
			"Page":          page,
			"PerPage":       perPage,
			"Search":        search,
			"HasPagination": hasPagination,
			"PrevPage":      prevPage,
			"NextPage":      nextPage,
		})
	}
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
