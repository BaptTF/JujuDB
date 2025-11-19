package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"jujudb/internal/models"
	"jujudb/internal/repositories"
	"jujudb/internal/services"
)

// ItemDTO represents the JSON structure for item API requests/responses
type ItemDTO struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	LocationID    *uint     `json:"location_id"`
	SubLocationID *uint     `json:"sub_location_id"`
	CategoryID    *uint     `json:"category_id"`
	Quantity      int       `json:"quantity"`
	ExpiryDate    *string   `json:"expiry_date"` // Accept string date from frontend
	AddedDate     time.Time `json:"added_date"`
	AddedAt       time.Time `json:"added_at"` // For frontend compatibility
	Notes         *string   `json:"notes"`
	// Display fields from JOINs
	Location    string `json:"location"`
	SubLocation string `json:"sub_location"`
	Category    string `json:"category"`
}

// ArticlesHandler handles all article-related operations
type ArticlesHandler struct {
	service services.ItemService
}

// NewArticlesHandler creates a new articles handler
func NewArticlesHandler(service services.ItemService) *ArticlesHandler {
	return &ArticlesHandler{
		service: service,
	}
}

// GetItems handles GET /api/items
func (h *ArticlesHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	filters := h.parseItemFilters(r)

	items, err := h.service.GetItemsWithRelations(filters)
	if err != nil {
		h.logError("Failed to get items", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	response := h.convertItemsToDTO(items)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateItem handles POST /api/items
func (h *ArticlesHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var itemDTO ItemDTO
	if err := json.NewDecoder(r.Body).Decode(&itemDTO); err != nil {
		h.logWarn("Invalid JSON payload for creating item", err, r)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert DTO to model
	item, err := h.dtoToModel(&itemDTO)
	if err != nil {
		h.logWarn("Invalid item data", err, r)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CreateItem(item); err != nil {
		h.logError("Failed to create item", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert back to DTO for response
	response := h.modelToDTO(item)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateItem handles PUT /api/items/{id}
func (h *ArticlesHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.logWarn("Invalid item ID", err, r)
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var itemDTO ItemDTO
	if err := json.NewDecoder(r.Body).Decode(&itemDTO); err != nil {
		h.logWarn("Invalid JSON payload for updating item", err, r)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert DTO to model
	item, err := h.dtoToModel(&itemDTO)
	if err != nil {
		h.logWarn("Invalid item data", err, r)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	item.ID = uint(id)
	if err := h.service.UpdateItem(item); err != nil {
		h.logError("Failed to update item", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert back to DTO for response
	response := h.modelToDTO(item)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteItem handles DELETE /api/items/{id}
func (h *ArticlesHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		h.logWarn("Invalid item ID", err, r)
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteItem(uint(id)); err != nil {
		h.logError("Failed to delete item", err, r)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper methods

// dtoToModel converts ItemDTO to models.Item
func (h *ArticlesHandler) dtoToModel(dto *ItemDTO) (*models.Item, error) {
	item := &models.Item{
		ID:            dto.ID,
		Name:          dto.Name,
		Description:   dto.Description,
		LocationID:    dto.LocationID,
		SubLocationID: dto.SubLocationID,
		CategoryID:    dto.CategoryID,
		Quantity:      dto.Quantity,
		AddedDate:     dto.AddedDate,
		Notes:         dto.Notes,
	}

	// Parse expiry date if provided
	if dto.ExpiryDate != nil && *dto.ExpiryDate != "" {
		expiryDate, err := time.Parse("2006-01-02", *dto.ExpiryDate)
		if err != nil {
			return nil, err
		}
		item.ExpiryDate = &expiryDate
	}

	return item, nil
}

// modelToDTO converts models.Item to ItemDTO
func (h *ArticlesHandler) modelToDTO(item *models.Item) *ItemDTO {
	dto := &ItemDTO{
		ID:            item.ID,
		Name:          item.Name,
		Description:   item.Description,
		LocationID:    item.LocationID,
		SubLocationID: item.SubLocationID,
		CategoryID:    item.CategoryID,
		Quantity:      item.Quantity,
		AddedDate:     item.AddedDate,
		AddedAt:       item.AddedDate, // For frontend compatibility
		Notes:         item.Notes,
	}

	// Format expiry date as string in ISO format for JavaScript Date parsing
	if item.ExpiryDate != nil {
		// Format as YYYY-MM-DD which JavaScript Date can parse
		expiryDateStr := item.ExpiryDate.Format("2006-01-02")
		dto.ExpiryDate = &expiryDateStr
	}

	// Add relation data if available
	if item.Location != nil {
		dto.Location = item.Location.Name
	}
	if item.SubLocation != nil {
		dto.SubLocation = item.SubLocation.Name
	}
	if item.Category != nil {
		dto.Category = item.Category.Name
	}

	return dto
}

// convertItemsToDTO converts a slice of model Items to DTO slice
func (h *ArticlesHandler) convertItemsToDTO(items []models.Item) []ItemDTO {
	response := make([]ItemDTO, len(items))
	for i, item := range items {
		response[i] = *h.modelToDTO(&item)
	}
	return response
}

// parseItemFilters parses query parameters into ItemFilters
func (h *ArticlesHandler) parseItemFilters(r *http.Request) repositories.ItemFilters {
	filters := repositories.ItemFilters{}

	// Parse location_id
	if locationIDStr := r.URL.Query().Get("location_id"); locationIDStr != "" {
		if locationID, err := strconv.ParseUint(locationIDStr, 10, 32); err == nil {
			locID := uint(locationID)
			filters.LocationID = &locID
		}
	}

	// Parse sub_location_id
	if subLocationIDStr := r.URL.Query().Get("sub_location_id"); subLocationIDStr != "" {
		if subLocationID, err := strconv.ParseUint(subLocationIDStr, 10, 32); err == nil {
			subID := uint(subLocationID)
			filters.SubLocationID = &subID
		}
	}

	// Parse category_id
	if categoryIDStr := r.URL.Query().Get("category_id"); categoryIDStr != "" {
		if categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32); err == nil {
			catID := uint(categoryID)
			filters.CategoryID = &catID
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

	// Parse ordering
	filters.OrderBy = r.URL.Query().Get("order_by")
	filters.OrderDir = r.URL.Query().Get("order_dir")

	return filters
}

// Logging helpers

func (h *ArticlesHandler) logError(message string, err error, r *http.Request) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"handler": "articles",
		"method":  r.Method,
		"path":    r.URL.Path,
	}).Error(message)
}

func (h *ArticlesHandler) logWarn(message string, err error, r *http.Request) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"handler": "articles",
		"method":  r.Method,
		"path":    r.URL.Path,
	}).Warn(message)
}
