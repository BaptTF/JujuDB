package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"jujudb/internal/models"
	"jujudb/internal/repositories"
	"jujudb/internal/services"
)

// CategoriesHandler handles all category-related operations
type CategoriesHandler struct {
	service services.CategoryService
}

// NewCategoriesHandler creates a new categories handler
func NewCategoriesHandler(service services.CategoryService) *CategoriesHandler {
	return &CategoriesHandler{
		service: service,
	}
}

// GetCategories handles GET /api/categories
func (h *CategoriesHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	filters := h.parseCategoryFilters(r)

	categories, err := h.service.GetCategories(filters)
	if err != nil {
		h.logError("Failed to get categories", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// CreateCategory handles POST /api/categories
func (h *CategoriesHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		h.logWarn("Invalid JSON payload for creating category", err, r)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CreateCategory(&category); err != nil {
		h.logError("Failed to create category", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// UpdateCategory handles PUT /api/categories/{id}
func (h *CategoriesHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.logWarn("Invalid category ID", err, r)
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		h.logWarn("Invalid JSON payload for updating category", err, r)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category.ID = uint(id)
	if err := h.service.UpdateCategory(&category); err != nil {
		h.logError("Failed to update category", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// DeleteCategory handles DELETE /api/categories/{id}
func (h *CategoriesHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.logWarn("Invalid category ID", err, r)
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	force := r.URL.Query().Get("force") == "true"

	if err := h.service.DeleteCategory(uint(id), force); err != nil {
		// Check if this is a dependency error
		if err.Error() == "category has 0 items" {
			// Return conflict with dependencies information
			deps, depsErr := h.service.GetCategoryDependencies(uint(id))
			if depsErr == nil {
				h.handleDependenciesError(w, deps, "category")
				return
			}
		}

		h.logError("Failed to delete category", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

// parseCategoryFilters parses query parameters into CategoryFilters
func (h *CategoriesHandler) parseCategoryFilters(r *http.Request) repositories.CategoryFilters {
	filters := repositories.CategoryFilters{}

	// Parse name filter
	filters.Name = r.URL.Query().Get("name")

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	return filters
}

// handleDependenciesError returns a conflict response with dependencies information
func (h *CategoriesHandler) handleDependenciesError(w http.ResponseWriter, deps interface{}, resourceType string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusConflict)

	response := map[string]interface{}{
		"code":    "HAS_DEPENDENCIES",
		"type":    resourceType,
		"message": "Des éléments sont liés à cette ressource",
	}

	switch resourceType {
	case "location":
		if locationDeps, ok := deps.(*services.LocationDependencies); ok {
			response["related_items"] = locationDeps.Items
			response["related_sublocations"] = locationDeps.SubLocations
		}
	case "category":
		if categoryDeps, ok := deps.(*services.CategoryDependencies); ok {
			response["related_items"] = categoryDeps.Items
		}
	case "sub_location":
		if subLocationDeps, ok := deps.(*services.SubLocationDependencies); ok {
			response["related_items"] = subLocationDeps.Items
		}
	}

	json.NewEncoder(w).Encode(response)
}

// Logging helpers

func (h *CategoriesHandler) logError(message string, err error, r *http.Request) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"handler": "categories",
		"method":  r.Method,
		"path":    r.URL.Path,
	}).Error(message)
}

func (h *CategoriesHandler) logWarn(message string, err error, r *http.Request) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"handler": "categories",
		"method":  r.Method,
		"path":    r.URL.Path,
	}).Warn(message)
}
