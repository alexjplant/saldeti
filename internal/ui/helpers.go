package ui

import (
	"html/template"
	"strings"
	"time"

	"github.com/saldeti/saldeti/internal/model"
)

// Build a map of service plan IDs to names from the default catalog
var servicePlanIDToName = buildServicePlanMap()

func buildServicePlanMap() map[string]string {
	m := make(map[string]string)
	for _, sku := range model.DefaultSubscribedSkus() {
		for _, plan := range sku.ServicePlans {
			m[plan.ServicePlanID] = plan.ServicePlanName
		}
	}
	return m
}

func planIDToName(planID string) string {
	if name, ok := servicePlanIDToName[planID]; ok {
		return name
	}
	return planID // Return GUID if not found
}

func planIDsToNames(planIDs []string) []string {
	names := make([]string, len(planIDs))
	for i, id := range planIDs {
		names[i] = planIDToName(id)
	}
	return names
}

func formatDate(t time.Time) string {
	return t.Format("Jan 02, 2006")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func yesno(b *bool) string {
	if b == nil {
		return "No"
	}
	if *b {
		return "Yes"
	}
	return "No"
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"formatDate":     formatDate,
		"truncate":       truncate,
		"yesno":          yesno,
		"join":           strings.Join,
		"planIDToName":   planIDToName,
		"planIDsToNames": planIDsToNames,
	}
}
