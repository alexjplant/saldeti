package ui

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/microsoftgraph/msgraph-sdk-go/serviceprincipals"
	"github.com/saldeti/saldeti/internal/model"
)

// SPRow represents a row in the service principal list table
type SPRow struct {
	SP model.ServicePrincipal
}

func SPListHandler(h *UIHandler) gin.HandlerFunc {
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

		// Use server-side pagination with Skip for service principals
		config := &serviceprincipals.ServicePrincipalsRequestBuilderGetRequestConfiguration{
			QueryParameters: &serviceprincipals.ServicePrincipalsRequestBuilderGetQueryParameters{
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

		result, err := h.client.ServicePrincipals().Get(ctx, config)
		if err != nil {
			h.render(c, "templates/serviceprincipals/list.html", gin.H{
				"ActiveNav": "serviceprincipals",
				"Error":     "Failed to load service principals",
			})
			return
		}

		sdkSPs := result.GetValue()

		rows := make([]SPRow, 0, len(sdkSPs))
		for _, sp := range sdkSPs {
			rows = append(rows, SPRow{
				SP: sdkServicePrincipalToModel(sp),
			})
		}

		// Get total count from @odata.count
		total := len(sdkSPs)
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

		h.render(c, "templates/serviceprincipals/list.html", gin.H{
			"ActiveNav":     "serviceprincipals",
			"SPRows":        rows,
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
