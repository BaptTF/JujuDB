package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Location represents a storage location
type Location struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// LocationsHandler handles all location-related operations
type LocationsHandler struct {
	DB *sql.DB
}

// NewLocationsHandler creates a new locations handler
func NewLocationsHandler(db *sql.DB) *LocationsHandler {
	return &LocationsHandler{DB: db}
}

// GetLocations handles GET /api/locations
func (h *LocationsHandler) GetLocations(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, name FROM locations ORDER BY name")
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "GetLocations",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Error("Failed to query locations")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var locations []Location
	for rows.Next() {
		var location Location
		err := rows.Scan(&location.ID, &location.Name)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"handler": "locations",
				"action":  "GetLocations",
				"method":  r.Method,
				"path":    r.URL.Path,
			}).Error("Failed to scan location row")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		locations = append(locations, location)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}

// CreateLocation handles POST /api/locations
func (h *LocationsHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	var location Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "CreateLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Warn("Invalid JSON payload for creating location")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.DB.QueryRow("INSERT INTO locations (name) VALUES ($1) RETURNING id", location.Name).Scan(&location.ID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "CreateLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"name":    location.Name,
		}).Error("Failed to insert location")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(location)
}

// UpdateLocation handles PUT /api/locations/{id}
func (h *LocationsHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "UpdateLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      vars["id"],
		}).Warn("Invalid location ID")
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}

	var location Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "UpdateLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Warn("Invalid JSON payload for updating location")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("UPDATE locations SET name = $1 WHERE id = $2", location.Name, id)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "UpdateLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"name":    location.Name,
		}).Error("Failed to update location")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "UpdateLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Error("Failed to get rows affected for update")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "UpdateLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Warn("Location not found for update")
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	location.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(location)
}

// DeleteLocation handles DELETE /api/locations/{id}
func (h *LocationsHandler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "DeleteLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      vars["id"],
		}).Warn("Invalid location ID")
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}

	// Support conditional deletion with dependency awareness
	force := r.URL.Query().Get("force") == "true"

	// Find related items either directly on location or via sub-locations
	type relatedItem struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	depQuery := `
		SELECT id, name FROM items
		WHERE location_id = $1
		   OR sub_location_id IN (SELECT id FROM sub_locations WHERE location_id = $1)
		ORDER BY name`

	rows, qerr := h.DB.Query(depQuery, id)
	if qerr != nil {
		logrus.WithError(qerr).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "DeleteLocation",
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
				"handler": "locations",
				"action":  "DeleteLocation",
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

	// Also find related sub-locations for this location
	type relatedSubLocation struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	slRows, slErr := h.DB.Query(`SELECT id, name FROM sub_locations WHERE location_id = $1 ORDER BY name`, id)
	if slErr != nil {
		logrus.WithError(slErr).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "DeleteLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "depQuerySubLocs",
		}).Error("Failed to load related sub-locations before delete")
		http.Error(w, slErr.Error(), http.StatusInternalServerError)
		return
	}
	defer slRows.Close()

	var relatedSubLocs []relatedSubLocation
	for slRows.Next() {
		var sl relatedSubLocation
		if err := slRows.Scan(&sl.ID, &sl.Name); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"handler": "locations",
				"action":  "DeleteLocation",
				"method":  r.Method,
				"path":    r.URL.Path,
				"id":      id,
				"stage":   "scanRelatedSubLocs",
			}).Error("Failed to scan related sub-location row")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		relatedSubLocs = append(relatedSubLocs, sl)
	}

	if (len(related) > 0 || len(relatedSubLocs) > 0) && !force {
		logrus.WithFields(logrus.Fields{
			"handler":       "locations",
			"action":        "DeleteLocation",
			"method":        r.Method,
			"path":          r.URL.Path,
			"id":            id,
			"related_count": len(related),
			"related_subloc_count": len(relatedSubLocs),
		}).Info("Location has related items or sub-locations, returning conflict for confirmation")
		// Return conflict with related items for client-side confirmation
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		resp := map[string]interface{}{
			"code":          "HAS_DEPENDENCIES",
			"type":          "location",
			"id":            id,
			"related_items": related,
			"related_sublocations": relatedSubLocs,
			"message":       "Des éléments sont liés à cet emplacement (articles et/ou sous-emplacements)",
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// Proceed with destructive deletion in a transaction
	tx, txErr := h.DB.Begin()
	if txErr != nil {
		logrus.WithError(txErr).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "DeleteLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "txBegin",
		}).Error("Failed to start transaction for delete")
		http.Error(w, txErr.Error(), http.StatusInternalServerError)
		return
	}
	// rollback on any subsequent error via named return err
	var txErrDefer error
	defer func() {
		if txErrDefer != nil {
			_ = tx.Rollback()
		}
	}()

	// If forcing and there are related items, delete them
	var deletedBySubLoc int64
	var deletedByLoc int64
	var deletedSubLocs int64
	if len(related) > 0 {
		if res, execErr := tx.Exec(`DELETE FROM items WHERE sub_location_id IN (SELECT id FROM sub_locations WHERE location_id = $1)`, id); execErr != nil {
			txErrDefer = execErr
			logrus.WithError(execErr).WithFields(logrus.Fields{
				"handler": "locations",
				"action":  "DeleteLocation",
				"method":  r.Method,
				"path":    r.URL.Path,
				"id":      id,
				"stage":   "deleteRelatedBySubLocations",
			}).Error("Failed to delete items via sub-locations")
			http.Error(w, execErr.Error(), http.StatusInternalServerError)
			return
		} else if c, _ := res.RowsAffected(); c >= 0 {
			deletedBySubLoc = c
		}
		if res, execErr := tx.Exec(`DELETE FROM items WHERE location_id = $1`, id); execErr != nil {
			txErrDefer = execErr
			logrus.WithError(execErr).WithFields(logrus.Fields{
				"handler": "locations",
				"action":  "DeleteLocation",
				"method":  r.Method,
				"path":    r.URL.Path,
				"id":      id,
				"stage":   "deleteRelatedByLocation",
			}).Error("Failed to delete items directly on location")
			http.Error(w, execErr.Error(), http.StatusInternalServerError)
			return
		} else if c, _ := res.RowsAffected(); c >= 0 {
			deletedByLoc = c
		}
	}

	// Delete sub-locations for this location
	if res, execErr := tx.Exec(`DELETE FROM sub_locations WHERE location_id = $1`, id); execErr != nil {
		txErrDefer = execErr
		logrus.WithError(execErr).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "DeleteLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "deleteSubLocations",
		}).Error("Failed to delete sub-locations for location")
		http.Error(w, execErr.Error(), http.StatusInternalServerError)
		return
	} else if c, _ := res.RowsAffected(); c >= 0 {
		deletedSubLocs = c
	}

	// Delete the location itself
	result, derr := tx.Exec(`DELETE FROM locations WHERE id = $1`, id)
	if derr != nil {
		logrus.WithError(derr).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "DeleteLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "deleteLocation",
		}).Error("Failed to delete location")
		http.Error(w, derr.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, aerr := result.RowsAffected()
	if aerr != nil {
		logrus.WithError(aerr).WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "DeleteLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "rowsAffected",
		}).Error("Failed to get rows affected for location delete")
		http.Error(w, aerr.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"handler": "locations",
			"action":  "DeleteLocation",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Warn("Location not found for delete")
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	if txErrDefer = tx.Commit(); txErrDefer != nil {
		logrus.WithError(txErrDefer).WithFields(logrus.Fields{
			"handler":           "locations",
			"action":            "DeleteLocation",
			"method":            r.Method,
			"path":              r.URL.Path,
			"id":                id,
			"deleted_items_sub":  deletedBySubLoc,
			"deleted_items_loc":  deletedByLoc,
			"deleted_sub_locs":   deletedSubLocs,
		}).Error("Failed to commit transaction for delete")
		http.Error(w, txErrDefer.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithFields(logrus.Fields{
		"handler":           "locations",
		"action":            "DeleteLocation",
		"method":            r.Method,
		"path":              r.URL.Path,
		"id":                id,
		"force":             force,
		"deleted_items_sub": deletedBySubLoc,
		"deleted_items_loc": deletedByLoc,
		"deleted_sub_locs":  deletedSubLocs,
	}).Info("Location deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}
