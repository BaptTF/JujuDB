package repositories

import (
	"fmt"

	"jujudb/internal/models"

	"gorm.io/gorm"
)

// locationRepository implements LocationRepository interface
type locationRepository struct {
	*BaseRepository
}

// NewLocationRepository creates a new location repository
func NewLocationRepository(db *gorm.DB) LocationRepository {
	return &locationRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new location
func (r *locationRepository) Create(location *models.Location) error {
	if err := r.db.Create(location).Error; err != nil {
		return fmt.Errorf("failed to create location: %w", err)
	}
	return nil
}

// GetByID retrieves a location by ID
func (r *locationRepository) GetByID(id uint) (*models.Location, error) {
	var location models.Location
	err := r.db.First(&location, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("location with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get location by ID %d: %w", id, err)
	}
	return &location, nil
}

// GetByIDWithRelations retrieves a location by ID with its relations
func (r *locationRepository) GetByIDWithRelations(id uint) (*models.Location, error) {
	var location models.Location
	err := r.db.Preload("SubLocations").Preload("Items").First(&location, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("location with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get location by ID %d with relations: %w", id, err)
	}
	return &location, nil
}

// Update updates an existing location
func (r *locationRepository) Update(location *models.Location) error {
	result := r.db.Save(location)
	if result.Error != nil {
		return fmt.Errorf("failed to update location: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("location with ID %d not found", location.ID)
	}
	return nil
}

// Delete hard deletes a location
func (r *locationRepository) Delete(id uint) error {
	result := r.db.Unscoped().Delete(&models.Location{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete location: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("location with ID %d not found", id)
	}
	return nil
}

// SoftDelete soft deletes a location
func (r *locationRepository) SoftDelete(id uint) error {
	result := r.db.Delete(&models.Location{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to soft delete location: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("location with ID %d not found", id)
	}
	return nil
}

// GetAll retrieves locations with optional filters
func (r *locationRepository) GetAll(filters LocationFilters) ([]models.Location, error) {
	var locations []models.Location
	query := r.buildQuery(filters)

	err := query.Find(&locations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get locations: %w", err)
	}
	return locations, nil
}

// Count returns the count of locations matching the filters
func (r *locationRepository) Count(filters LocationFilters) (int64, error) {
	var count int64
	query := r.buildQuery(filters)
	err := query.Model(&models.Location{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count locations: %w", err)
	}
	return count, nil
}

// GetWithItems retrieves a location with its items
func (r *locationRepository) GetWithItems(id uint) (*models.Location, error) {
	var location models.Location
	err := r.db.Preload("Items").Preload("Items.Category").Preload("Items.SubLocation").
		First(&location, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("location with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get location with items: %w", err)
	}
	return &location, nil
}

// GetWithSubLocations retrieves a location with its sub-locations
func (r *locationRepository) GetWithSubLocations(id uint) (*models.Location, error) {
	var location models.Location
	err := r.db.Preload("SubLocations").First(&location, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("location with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get location with sub-locations: %w", err)
	}
	return &location, nil
}

// HasItems checks if a location has associated items
func (r *locationRepository) HasItems(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Item{}).Where("location_id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if location has items: %w", err)
	}
	return count > 0, nil
}

// HasSubLocations checks if a location has associated sub-locations
func (r *locationRepository) HasSubLocations(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.SubLocation{}).Where("location_id = ?", id).Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check if location has sub-locations: %w", err)
	}
	return count > 0, nil
}

// GetRelatedItemsCount returns the count of items directly associated with the location
func (r *locationRepository) GetRelatedItemsCount(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Item{}).Where("location_id = ?", id).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get related items count: %w", err)
	}
	return count, nil
}

// GetRelatedSubLocationsCount returns the count of sub-locations associated with the location
func (r *locationRepository) GetRelatedSubLocationsCount(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.SubLocation{}).Where("location_id = ?", id).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get related sub-locations count: %w", err)
	}
	return count, nil
}

// SearchByName searches locations by name
func (r *locationRepository) SearchByName(query string) ([]models.Location, error) {
	var locations []models.Location
	dbQuery := r.db.Model(&models.Location{})

	if query != "" {
		dbQuery = dbQuery.Where("name ILIKE ?", "%"+query+"%")
	}

	err := dbQuery.Order("name ASC").Find(&locations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search locations by name: %w", err)
	}
	return locations, nil
}

// buildQuery builds a GORM query based on the provided filters
func (r *locationRepository) buildQuery(filters LocationFilters) *gorm.DB {
	query := r.db.Model(&models.Location{})

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
