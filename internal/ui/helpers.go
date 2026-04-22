package ui

import (
	"html/template"
	"strings"
	"time"
)

func icon(name string) template.HTML {
	return template.HTML("<i data-lucide=\"" + name + "\"></i>")
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
		"icon":       icon,
		"formatDate": formatDate,
		"truncate":   truncate,
		"yesno":      yesno,
		"join":       strings.Join,
	}
}
