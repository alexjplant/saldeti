package ui

import (
	"html/template"
	"strings"
	"time"
)

func formatDate(t time.Time) string {
	return t.Format("Jan 02, 2006")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func yesno(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"formatDate": formatDate,
		"truncate":   truncate,
		"yesno":      yesno,
		"join":       strings.Join,
		"contains":   contains,
	}
}