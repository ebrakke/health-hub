package templates

import (
	"fmt"
	"html/template"
	"path/filepath"
	"time"
)

type Templates struct {
	templates map[string]*template.Template
}

func NewTemplates() *Templates {
	return &Templates{
		templates: make(map[string]*template.Template),
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
		tmpl := template.New(page).Funcs(funcMap)
		
		// Parse base layout
		tmpl, err := tmpl.ParseFiles("templates/layouts/base.html")
		if err != nil {
			return err
		}
		
		// Parse page template
		tmpl, err = tmpl.ParseFiles(filepath.Join("templates/pages", page+".html"))
		if err != nil {
			return err
		}
		
		t.templates[page] = tmpl
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