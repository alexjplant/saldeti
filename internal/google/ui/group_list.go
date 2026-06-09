package ui

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/saldeti/saldeti/internal/google/model"
)

func GroupListHandler(h *UIHandler) gin.HandlerFunc {
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

		groups, err := h.client.GetGroups(ctx)
		if err != nil {
			h.render(c, "templates/groups/list.html", gin.H{
				"ActiveNav": "groups",
				"Error":     "Failed to load groups",
			})
			return
		}

		// Apply search filter if provided
		if search != "" {
			var filtered []model.Group
			for _, g := range groups {
				if containsIgnoreCase(g.Email, search) || containsIgnoreCase(g.Name, search) || containsIgnoreCase(g.Description, search) {
					filtered = append(filtered, g)
				}
			}
			groups = filtered
		}

		total := len(groups)

		// Apply client-side pagination
		skip := (page - 1) * perPage
		start := skip
		if start >= total {
			groups = []model.Group{}
		} else {
			end := start + perPage
			if end > total {
				end = total
			}
			groups = groups[start:end]
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

		h.render(c, "templates/groups/list.html", gin.H{
			"ActiveNav":     "groups",
			"Groups":        groups,
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