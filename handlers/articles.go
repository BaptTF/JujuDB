package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"jujudb/services"
)

// Item represents an item in the inventory
type Item struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	LocationID    *int      `json:"location_id"`
	SubLocationID *int      `json:"sub_location_id"`
	CategoryID    *int      `json:"category_id"`
	Quantity      int       `json:"quantity"`
	ExpiryDate    *string   `json:"expiry_date"`
	AddedDate     time.Time `json:"added_at"`
	Notes         *string   `json:"notes"`
	// Display fields from JOINs
	Location    string `json:"location"`
	SubLocation string `json:"sub_location"`
	Category    string `json:"category"`
}

// ArticlesHandler handles all article-related operations
type ArticlesHandler struct {
	DB   *sql.DB
	Sync *services.SyncService
}

// NewArticlesHandler creates a new articles handler
func NewArticlesHandler(db *sql.DB, syncService *services.SyncService) *ArticlesHandler {
	return &ArticlesHandler{
		DB:   db,
		Sync: syncService,
	}
}

// GetItems handles GET /api/items
func (h *ArticlesHandler) GetItems(w http.ResponseWriter, r *http.Request) {
	locationID := r.URL.Query().Get("location_id")
	subLocationID := r.URL.Query().Get("sub_location_id")
	categoryID := r.URL.Query().Get("category_id")

	query := `
		SELECT i.id, i.name, i.description, i.location_id, i.sub_location_id, i.category_id,
		       i.quantity, i.expiry_date, i.added_date, i.notes,
		       COALESCE(l.name, '') as location_name,
		       COALESCE(sl.name, '') as sub_location_name,
		       COALESCE(c.name, '') as category_name
		FROM items i
		LEFT JOIN locations l ON i.location_id = l.id
		LEFT JOIN sub_locations sl ON i.sub_location_id = sl.id
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE 1=1`

	var args []interface{}
	argCount := 0

	if locationID != "" {
		argCount++
		query += fmt.Sprintf(" AND i.location_id = $%d", argCount)
		args = append(args, locationID)
	}

	if subLocationID != "" {
		argCount++
		query += fmt.Sprintf(" AND i.sub_location_id = $%d", argCount)
		args = append(args, subLocationID)
	}

	if categoryID != "" {
		argCount++
		query += fmt.Sprintf(" AND i.category_id = $%d", argCount)
		args = append(args, categoryID)
	}

	query += " ORDER BY i.added_date DESC"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler":       "articles",
			"action":        "GetItems",
			"method":        r.Method,
			"path":          r.URL.Path,
			"location_id":   locationID,
			"sub_location_id": subLocationID,
			"category_id":   categoryID,
		}).Error("Failed to query items")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var expiryDate sql.NullString
		var notes sql.NullString
		var locationID, subLocationID, categoryID sql.NullInt64
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &locationID, &subLocationID, &categoryID,
			&item.Quantity, &expiryDate, &item.AddedDate, &notes,
			&item.Location, &item.SubLocation, &item.Category)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"handler": "articles",
				"action":  "GetItems",
				"method":  r.Method,
				"path":    r.URL.Path,
			}).Error("Failed to scan item row")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if locationID.Valid {
			item.LocationID = new(int)
			*item.LocationID = int(locationID.Int64)
		}
		if subLocationID.Valid {
			item.SubLocationID = new(int)
			*item.SubLocationID = int(subLocationID.Int64)
		}
		if categoryID.Valid {
			item.CategoryID = new(int)
			*item.CategoryID = int(categoryID.Int64)
		}
		if expiryDate.Valid {
			item.ExpiryDate = &expiryDate.String
		}
		if notes.Valid {
			item.Notes = &notes.String
		}

		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// CreateItem handles POST /api/items
func (h *ArticlesHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "articles",
			"action":  "CreateItem",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Warn("Invalid JSON payload for creating item")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO items (name, description, location_id, sub_location_id, category_id, quantity, expiry_date, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, added_date`

	err := h.DB.QueryRow(query, item.Name, item.Description, item.LocationID, item.SubLocationID,
		item.CategoryID, item.Quantity, item.ExpiryDate, item.Notes).Scan(&item.ID, &item.AddedDate)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler":        "articles",
			"action":         "CreateItem",
			"method":         r.Method,
			"path":           r.URL.Path,
			"name":           item.Name,
			"location_id":    item.LocationID,
			"sub_location_id": item.SubLocationID,
			"category_id":    item.CategoryID,
			"quantity":       item.Quantity,
		}).Error("Failed to insert item")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Sync item to Meilisearch
	if h.Sync != nil {
		go func() {
			if err := h.Sync.SyncItem(item.ID); err != nil {
				logrus.WithError(err).WithField("item_id", item.ID).Error("Failed to sync created item to Meilisearch")
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

// UpdateItem handles PUT /api/items/{id}
func (h *ArticlesHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"handler": "articles",
			"action":  "UpdateItem",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      vars["id"],
		}).Warn("Invalid item ID")
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
		UPDATE items 
		SET name = $1, description = $2, location_id = $3, sub_location_id = $4, 
		    category_id = $5, quantity = $6, expiry_date = $7, notes = $8
		WHERE id = $9`

	result, err := h.DB.Exec(query, item.Name, item.Description, item.LocationID, item.SubLocationID,
		item.CategoryID, item.Quantity, item.ExpiryDate, item.Notes, id)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "articles",
			"action":  "UpdateItem",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"name":    item.Name,
		}).Error("Failed to update item")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "articles",
			"action":  "UpdateItem",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Error("Failed to get rows affected for update")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"handler": "articles",
			"action":  "UpdateItem",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Warn("Item not found for update")
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	item.ID = id

	// Sync updated item to Meilisearch
	if h.Sync != nil {
		go func() {
			if err := h.Sync.SyncItem(id); err != nil {
				logrus.WithError(err).WithField("item_id", id).Error("Failed to sync updated item to Meilisearch")
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// DeleteItem handles DELETE /api/items/{id}
func (h *ArticlesHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"handler": "articles",
			"action":  "DeleteItem",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      vars["id"],
		}).Warn("Invalid item ID")
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("DELETE FROM items WHERE id = $1", id)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "articles",
			"action":  "DeleteItem",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Error("Failed to delete item")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	// Remove item from Meilisearch
	if h.Sync != nil {
		go func() {
			if err := h.Sync.DeleteItem(id); err != nil {
				logrus.WithError(err).WithField("item_id", id).Error("Failed to delete item from Meilisearch")
			}
		}()
	}

	w.WriteHeader(http.StatusNoContent)
}
