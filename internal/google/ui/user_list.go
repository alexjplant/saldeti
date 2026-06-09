package ui

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
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

		users, err := h.client.GetUsers(ctx)
		if err != nil {
			h.render(c, "templates/users/list.html", gin.H{
				"ActiveNav": "users",
				"Error":     "Failed to load users",
			})
			return
		}

		// Apply search filter if provided
		if search != "" {
			var filtered []model.User
			for _, u := range users {
				if containsIgnoreCase(u.PrimaryEmail, search) || containsIgnoreCase(u.DisplayName, search) || containsIgnoreCase(u.GivenName, search) || containsIgnoreCase(u.FamilyName, search) {
					filtered = append(filtered, u)
				}
			}
			users = filtered
		}

		total := len(users)

		// Apply client-side pagination
		skip := (page - 1) * perPage
		start := skip
		if start >= total {
			users = []model.User{}
		} else {
			end := start + perPage
			if end > total {
				end = total
			}
			users = users[start:end]
		}

		totalPages := (total + perPage - 1) / perPage
		if totalPages == 0 {
			totalPages = 1
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
			"Users":         users,
			"TotalCount":    total,
			"Page":          page,
			"PerPage":       perPage,
			"Search":        search,
			"HasPagination": hasPagination,
			"PrevPage":      prevPage,
			"NextPage":      nextPage,
			"CurrentPage":   page,
			"TotalPages":    totalPages,
		})
	}
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}