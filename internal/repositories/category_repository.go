package repositories

import (
	"fmt"

	"jujudb/internal/models"

	"gorm.io/gorm"
)

// categoryRepository implements CategoryRepository interface
type categoryRepository struct {
	*BaseRepository
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new category
func (r *categoryRepository) Create(category *models.Category) error {
	if err := r.db.Create(category).Error; err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

// GetByID retrieves a category by ID
func (r *categoryRepository) GetByID(id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.First(&category, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get category by ID %d: %w", id, err)
	}
	return &category, nil
}

// GetByIDWithRelations retrieves a category by ID with its relations
func (r *categoryRepository) GetByIDWithRelations(id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.Preload("Items").First(&category, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get category by ID %d with relations: %w", id, err)
	}
	return &category, nil
}

// Update updates an existing category
func (r *categoryRepository) Update(category *models.Category) error {
	result := r.db.Save(category)
	if result.Error != nil {
		return fmt.Errorf("failed to update category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("category with ID %d not found", category.ID)
	}
	return nil
}

// Delete hard deletes a category
func (r *categoryRepository) Delete(id uint) error {
	result := r.db.Unscoped().Delete(&models.Category{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("category with ID %d not found", id)
	}
	return nil
}

// SoftDelete soft deletes a category
func (r *categoryRepository) SoftDelete(id uint) error {
	result := r.db.Delete(&models.Category{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to soft delete category: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("category with ID %d not found", id)
	}
	return nil
}

// GetAll retrieves categories with optional filters
func (r *categoryRepository) GetAll(filters CategoryFilters) ([]models.Category, error) {
	var categories []models.Category
	query := r.buildQuery(filters)

	err := query.Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	return categories, nil
}

// Count returns the count of categories matching the filters
func (r *categoryRepository) Count(filters CategoryFilters) (int64, error) {
	var count int64
	query := r.buildQuery(filters)
	err := query.Model(&models.Category{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count categories: %w", err)
	}
	return count, nil
}

// GetWithItems retrieves a category with its items
func (r *categoryRepository) GetWithItems(id uint) (*models.Category, error) {
	var category models.Category
	err := r.db.Preload("Items").Preload("Items.Location").Preload("Items.SubLocation").
		First(&category, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get category with items: %w", err)
	}
	return &category, nil
}

// HasItems checks if a category has associated items
func (r *categoryRepository) HasItems(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Item{}).Where("category_id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if category has items: %w", err)
	}
	return count > 0, nil
}

// GetRelatedItemsCount returns the count of items associated with the category
func (r *categoryRepository) GetRelatedItemsCount(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Item{}).Where("category_id = ?", id).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get related items count: %w", err)
	}
	return count, nil
}

// SearchByName searches categories by name
func (r *categoryRepository) SearchByName(query string) ([]models.Category, error) {
	var categories []models.Category
	dbQuery := r.db.Model(&models.Category{})

	if query != "" {
		dbQuery = dbQuery.Where("name ILIKE ?", "%"+query+"%")
	}

	err := dbQuery.Order("name ASC").Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search categories by name: %w", err)
	}
	return categories, nil
}

// buildQuery builds a GORM query based on the provided filters
func (r *categoryRepository) buildQuery(filters CategoryFilters) *gorm.DB {
	query := r.db.Model(&models.Category{})

	// Apply filters
	if filters.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filters.Name+"%")
	}

	// Apply pagination
	query = applyPagination(query, filters.Limit, filters.Offset)

	// Default ordering
	query = query.Order("name ASC")

	return query
}
