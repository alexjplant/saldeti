package ui

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/saldeti/saldeti/internal/model"
)

// GroupRow represents a row in the group list table with computed fields
type GroupRow struct {
	Group       model.Group
	TypeLabel   string
	MemberCount int
}

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
		
		// Use server-side pagination with Skip for groups
		config := &groups.GroupsRequestBuilderGetRequestConfiguration{
			QueryParameters: &groups.GroupsRequestBuilderGetQueryParameters{
				Top:  ptrInt32(int32(perPage)),
				Skip: ptrInt32(int32((page - 1) * perPage)),
			},
		}
		if search != "" {
			// Use $filter instead of $search for better compatibility
			escapedSearch := strings.ReplaceAll(search, "'", "''")
			config.QueryParameters.Filter = ptrString("startswith(displayName,'" + escapedSearch + "')")
		}

		result, err := h.client.Groups().Get(ctx, config)
		if err != nil {
			h.render(c, "templates/groups/list.html", gin.H{
				"ActiveNav": "groups",
				"Error":     "Failed to load groups",
			})
			return
		}

		sdkGroups := result.GetValue()

		// For each group, get member count (N+1 but fine for simulator)
		rows := make([]GroupRow, 0, len(sdkGroups))
		for _, g := range sdkGroups {
			members, _ := h.client.Groups().ByGroupId(*g.GetId()).Members().Get(ctx, nil)
			memberCount := 0
			if members != nil {
				memberCount = len(members.GetValue())
			}
			rows = append(rows, sdkGroupToGroupRow(g, memberCount))
		}

		// Get total count
		totalResult, _ := h.client.Groups().Get(ctx, &groups.GroupsRequestBuilderGetRequestConfiguration{
			QueryParameters: &groups.GroupsRequestBuilderGetQueryParameters{
				Top: ptrInt32(999),
			},
		})
		total := len(totalResult.GetValue())

		// Client-side search filtering if server-side search didn't work
		if search != "" && len(rows) == total {
			var filtered []GroupRow
			for _, row := range rows {
				if containsIgnoreCase(row.Group.DisplayName, search) {
					filtered = append(filtered, row)
				}
			}
			total = len(filtered)
			start := (page - 1) * perPage
			if start >= len(filtered) {
				rows = []GroupRow{}
			} else {
				end := start + perPage
				if end > len(filtered) {
					end = len(filtered)
				}
				rows = filtered[start:end]
			}
		}

		hasPagination := total > perPage
		prevPage, nextPage := 0, 0
		if page > 1 {
			prevPage = page - 1
		}
		if page*perPage < total {
			nextPage = page + 1
		}

		h.render(c, "templates/groups/list.html", gin.H{
			"ActiveNav":     "groups",
			"GroupRows":     rows,
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
