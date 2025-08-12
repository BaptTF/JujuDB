package services

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

// SyncService handles synchronization between database and Meilisearch
type SyncService struct {
	db          *sql.DB
	meilisearch *MeilisearchService
}

// NewSyncService creates a new sync service
func NewSyncService(db *sql.DB, meilisearch *MeilisearchService) *SyncService {
	return &SyncService{
		db:          db,
		meilisearch: meilisearch,
	}
}

// SyncAllItems indexes all existing items from the database to Meilisearch
func (s *SyncService) SyncAllItems() error {
	logrus.Info("Starting full sync of items to Meilisearch")

	// Query all items from database
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
		ORDER BY i.added_date DESC`

	rows, err := s.db.Query(sqlQuery)
	if err != nil {
		return fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var searchableItems []SearchableItem
	for rows.Next() {
		var item SearchableItem
		var expiryDate sql.NullString
		var locationID, subLocationID, categoryID sql.NullInt64

		err := rows.Scan(&item.ID, &item.Name, &item.Description, &locationID, &subLocationID, &categoryID,
			&item.Quantity, &expiryDate, &item.AddedDate, &item.Notes,
			&item.Location, &item.SubLocation, &item.Category)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan item row")
			continue
		}

		// Handle nullable fields
		if locationID.Valid {
			id := int(locationID.Int64)
			item.LocationID = &id
		}
		if subLocationID.Valid {
			id := int(subLocationID.Int64)
			item.SubLocationID = &id
		}
		if categoryID.Valid {
			id := int(categoryID.Int64)
			item.CategoryID = &id
		}
		if expiryDate.Valid {
			item.ExpiryDate = &expiryDate.String
		}

		searchableItems = append(searchableItems, item)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over rows: %w", err)
	}

	// Index all items to Meilisearch
	if len(searchableItems) > 0 {
		err = s.meilisearch.IndexItems(searchableItems)
		if err != nil {
			return fmt.Errorf("failed to index items to Meilisearch: %w", err)
		}
	}

	logrus.WithField("items_count", len(searchableItems)).Info("Successfully synced all items to Meilisearch")
	return nil
}

// SyncItem indexes a single item to Meilisearch
func (s *SyncService) SyncItem(itemID int) error {
	// Query single item from database
	sqlQuery := `
		SELECT i.id, i.name, i.description, i.location_id, i.sub_location_id, i.category_id,
		       i.quantity, i.expiry_date, i.added_date, i.notes,
		       COALESCE(l.name, '') as location_name,
		       COALESCE(sl.name, '') as sub_location_name,
		       COALESCE(c.name, '') as category_name
		FROM items i
		LEFT JOIN locations l ON i.location_id = l.id
		LEFT JOIN sub_locations sl ON i.sub_location_idmeilisearch = sl.id
		LEFT JOIN categories c ON i.category_id = c.id
		WHERE i.id = $1`

	var item SearchableItem
	var expiryDate sql.NullString
	var locationID, subLocationID, categoryID sql.NullInt64

	err := s.db.QueryRow(sqlQuery, itemID).Scan(&item.ID, &item.Name, &item.Description, &locationID, &subLocationID, &categoryID,
		&item.Quantity, &expiryDate, &item.AddedDate, &item.Notes,
		&item.Location, &item.SubLocation, &item.Category)
	if err != nil {
		if err == sql.ErrNoRows {
			// Item doesn't exist, remove from Meilisearch
			return s.meilisearch.DeleteItem(itemID)
		}
		return fmt.Errorf("failed to query item %d: %w", itemID, err)
	}

	// Handle nullable fields
	if locationID.Valid {
		id := int(locationID.Int64)
		item.LocationID = &id
	}
	if subLocationID.Valid {
		id := int(subLocationID.Int64)
		item.SubLocationID = &id
	}
	if categoryID.Valid {
		id := int(categoryID.Int64)
		item.CategoryID = &id
	}
	if expiryDate.Valid {
		item.ExpiryDate = &expiryDate.String
	}

	// Index item to Meilisearch
	err = s.meilisearch.IndexItem(item)
	if err != nil {
		return fmt.Errorf("failed to index item %d to Meilisearch: %w", itemID, err)
	}

	logrus.WithField("item_id", itemID).Debug("Successfully synced item to Meilisearch")
	return nil
}

// DeleteItem removes an item from Meilisearch
func (s *SyncService) DeleteItem(itemID int) error {
	err := s.meilisearch.DeleteItem(itemID)
	if err != nil {
		return fmt.Errorf("failed to delete item %d from Meilisearch: %w", itemID, err)
	}

	logrus.WithField("item_id", itemID).Debug("Successfully deleted item from Meilisearch")
	return nil
}
