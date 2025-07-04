package templates

import (
	"embed"
	"fmt"
	"html/template"
	"path/filepath"
	"time"
)

type Templates struct {
	templates map[string]*template.Template
	fs        embed.FS
}

func NewTemplates(fs embed.FS) *Templates {
	return &Templates{
		templates: make(map[string]*template.Template),
		fs:        fs,
	}
}

func (t *Templates) LoadTemplates() error {
	// Define template functions
	funcMap := template.FuncMap{
		"formatDuration": formatDuration,
		"divf":           divf,
	}

	// Define pages that need templates
	pages := []string{"home", "activities", "stats", "bulk-upload", "activity-detail", "gps-track"}

	for _, page := range pages {
		// Parse both base and page template together from embedded filesystem
		baseFile := "templates/layouts/base.html"
		pageFile := filepath.Join("templates/pages", page+".html")
		
		tmpl, err := template.New(page).Funcs(funcMap).ParseFS(t.fs, baseFile, pageFile)
		if err != nil {
			fmt.Printf("Error parsing embedded templates for %s (base: %s, page: %s): %v\n", page, baseFile, pageFile, err)
			return err
		}
		
		t.templates[page] = tmpl
		fmt.Printf("Successfully loaded embedded template: %s\n", page)
	}

	return nil
}

func (t *Templates) GetTemplate(name string) *template.Template {
	return t.templates[name]
}

// Template helper functions
func formatDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%d:%02d", minutes, secs)
}

func divf(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}