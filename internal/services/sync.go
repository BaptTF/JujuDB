package services

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"jujudb/internal/models"
)

// SyncService handles synchronization between database and Meilisearch
type SyncService struct {
	db          *gorm.DB
	meilisearch *MeilisearchService
}

// NewSyncService creates a new sync service
func NewSyncService(db *gorm.DB, meilisearch *MeilisearchService) *SyncService {
	return &SyncService{
		db:          db,
		meilisearch: meilisearch,
	}
}

// SyncAllItems indexes all existing items from the database to Meilisearch
func (s *SyncService) SyncAllItems() error {
	logrus.Info("Starting full sync of items to Meilisearch")

	// Query all items from database with relations
	var items []models.Item
	err := s.db.Preload("Location").Preload("SubLocation").Preload("Category").
		Find(&items).Error
	if err != nil {
		return fmt.Errorf("failed to query items: %w", err)
	}

	// Convert to searchable items
	searchableItems := make([]SearchableItem, len(items))
	for i, item := range items {
		searchableItems[i] = s.convertToSearchableItem(item)
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
	// Query single item from database with relations
	var item models.Item
	err := s.db.Preload("Location").Preload("SubLocation").Preload("Category").
		First(&item, itemID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Item doesn't exist, remove from Meilisearch
			return s.meilisearch.DeleteItem(itemID)
		}
		return fmt.Errorf("failed to query item %d: %w", itemID, err)
	}

	// Convert to searchable item
	searchableItem := s.convertToSearchableItem(item)

	// Index item to Meilisearch
	err = s.meilisearch.IndexItem(searchableItem)
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

// convertToSearchableItem converts a model Item to SearchableItem
func (s *SyncService) convertToSearchableItem(item models.Item) SearchableItem {
	searchableItem := SearchableItem{
		ID:          int(item.ID),
		Name:        item.Name,
		Description: item.Description,
		Quantity:    item.Quantity,
		AddedDate:   item.AddedDate,
		Notes:       item.Notes,
	}

	// Handle nullable fields
	if item.LocationID != nil {
		locID := int(*item.LocationID)
		searchableItem.LocationID = &locID
	}
	if item.SubLocationID != nil {
		subID := int(*item.SubLocationID)
		searchableItem.SubLocationID = &subID
	}
	if item.CategoryID != nil {
		catID := int(*item.CategoryID)
		searchableItem.CategoryID = &catID
	}
	if item.ExpiryDate != nil {
		expiryDateStr := item.ExpiryDate.Format("2006-01-02")
		searchableItem.ExpiryDate = &expiryDateStr
	}

	// Add relation data
	if item.Location != nil {
		searchableItem.Location = item.Location.Name
	}
	if item.SubLocation != nil {
		searchableItem.SubLocation = item.SubLocation.Name
	}
	if item.Category != nil {
		searchableItem.Category = item.Category.Name
	}

	return searchableItem
}
