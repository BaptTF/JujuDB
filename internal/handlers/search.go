package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
	"jujudb/internal/services"
)

// SearchHandler handles search-related operations
type SearchHandler struct {
	Meilisearch *services.MeilisearchService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(meilisearch *services.MeilisearchService) *SearchHandler {
	return &SearchHandler{
		Meilisearch: meilisearch,
	}
}

// Search handles GET /api/search
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Build search request
	searchReq := services.SearchRequest{
		Query: query,
	}

	// Parse filters
	if locationIDStr := r.URL.Query().Get("location_id"); locationIDStr != "" {
		searchReq.LocationID = locationIDStr
	}
	if subLocationIDStr := r.URL.Query().Get("sub_location_id"); subLocationIDStr != "" {
		searchReq.SubLocationID = subLocationIDStr
	}
	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		searchReq.CategoryID = categoryIDStr
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			searchReq.Limit = limit
		}
	} else {
		searchReq.Limit = 20 // default limit
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			searchReq.Offset = offset
		}
	}

	// Perform search
	results, err := h.Meilisearch.Search(searchReq)
	if err != nil {
		h.logError("Failed to perform search", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := h.convertSearchableItemsToResponse(results)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods

// convertSearchableItemsToResponse converts SearchableItem slice to response format
func (h *SearchHandler) convertSearchableItemsToResponse(items []services.SearchableItem) []map[string]interface{} {
	response := make([]map[string]interface{}, len(items))
	for i, item := range items {
		response[i] = map[string]interface{}{
			"id":           item.ID,
			"name":         item.Name,
			"description":  item.Description,
			"quantity":     item.Quantity,
			"added_date":   item.AddedDate,
			"notes":        item.Notes,
			"location":     item.Location,
			"sub_location": item.SubLocation,
			"category":     item.Category,
		}

		// Handle nullable fields
		if item.LocationID != nil {
			response[i]["location_id"] = *item.LocationID
		}
		if item.SubLocationID != nil {
			response[i]["sub_location_id"] = *item.SubLocationID
		}
		if item.CategoryID != nil {
			response[i]["category_id"] = *item.CategoryID
		}
		if item.ExpiryDate != nil {
			response[i]["expiry_date"] = *item.ExpiryDate
		}
	}
	return response
}

// Logging helpers

func (h *SearchHandler) logError(message string, err error, r *http.Request) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"handler": "search",
		"method":  r.Method,
		"path":    r.URL.Path,
	}).Error(message)
}
