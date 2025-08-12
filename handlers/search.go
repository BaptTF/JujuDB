package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"jujudb/services"
	"github.com/sirupsen/logrus"
)

// SearchHandler handles search-related operations
type SearchHandler struct {
	DB             *sql.DB
	Meilisearch    *services.MeilisearchService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(db *sql.DB, meilisearch *services.MeilisearchService) *SearchHandler {
	return &SearchHandler{
		DB:          db,
		Meilisearch: meilisearch,
	}
}

// Search handles GET /api/search
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	locationID := r.URL.Query().Get("location_id")
	subLocationID := r.URL.Query().Get("sub_location_id")
	categoryID := r.URL.Query().Get("category_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if query == "" {
		logrus.WithFields(logrus.Fields{
			"handler": "search",
			"action":  "Search",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Warn("Missing search query")
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limit := 50 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // default offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Perform Meilisearch query
	searchReq := services.SearchRequest{
		Query:         query,
		LocationID:    locationID,
		SubLocationID: subLocationID,
		CategoryID:    categoryID,
		Limit:         limit,
		Offset:        offset,
	}

	searchableItems, err := h.Meilisearch.Search(searchReq)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler":         "search",
			"action":          "Search",
			"method":          r.Method,
			"path":            r.URL.Path,
			"q":               query,
			"location_id":     locationID,
			"sub_location_id": subLocationID,
			"category_id":     categoryID,
		}).Error("Meilisearch query failed")
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	// Convert SearchableItem to Item for response
	var items []Item
	for _, searchableItem := range searchableItems {
		item := Item{
			ID:            searchableItem.ID,
			Name:          searchableItem.Name,
			Description:   searchableItem.Description,
			LocationID:    searchableItem.LocationID,
			SubLocationID: searchableItem.SubLocationID,
			CategoryID:    searchableItem.CategoryID,
			Location:      searchableItem.Location,
			SubLocation:   searchableItem.SubLocation,
			Category:      searchableItem.Category,
			Quantity:      searchableItem.Quantity,
			ExpiryDate:    searchableItem.ExpiryDate,
			AddedDate:     searchableItem.AddedDate,
			Notes:         searchableItem.Notes,
		}
		items = append(items, item)
	}

	logrus.WithFields(logrus.Fields{
		"query":   query,
		"results": len(items),
	}).Info("Search completed successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
