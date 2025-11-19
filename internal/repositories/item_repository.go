package repositories

import (
	"fmt"
	"time"

	"jujudb/internal/models"

	"gorm.io/gorm"
)

// itemRepository implements ItemRepository interface
type itemRepository struct {
	*BaseRepository
}

// NewItemRepository creates a new item repository
func NewItemRepository(db *gorm.DB) ItemRepository {
	return &itemRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new item
func (r *itemRepository) Create(item *models.Item) error {
	if err := r.db.Create(item).Error; err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}
	return nil
}

// GetByID retrieves an item by ID
func (r *itemRepository) GetByID(id uint) (*models.Item, error) {
	var item models.Item
	err := r.db.First(&item, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("item with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get item by ID %d: %w", id, err)
	}
	return &item, nil
}

// GetByIDWithRelations retrieves an item by ID with its relations
func (r *itemRepository) GetByIDWithRelations(id uint) (*models.Item, error) {
	var item models.Item
	err := r.db.Preload("Location").Preload("SubLocation").Preload("Category").First(&item, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("item with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get item by ID %d with relations: %w", id, err)
	}
	return &item, nil
}

// Update updates an existing item
func (r *itemRepository) Update(item *models.Item) error {
	result := r.db.Save(item)
	if result.Error != nil {
		return fmt.Errorf("failed to update item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("item with ID %d not found", item.ID)
	}
	return nil
}

// Delete hard deletes an item
func (r *itemRepository) Delete(id uint) error {
	result := r.db.Unscoped().Delete(&models.Item{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("item with ID %d not found", id)
	}
	return nil
}

// SoftDelete soft deletes an item
func (r *itemRepository) SoftDelete(id uint) error {
	result := r.db.Delete(&models.Item{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to soft delete item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("item with ID %d not found", id)
	}
	return nil
}

// GetAll retrieves items with optional filters
func (r *itemRepository) GetAll(filters ItemFilters) ([]models.Item, error) {
	var items []models.Item
	query := r.buildQuery(filters)

	err := query.Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	return items, nil
}

// GetAllWithRelations retrieves items with relations and optional filters
func (r *itemRepository) GetAllWithRelations(filters ItemFilters) ([]models.Item, error) {
	var items []models.Item
	query := r.buildQuery(filters).Preload("Location").Preload("SubLocation").Preload("Category")

	err := query.Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get items with relations: %w", err)
	}
	return items, nil
}

// Count returns the count of items matching the filters
func (r *itemRepository) Count(filters ItemFilters) (int64, error) {
	var count int64
	query := r.buildQuery(filters)
	err := query.Model(&models.Item{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count items: %w", err)
	}
	return count, nil
}

// CreateBatch creates multiple items in a single transaction
func (r *itemRepository) CreateBatch(items []models.Item) error {
	if len(items) == 0 {
		return nil
	}

	err := r.db.CreateInBatches(items, 100).Error
	if err != nil {
		return fmt.Errorf("failed to create items in batch: %w", err)
	}
	return nil
}

// UpdateBatch updates multiple items in a single transaction
func (r *itemRepository) UpdateBatch(items []models.Item) error {
	if len(items) == 0 {
		return nil
	}

	return r.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Save(&item).Error; err != nil {
				return fmt.Errorf("failed to update item %d: %w", item.ID, err)
			}
		}
		return nil
	})
}

// DeleteBatch deletes multiple items in a single transaction
func (r *itemRepository) DeleteBatch(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	result := r.db.Delete(&models.Item{}, ids)
	if result.Error != nil {
		return fmt.Errorf("failed to delete items in batch: %w", result.Error)
	}
	return nil
}

// SearchByName searches items by name with optional filters
func (r *itemRepository) SearchByName(query string, filters ItemFilters) ([]models.Item, error) {
	var items []models.Item
	dbQuery := r.buildQuery(filters).Preload("Location").Preload("SubLocation").Preload("Category")

	if query != "" {
		dbQuery = dbQuery.Where("name ILIKE ?", "%"+query+"%")
	}

	err := dbQuery.Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search items by name: %w", err)
	}
	return items, nil
}

// GetExpiringItems retrieves items that will expire within the specified number of days
func (r *itemRepository) GetExpiringItems(days int) ([]models.Item, error) {
	var items []models.Item
	expiryDate := time.Now().AddDate(0, 0, days)

	err := r.db.Preload("Location").Preload("SubLocation").Preload("Category").
		Where("expiry_date IS NOT NULL AND expiry_date <= ?", expiryDate).
		Where("expiry_date >= ?", time.Now()).
		Order("expiry_date ASC").
		Find(&items).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get expiring items: %w", err)
	}
	return items, nil
}

// GetLowStockItems retrieves items with quantity below the threshold
func (r *itemRepository) GetLowStockItems(threshold int) ([]models.Item, error) {
	var items []models.Item

	err := r.db.Preload("Location").Preload("SubLocation").Preload("Category").
		Where("quantity <= ?", threshold).
		Order("quantity ASC").
		Find(&items).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get low stock items: %w", err)
	}
	return items, nil
}

// GetByLocationID retrieves items by location ID
func (r *itemRepository) GetByLocationID(locationID uint) ([]models.Item, error) {
	var items []models.Item

	err := r.db.Preload("Location").Preload("SubLocation").Preload("Category").
		Where("location_id = ?", locationID).
		Order("added_date DESC").
		Find(&items).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get items by location ID %d: %w", locationID, err)
	}
	return items, nil
}

// GetByCategoryID retrieves items by category ID
func (r *itemRepository) GetByCategoryID(categoryID uint) ([]models.Item, error) {
	var items []models.Item

	err := r.db.Preload("Location").Preload("SubLocation").Preload("Category").
		Where("category_id = ?", categoryID).
		Order("added_date DESC").
		Find(&items).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get items by category ID %d: %w", categoryID, err)
	}
	return items, nil
}

// GetBySubLocationID retrieves items by sub-location ID
func (r *itemRepository) GetBySubLocationID(subLocationID uint) ([]models.Item, error) {
	var items []models.Item

	err := r.db.Preload("Location").Preload("SubLocation").Preload("Category").
		Where("sub_location_id = ?", subLocationID).
		Order("added_date DESC").
		Find(&items).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get items by sub-location ID %d: %w", subLocationID, err)
	}
	return items, nil
}

// buildQuery builds a GORM query based on the provided filters
func (r *itemRepository) buildQuery(filters ItemFilters) *gorm.DB {
	query := r.db.Model(&models.Item{})

	// Apply filters
	if filters.LocationID != nil {
		query = query.Where("location_id = ?", *filters.LocationID)
	}
	if filters.SubLocationID != nil {
		query = query.Where("sub_location_id = ?", *filters.SubLocationID)
	}
	if filters.CategoryID != nil {
		query = query.Where("category_id = ?", *filters.CategoryID)
	}
	if filters.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filters.Name+"%")
	}
	if filters.ExpiryBefore != nil {
		query = query.Where("expiry_date <= ?", *filters.ExpiryBefore)
	}
	if filters.ExpiryAfter != nil {
		query = query.Where("expiry_date >= ?", *filters.ExpiryAfter)
	}
	if filters.QuantityMin != nil {
		query = query.Where("quantity >= ?", *filters.QuantityMin)
	}
	if filters.QuantityMax != nil {
		query = query.Where("quantity <= ?", *filters.QuantityMax)
	}

	// Apply pagination and ordering
	query = applyPagination(query, filters.Limit, filters.Offset)
	query = applyOrdering(query, filters.OrderBy, filters.OrderDir)

	// Default ordering if not specified
	if filters.OrderBy == "" {
		query = query.Order("added_date DESC")
	}

	return query
}
