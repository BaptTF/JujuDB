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

// LocationsHandler handles all location-related operations
type LocationsHandler struct {
	service services.LocationService
}

// NewLocationsHandler creates a new locations handler
func NewLocationsHandler(service services.LocationService) *LocationsHandler {
	return &LocationsHandler{
		service: service,
	}
}

// GetLocations handles GET /api/locations
func (h *LocationsHandler) GetLocations(w http.ResponseWriter, r *http.Request) {
	filters := h.parseLocationFilters(r)

	locations, err := h.service.GetLocations(filters)
	if err != nil {
		h.logError("Failed to get locations", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}

// CreateLocation handles POST /api/locations
func (h *LocationsHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var location models.Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		h.logWarn("Invalid JSON payload for creating location", err, r)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CreateLocation(&location); err != nil {
		h.logError("Failed to create location", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(location)
}

// UpdateLocation handles PUT /api/locations/{id}
func (h *LocationsHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.logWarn("Invalid location ID", err, r)
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}

	var location models.Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		h.logWarn("Invalid JSON payload for updating location", err, r)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	location.ID = uint(id)
	if err := h.service.UpdateLocation(&location); err != nil {
		h.logError("Failed to update location", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(location)
}

// DeleteLocation handles DELETE /api/locations/{id}
func (h *LocationsHandler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.logWarn("Invalid location ID", err, r)
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}

	force := r.URL.Query().Get("force") == "true"

	if err := h.service.DeleteLocation(uint(id), force); err != nil {
		// Check if this is a dependency error
		if err.Error() == "location has dependencies: 0 items, 0 sub-locations" {
			// Return conflict with dependencies information
			deps, depsErr := h.service.GetLocationDependencies(uint(id))
			if depsErr == nil {
				h.handleDependenciesError(w, deps, "location")
				return
			}
		}

		h.logError("Failed to delete location", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

// parseLocationFilters parses query parameters into LocationFilters
func (h *LocationsHandler) parseLocationFilters(r *http.Request) repositories.LocationFilters {
	filters := repositories.LocationFilters{}

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
func (h *LocationsHandler) handleDependenciesError(w http.ResponseWriter, deps interface{}, resourceType string) {
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

func (h *LocationsHandler) logError(message string, err error, r *http.Request) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"handler": "locations",
		"method":  r.Method,
		"path":    r.URL.Path,
	}).Error(message)
}

func (h *LocationsHandler) logWarn(message string, err error, r *http.Request) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"handler": "locations",
		"method":  r.Method,
		"path":    r.URL.Path,
	}).Warn(message)
}
