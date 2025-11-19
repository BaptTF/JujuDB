package services

import (
	"fmt"

	"jujudb/internal/models"
	"jujudb/internal/repositories"

	"github.com/sirupsen/logrus"
)

// itemService implements ItemService interface
type itemService struct {
	repo repositories.ItemRepository
	sync *SyncService
	*BaseService
}

// NewItemService creates a new item service
func NewItemService(repo repositories.ItemRepository, sync *SyncService) ItemService {
	return &itemService{
		repo:        repo,
		sync:        sync,
		BaseService: NewBaseService(),
	}
}

// CreateItem creates a new item with validation and sync
func (s *itemService) CreateItem(item *models.Item) error {
	// Validate item
	if err := s.ValidateItem(item); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create item
	if err := s.repo.Create(item); err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	// Sync to Meilisearch asynchronously
	if s.sync != nil {
		go func() {
			if err := s.sync.SyncItem(int(item.ID)); err != nil {
				logrus.WithError(err).WithField("item_id", item.ID).Error("Failed to sync created item to Meilisearch")
			}
		}()
	}

	logrus.WithField("item_id", item.ID).Info("Item created successfully")
	return nil
}

// GetItem retrieves an item by ID
func (s *itemService) GetItem(id uint) (*models.Item, error) {
	return s.repo.GetByID(id)
}

// GetItemWithRelations retrieves an item by ID with its relations
func (s *itemService) GetItemWithRelations(id uint) (*models.Item, error) {
	return s.repo.GetByIDWithRelations(id)
}

// UpdateItem updates an existing item with validation and sync
func (s *itemService) UpdateItem(item *models.Item) error {
	// Validate item
	if err := s.ValidateItem(item); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if item exists
	existing, err := s.repo.GetByID(item.ID)
	if err != nil {
		return fmt.Errorf("item not found: %w", err)
	}

	// Update item
	if err := s.repo.Update(item); err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	// Sync to Meilisearch asynchronously
	if s.sync != nil {
		go func() {
			if err := s.sync.SyncItem(int(item.ID)); err != nil {
				logrus.WithError(err).WithField("item_id", item.ID).Error("Failed to sync updated item to Meilisearch")
			}
		}()
	}

	logrus.WithFields(logrus.Fields{
		"item_id":  item.ID,
		"old_name": existing.Name,
		"new_name": item.Name,
	}).Info("Item updated successfully")
	return nil
}

// DeleteItem deletes an item with validation and sync
func (s *itemService) DeleteItem(id uint) error {
	// Check if item exists and can be deleted
	canDelete, err := s.CanDeleteItem(id)
	if err != nil {
		return fmt.Errorf("failed to check if item can be deleted: %w", err)
	}
	if !canDelete {
		return fmt.Errorf("item cannot be deleted")
	}

	// Delete item
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	// Remove from Meilisearch asynchronously
	if s.sync != nil {
		go func() {
			if err := s.sync.DeleteItem(int(id)); err != nil {
				logrus.WithError(err).WithField("item_id", id).Error("Failed to delete item from Meilisearch")
			}
		}()
	}

	logrus.WithField("item_id", id).Info("Item deleted successfully")
	return nil
}

// GetItems retrieves items with optional filters
func (s *itemService) GetItems(filters repositories.ItemFilters) ([]models.Item, error) {
	return s.repo.GetAll(filters)
}

// GetItemsWithRelations retrieves items with relations and optional filters
func (s *itemService) GetItemsWithRelations(filters repositories.ItemFilters) ([]models.Item, error) {
	return s.repo.GetAllWithRelations(filters)
}

// CountItems returns the count of items matching the filters
func (s *itemService) CountItems(filters repositories.ItemFilters) (int64, error) {
	return s.repo.Count(filters)
}

// CreateItems creates multiple items in a single transaction
func (s *itemService) CreateItems(items []models.Item) error {
	if len(items) == 0 {
		return nil
	}

	// Validate all items
	for i, item := range items {
		if err := s.ValidateItem(&item); err != nil {
			return fmt.Errorf("validation failed for item %d: %w", i, err)
		}
	}

	// Create items
	if err := s.repo.CreateBatch(items); err != nil {
		return fmt.Errorf("failed to create items: %w", err)
	}

	// Sync to Meilisearch asynchronously
	if s.sync != nil {
		go func() {
			for _, item := range items {
				if err := s.sync.SyncItem(int(item.ID)); err != nil {
					logrus.WithError(err).WithField("item_id", item.ID).Error("Failed to sync created item to Meilisearch")
				}
			}
		}()
	}

	logrus.WithField("count", len(items)).Info("Items created successfully")
	return nil
}

// UpdateItems updates multiple items in a single transaction
func (s *itemService) UpdateItems(items []models.Item) error {
	if len(items) == 0 {
		return nil
	}

	// Validate all items
	for i, item := range items {
		if err := s.ValidateItem(&item); err != nil {
			return fmt.Errorf("validation failed for item %d: %w", i, err)
		}
	}

	// Update items
	if err := s.repo.UpdateBatch(items); err != nil {
		return fmt.Errorf("failed to update items: %w", err)
	}

	// Sync to Meilisearch asynchronously
	if s.sync != nil {
		go func() {
			for _, item := range items {
				if err := s.sync.SyncItem(int(item.ID)); err != nil {
					logrus.WithError(err).WithField("item_id", item.ID).Error("Failed to sync updated item to Meilisearch")
				}
			}
		}()
	}

	logrus.WithField("count", len(items)).Info("Items updated successfully")
	return nil
}

// DeleteItems deletes multiple items in a single transaction
func (s *itemService) DeleteItems(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	// Check if all items can be deleted
	for _, id := range ids {
		canDelete, err := s.CanDeleteItem(id)
		if err != nil {
			return fmt.Errorf("failed to check if item %d can be deleted: %w", id, err)
		}
		if !canDelete {
			return fmt.Errorf("item %d cannot be deleted", id)
		}
	}

	// Delete items
	if err := s.repo.DeleteBatch(ids); err != nil {
		return fmt.Errorf("failed to delete items: %w", err)
	}

	// Remove from Meilisearch asynchronously
	if s.sync != nil {
		go func() {
			for _, id := range ids {
				if err := s.sync.DeleteItem(int(id)); err != nil {
					logrus.WithError(err).WithField("item_id", id).Error("Failed to delete item from Meilisearch")
				}
			}
		}()
	}

	logrus.WithField("count", len(ids)).Info("Items deleted successfully")
	return nil
}

// SearchItems searches items by name with optional filters
func (s *itemService) SearchItems(query string, filters repositories.ItemFilters) ([]models.Item, error) {
	return s.repo.SearchByName(query, filters)
}

// GetExpiringItems retrieves items that will expire within the specified number of days
func (s *itemService) GetExpiringItems(days int) ([]models.Item, error) {
	return s.repo.GetExpiringItems(days)
}

// GetLowStockItems retrieves items with quantity below the threshold
func (s *itemService) GetLowStockItems(threshold int) ([]models.Item, error) {
	return s.repo.GetLowStockItems(threshold)
}

// GetItemsByLocation retrieves items by location ID
func (s *itemService) GetItemsByLocation(locationID uint) ([]models.Item, error) {
	return s.repo.GetByLocationID(locationID)
}

// GetItemsByCategory retrieves items by category ID
func (s *itemService) GetItemsByCategory(categoryID uint) ([]models.Item, error) {
	return s.repo.GetByCategoryID(categoryID)
}

// GetItemsBySubLocation retrieves items by sub-location ID
func (s *itemService) GetItemsBySubLocation(subLocationID uint) ([]models.Item, error) {
	return s.repo.GetBySubLocationID(subLocationID)
}

// ValidateItem validates an item's data
func (s *itemService) ValidateItem(item *models.Item) error {
	return item.Validate()
}

// CanDeleteItem checks if an item can be deleted
func (s *itemService) CanDeleteItem(id uint) (bool, error) {
	// Check if item exists
	_, err := s.repo.GetByID(id)
	if err != nil {
		return false, err
	}

	// Items can always be deleted (no dependencies)
	return true, nil
}
