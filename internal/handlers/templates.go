package handlers

import (
	"html/template"
	"net/http"
)

// TemplatesHandler handles template rendering
type TemplatesHandler struct {
	Templates *template.Template
}

// NewTemplatesHandler creates a new templates handler
func NewTemplatesHandler() *TemplatesHandler {
	templates := template.Must(template.ParseGlob("web/static/html/*.html"))
	return &TemplatesHandler{Templates: templates}
}

// LoginPage handles GET /
func (h *TemplatesHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	showError := r.URL.Query().Get("error") == "1"
	data := struct {
		Error bool
	}{
		Error: showError,
	}
	h.Templates.ExecuteTemplate(w, "login.html", data)
}

// Dashboard handles GET /dashboard
func (h *TemplatesHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	h.Templates.ExecuteTemplate(w, "dashboard.html", nil)
}

// LocationsPage handles GET /locations
func (h *TemplatesHandler) LocationsPage(w http.ResponseWriter, r *http.Request) {
	h.Templates.ExecuteTemplate(w, "locations.html", nil)
}
