package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"jujudb/internal/models"
	"jujudb/internal/repositories"
	"jujudb/internal/services"
)

// SubLocationsHandler handles all sub-location-related operations
type SubLocationsHandler struct {
	service services.SubLocationService
}

// NewSubLocationsHandler creates a new sub-locations handler
func NewSubLocationsHandler(service services.SubLocationService) *SubLocationsHandler {
	return &SubLocationsHandler{
		service: service,
	}
}

// GetSubLocations handles GET /api/sub-locations
func (h *SubLocationsHandler) GetSubLocations(w http.ResponseWriter, r *http.Request) {
	filters := h.parseSubLocationFilters(r)

	subLocations, err := h.service.GetSubLocations(filters)
	if err != nil {
		h.logError("Failed to get sub-locations", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subLocations)
}

// CreateSubLocation handles POST /api/sub-locations
func (h *SubLocationsHandler) CreateSubLocation(w http.ResponseWriter, r *http.Request) {
	var subLocation models.SubLocation
	if err := json.NewDecoder(r.Body).Decode(&subLocation); err != nil {
		h.logWarn("Invalid JSON payload for creating sub-location", err, r)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CreateSubLocation(&subLocation); err != nil {
		h.logError("Failed to create sub-location", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subLocation)
}

// UpdateSubLocation handles PUT /api/sub-locations/{id}
func (h *SubLocationsHandler) UpdateSubLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.logWarn("Invalid sub-location ID", err, r)
		http.Error(w, "Invalid sub-location ID", http.StatusBadRequest)
		return
	}

	var subLocation models.SubLocation
	if err := json.NewDecoder(r.Body).Decode(&subLocation); err != nil {
		h.logWarn("Invalid JSON payload for updating sub-location", err, r)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	subLocation.ID = uint(id)
	if err := h.service.UpdateSubLocation(&subLocation); err != nil {
		var validationErr *services.ValidationError
		var notFoundErr *services.NotFoundError
		switch {
		case errors.As(err, &validationErr):
			h.logWarn("Invalid sub-location update request", err, r)
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.As(err, &notFoundErr):
			h.logWarn("Sub-location not found", err, r)
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			h.logError("Failed to update sub-location", err, r)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subLocation)
}

// DeleteSubLocation handles DELETE /api/sub-locations/{id}
func (h *SubLocationsHandler) DeleteSubLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.logWarn("Invalid sub-location ID", err, r)
		http.Error(w, "Invalid sub-location ID", http.StatusBadRequest)
		return
	}

	force := r.URL.Query().Get("force") == "true"

	if err := h.service.DeleteSubLocation(uint(id), force); err != nil {
		// Check if this is a dependency error
		if err.Error() == "sub-location has 0 items" {
			// Return conflict with dependencies information
			deps, depsErr := h.service.GetSubLocationDependencies(uint(id))
			if depsErr == nil {
				h.handleDependenciesError(w, deps, "sub_location")
				return
			}
		}

		h.logError("Failed to delete sub-location", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

// parseSubLocationFilters parses query parameters into SubLocationFilters
func (h *SubLocationsHandler) parseSubLocationFilters(r *http.Request) repositories.SubLocationFilters {
	filters := repositories.SubLocationFilters{}

	// Parse location_id
	if locationIDStr := r.URL.Query().Get("location_id"); locationIDStr != "" {
		if locationID, err := strconv.ParseUint(locationIDStr, 10, 32); err == nil {
			locID := uint(locationID)
			filters.LocationID = &locID
		}
	}

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
func (h *SubLocationsHandler) handleDependenciesError(w http.ResponseWriter, deps interface{}, resourceType string) {
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

func (h *SubLocationsHandler) logError(message string, err error, r *http.Request) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"handler": "sub_locations",
		"method":  r.Method,
		"path":    r.URL.Path,
	}).Error(message)
}

func (h *SubLocationsHandler) logWarn(message string, err error, r *http.Request) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"handler": "sub_locations",
		"method":  r.Method,
		"path":    r.URL.Path,
	}).Warn(message)
}
