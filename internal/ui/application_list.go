package ui

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/applications"
	"github.com/saldeti/saldeti/internal/model"
)

// ApplicationRow represents a row in the application list table with computed fields
type ApplicationRow struct {
	Application model.Application
}

func ApplicationListHandler(h *UIHandler) gin.HandlerFunc {
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

		// Use server-side pagination with Skip for applications
		config := &applications.ApplicationsRequestBuilderGetRequestConfiguration{
			QueryParameters: &applications.ApplicationsRequestBuilderGetQueryParameters{
				Top:   ptrInt32(int32(perPage)),
				Skip:  ptrInt32(int32((page - 1) * perPage)),
				Count: ptrBool(true),
			},
		}
		if search != "" {
			safeSearch := sanitizeODataSearch(search)
			if safeSearch != "" {
				config.QueryParameters.Filter = ptrString("startswith(displayName,'" + safeSearch + "')")
			}
		}

		result, err := h.client.Applications().Get(ctx, config)
		if err != nil {
			h.render(c, "templates/applications/list.html", gin.H{
				"ActiveNav": "applications",
				"Error":     "Failed to load applications",
			})
			return
		}

		sdkApps := result.GetValue()

		rows := make([]ApplicationRow, 0, len(sdkApps))
		for _, a := range sdkApps {
			rows = append(rows, ApplicationRow{
				Application: sdkApplicationToModel(a),
			})
		}

		// Get total count from @odata.count
		total := len(sdkApps)
		if odataCount := result.GetOdataCount(); odataCount != nil {
			total = int(*odataCount)
		}

		hasPagination := total > perPage
		prevPage, nextPage := 0, 0
		if page > 1 {
			prevPage = page - 1
		}
		if page*perPage < total {
			nextPage = page + 1
		}

		h.render(c, "templates/applications/list.html", gin.H{
			"ActiveNav":     "applications",
			"AppRows":       rows,
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
