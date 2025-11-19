package repositories

import (
	"fmt"

	"jujudb/internal/models"

	"gorm.io/gorm"
)

// subLocationRepository implements SubLocationRepository interface
type subLocationRepository struct {
	*BaseRepository
}

// NewSubLocationRepository creates a new sub-location repository
func NewSubLocationRepository(db *gorm.DB) SubLocationRepository {
	return &subLocationRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new sub-location
func (r *subLocationRepository) Create(subLocation *models.SubLocation) error {
	if err := r.db.Create(subLocation).Error; err != nil {
		return fmt.Errorf("failed to create sub-location: %w", err)
	}
	return nil
}

// GetByID retrieves a sub-location by ID
func (r *subLocationRepository) GetByID(id uint) (*models.SubLocation, error) {
	var subLocation models.SubLocation
	err := r.db.First(&subLocation, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("sub-location with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get sub-location by ID %d: %w", id, err)
	}
	return &subLocation, nil
}

// GetByIDWithRelations retrieves a sub-location by ID with its relations
func (r *subLocationRepository) GetByIDWithRelations(id uint) (*models.SubLocation, error) {
	var subLocation models.SubLocation
	err := r.db.Preload("Location").Preload("Items").First(&subLocation, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("sub-location with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get sub-location by ID %d with relations: %w", id, err)
	}
	return &subLocation, nil
}

// Update updates an existing sub-location
func (r *subLocationRepository) Update(subLocation *models.SubLocation) error {
	result := r.db.Save(subLocation)
	if result.Error != nil {
		return fmt.Errorf("failed to update sub-location: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("sub-location with ID %d not found", subLocation.ID)
	}
	return nil
}

// Delete hard deletes a sub-location
func (r *subLocationRepository) Delete(id uint) error {
	result := r.db.Unscoped().Delete(&models.SubLocation{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete sub-location: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("sub-location with ID %d not found", id)
	}
	return nil
}

// SoftDelete soft deletes a sub-location
func (r *subLocationRepository) SoftDelete(id uint) error {
	result := r.db.Delete(&models.SubLocation{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to soft delete sub-location: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("sub-location with ID %d not found", id)
	}
	return nil
}

// GetAll retrieves sub-locations with optional filters
func (r *subLocationRepository) GetAll(filters SubLocationFilters) ([]models.SubLocation, error) {
	var subLocations []models.SubLocation
	query := r.buildQuery(filters)

	err := query.Find(&subLocations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-locations: %w", err)
	}
	return subLocations, nil
}

// Count returns the count of sub-locations matching the filters
func (r *subLocationRepository) Count(filters SubLocationFilters) (int64, error) {
	var count int64
	query := r.buildQuery(filters)
	err := query.Model(&models.SubLocation{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count sub-locations: %w", err)
	}
	return count, nil
}

// GetWithItems retrieves a sub-location with its items
func (r *subLocationRepository) GetWithItems(id uint) (*models.SubLocation, error) {
	var subLocation models.SubLocation
	err := r.db.Preload("Items").Preload("Items.Category").Preload("Items.Location").
		First(&subLocation, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("sub-location with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get sub-location with items: %w", err)
	}
	return &subLocation, nil
}

// GetByLocationID retrieves sub-locations by location ID
func (r *subLocationRepository) GetByLocationID(locationID uint) ([]models.SubLocation, error) {
	var subLocations []models.SubLocation

	err := r.db.Where("location_id = ?", locationID).
		Order("name ASC").
		Find(&subLocations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get sub-locations by location ID %d: %w", locationID, err)
	}
	return subLocations, nil
}

// HasItems checks if a sub-location has associated items
func (r *subLocationRepository) HasItems(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Item{}).Where("sub_location_id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if sub-location has items: %w", err)
	}
	return count > 0, nil
}

// GetRelatedItemsCount returns the count of items associated with the sub-location
func (r *subLocationRepository) GetRelatedItemsCount(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Item{}).Where("sub_location_id = ?", id).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get related items count: %w", err)
	}
	return count, nil
}

// SearchByName searches sub-locations by name, optionally filtered by location
func (r *subLocationRepository) SearchByName(query string, locationID *uint) ([]models.SubLocation, error) {
	var subLocations []models.SubLocation
	dbQuery := r.db.Model(&models.SubLocation{})

	if query != "" {
		dbQuery = dbQuery.Where("name ILIKE ?", "%"+query+"%")
	}
	if locationID != nil {
		dbQuery = dbQuery.Where("location_id = ?", *locationID)
	}

	err := dbQuery.Preload("Location").Order("name ASC").Find(&subLocations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search sub-locations by name: %w", err)
	}
	return subLocations, nil
}

// buildQuery builds a GORM query based on the provided filters
func (r *subLocationRepository) buildQuery(filters SubLocationFilters) *gorm.DB {
	query := r.db.Model(&models.SubLocation{})

	// Apply filters
	if filters.LocationID != nil {
		query = query.Where("location_id = ?", *filters.LocationID)
	}
	if filters.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filters.Name+"%")
	}

	// Apply pagination
	query = applyPagination(query, filters.Limit, filters.Offset)

	// Default ordering
	query = query.Order("name ASC")

	return query
}
