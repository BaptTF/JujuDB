package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// SubLocation represents a sub-location within a location
type SubLocation struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	LocationID int    `json:"location_id"`
	Location   string `json:"location"` // For display purposes
}

// SubLocationsHandler handles all sub-location-related operations
type SubLocationsHandler struct {
	DB *sql.DB
}

// NewSubLocationsHandler creates a new sub-locations handler
func NewSubLocationsHandler(db *sql.DB) *SubLocationsHandler {
	return &SubLocationsHandler{DB: db}
}

// GetSubLocations handles GET /api/sub-locations
func (h *SubLocationsHandler) GetSubLocations(w http.ResponseWriter, r *http.Request) {
	locationID := r.URL.Query().Get("location_id")

	query := `
		SELECT sl.id, sl.name, sl.location_id, COALESCE(l.name, '') as location_name
		FROM sub_locations sl
		LEFT JOIN locations l ON sl.location_id = l.id`

	var args []interface{}
	if locationID != "" {
		query += " WHERE sl.location_id = $1"
		args = append(args, locationID)
	}

	query += " ORDER BY l.name, sl.name"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler":     "sub_locations",
			"action":      "GetSubLocations",
			"method":      r.Method,
			"path":        r.URL.Path,
			"location_id": locationID,
		}).Error("Failed to query sub-locations")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var subLocations []SubLocation
	for rows.Next() {
		var subLocation SubLocation
		err := rows.Scan(&subLocation.ID, &subLocation.Name, &subLocation.LocationID, &subLocation.Location)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"handler": "sub_locations",
				"action":  "GetSubLocations",
				"method":  r.Method,
				"path":    r.URL.Path,
			}).Error("Failed to scan sub-location row")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		subLocations = append(subLocations, subLocation)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subLocations)
}

// CreateSubLocation handles POST /api/sub-locations
func (h *SubLocationsHandler) CreateSubLocation(w http.ResponseWriter, r *http.Request) {
	var subLocation SubLocation
	if err := json.NewDecoder(r.Body).Decode(&subLocation); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "CreateSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Warn("Invalid JSON payload for creating sub-location")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.DB.QueryRow("INSERT INTO sub_locations (name, location_id) VALUES ($1, $2) RETURNING id", 
		subLocation.Name, subLocation.LocationID).Scan(&subLocation.ID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler":     "sub_locations",
			"action":      "CreateSubLocation",
			"method":      r.Method,
			"path":        r.URL.Path,
			"name":        subLocation.Name,
			"location_id": subLocation.LocationID,
		}).Error("Failed to insert sub-location")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the location name for display
	err = h.DB.QueryRow("SELECT name FROM locations WHERE id = $1", subLocation.LocationID).Scan(&subLocation.Location)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler":     "sub_locations",
			"action":      "CreateSubLocation",
			"method":      r.Method,
			"path":        r.URL.Path,
			"location_id": subLocation.LocationID,
		}).Error("Failed to fetch parent location name after insert")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subLocation)
}

// UpdateSubLocation handles PUT /api/sub-locations/{id}
func (h *SubLocationsHandler) UpdateSubLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "UpdateSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      vars["id"],
		}).Warn("Invalid sub-location ID")
		http.Error(w, "Invalid sub-location ID", http.StatusBadRequest)
		return
	}

	var subLocation SubLocation
	if err := json.NewDecoder(r.Body).Decode(&subLocation); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "UpdateSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Warn("Invalid JSON payload for updating sub-location")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("UPDATE sub_locations SET name = $1, location_id = $2 WHERE id = $3", 
		subLocation.Name, subLocation.LocationID, id)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler":     "sub_locations",
			"action":      "UpdateSubLocation",
			"method":      r.Method,
			"path":        r.URL.Path,
			"id":          id,
			"name":        subLocation.Name,
			"location_id": subLocation.LocationID,
		}).Error("Failed to update sub-location")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "UpdateSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Error("Failed to get rows affected for update")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "UpdateSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Warn("Sub-location not found for update")
		http.Error(w, "Sub-location not found", http.StatusNotFound)
		return
	}

	// Get the location name for display
	err = h.DB.QueryRow("SELECT name FROM locations WHERE id = $1", subLocation.LocationID).Scan(&subLocation.Location)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler":     "sub_locations",
			"action":      "UpdateSubLocation",
			"method":      r.Method,
			"path":        r.URL.Path,
			"location_id": subLocation.LocationID,
		}).Error("Failed to fetch parent location name after update")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	subLocation.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subLocation)
}

// DeleteSubLocation handles DELETE /api/sub-locations/{id}
func (h *SubLocationsHandler) DeleteSubLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "DeleteSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      vars["id"],
		}).Warn("Invalid sub-location ID")
		http.Error(w, "Invalid sub-location ID", http.StatusBadRequest)
		return
	}

	force := r.URL.Query().Get("force") == "true"

	// Check for related items
	type relatedItem struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	rows, qerr := h.DB.Query(`SELECT id, name FROM items WHERE sub_location_id = $1 ORDER BY name`, id)
	if qerr != nil {
		logrus.WithError(qerr).WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "DeleteSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "depQuery",
		}).Error("Failed to load related items before delete")
		http.Error(w, qerr.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var related []relatedItem
	for rows.Next() {
		var it relatedItem
		if err := rows.Scan(&it.ID, &it.Name); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"handler": "sub_locations",
				"action":  "DeleteSubLocation",
				"method":  r.Method,
				"path":    r.URL.Path,
				"id":      id,
				"stage":   "scanRelated",
			}).Error("Failed to scan related item row")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		related = append(related, it)
	}

	if len(related) > 0 && !force {
		logrus.WithFields(logrus.Fields{
			"handler":       "sub_locations",
			"action":        "DeleteSubLocation",
			"method":        r.Method,
			"path":          r.URL.Path,
			"id":            id,
			"related_count": len(related),
		}).Info("Sub-location has related items, returning conflict for confirmation")
		// Return conflict with related items for client-side confirmation
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		resp := map[string]interface{}{
			"code":          "HAS_DEPENDENCIES",
			"type":          "sub_location",
			"id":            id,
			"related_items": related,
			"message":       "Des articles sont liés à ce sous-emplacement",
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// Proceed with transactional delete
	tx, txErr := h.DB.Begin()
	if txErr != nil {
		logrus.WithError(txErr).WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "DeleteSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "txBegin",
		}).Error("Failed to start transaction for delete")
		http.Error(w, txErr.Error(), http.StatusInternalServerError)
		return
	}
	var txErrDefer error
	defer func() {
		if txErrDefer != nil {
			_ = tx.Rollback()
		}
	}()

	if len(related) > 0 {
		if _, txErrDefer = tx.Exec(`DELETE FROM items WHERE sub_location_id = $1`, id); txErrDefer != nil {
			logrus.WithError(txErrDefer).WithFields(logrus.Fields{
				"handler": "sub_locations",
				"action":  "DeleteSubLocation",
				"method":  r.Method,
				"path":    r.URL.Path,
				"id":      id,
				"stage":   "deleteRelated",
			}).Error("Failed to delete related items before deleting sub-location")
			http.Error(w, txErrDefer.Error(), http.StatusInternalServerError)
			return
		}
	}

	result, derr := tx.Exec(`DELETE FROM sub_locations WHERE id = $1`, id)
	if derr != nil {
		logrus.WithError(derr).WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "DeleteSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "deleteSubLocation",
		}).Error("Failed to delete sub-location")
		http.Error(w, derr.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, aerr := result.RowsAffected()
	if aerr != nil {
		logrus.WithError(aerr).WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "DeleteSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "rowsAffected",
		}).Error("Failed to get rows affected for sub-location delete")
		http.Error(w, aerr.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "DeleteSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Warn("Sub-location not found for delete")
		http.Error(w, "Sub-location not found", http.StatusNotFound)
		return
	}

	if txErrDefer = tx.Commit(); txErrDefer != nil {
		logrus.WithError(txErrDefer).WithFields(logrus.Fields{
			"handler": "sub_locations",
			"action":  "DeleteSubLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "txCommit",
		}).Error("Failed to commit transaction for sub-location delete")
		http.Error(w, txErrDefer.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithFields(logrus.Fields{
		"handler":       "sub_locations",
		"action":        "DeleteSubLocation",
		"method":        r.Method,
		"path":          r.URL.Path,
		"id":            id,
		"force":         force,
		"related_count": len(related),
	}).Info("Sub-location deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}
