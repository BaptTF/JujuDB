package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/sirupsen/logrus"
)

// SearchHandler handles search-related operations
type SearchHandler struct {
	DB *sql.DB
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(db *sql.DB) *SearchHandler {
	return &SearchHandler{DB: db}
}

// Search handles GET /api/search
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	locationID := r.URL.Query().Get("location_id")
	subLocationID := r.URL.Query().Get("sub_location_id")
	categoryID := r.URL.Query().Get("category_id")

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

	sqlQuery := `
		SELECT i.id, i.name, i.description, i.location_id, i.sub_location_id, i.category_id,
		       i.quantity, i.expiry_date, i.added_date, i.notes,
		       COALESCE(l.name, '') as location_name,
		       COALESCE(sl.name, '') as sub_location_name,
		       COALESCE(c.name, '') as category_name
		FROM items i
		LEFT JOIN locations l ON i.location_id = l.id
		LEFT JOIN sub_locations sl ON i.sub_location_id = sl.id
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE TRUE`

	var args []interface{}
	argCount := 0

	if locationID != "" {
		argCount++
		sqlQuery += fmt.Sprintf(" AND i.location_id = $%d", argCount)
		args = append(args, locationID)
	}

	if subLocationID != "" {
		argCount++
		sqlQuery += fmt.Sprintf(" AND i.sub_location_id = $%d", argCount)
		args = append(args, subLocationID)
	}

	if categoryID != "" {
		argCount++
		sqlQuery += fmt.Sprintf(" AND i.category_id = $%d", argCount)
		args = append(args, categoryID)
	}

	sqlQuery += " ORDER BY i.added_date DESC"

	rows, err := h.DB.Query(sqlQuery, args...)
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
		}).Error("Failed to query items for search")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ItemWithDistance struct {
		Item
		Distance int
	}

	var itemsWithDistance []ItemWithDistance
	for rows.Next() {
		var item Item
		var expiryDate sql.NullString
		var locationID, subLocationID, categoryID sql.NullInt64
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &locationID, &subLocationID, &categoryID,
			&item.Quantity, &expiryDate, &item.AddedDate, &item.Notes,
			&item.Location, &item.SubLocation, &item.Category)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"handler": "search",
				"action":  "Search",
				"method":  r.Method,
				"path":    r.URL.Path,
			}).Error("Failed to scan search result row")
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

		// Use RankMatch for better fuzzy matching (returns -1 if no match, or distance if match)
		nameRank := fuzzy.RankMatchNormalized(strings.ToLower(query), strings.ToLower(item.Name))
		descRank := fuzzy.RankMatchNormalized(strings.ToLower(query), strings.ToLower(item.Description))
		
		// Only include items that have a fuzzy match (RankMatch returns -1 for no match)
		if nameRank != -1 || descRank != -1 {
			// Use the best (lowest) rank between name and description
			// If one field doesn't match (-1), use the other field's rank
			totalDistance := nameRank
			if descRank != -1 && (nameRank == -1 || descRank < nameRank) {
				totalDistance = descRank
			}
			
			itemsWithDistance = append(itemsWithDistance, ItemWithDistance{
				Item:     item,
				Distance: totalDistance,
			})
		}
	}

	// Sort by distance (best matches first)
	sort.Slice(itemsWithDistance, func(i, j int) bool {
		return itemsWithDistance[i].Distance < itemsWithDistance[j].Distance
	})

	// Extract just the items for response
	var items []Item
	for _, itemWithDist := range itemsWithDistance {
		items = append(items, itemWithDist.Item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
