package services

import (
	"fmt"

	"jujudb/internal/models"
	"jujudb/internal/repositories"

	"github.com/sirupsen/logrus"
)

// subLocationService implements SubLocationService interface
type subLocationService struct {
	repo repositories.SubLocationRepository
	*BaseService
}

// NewSubLocationService creates a new sub-location service
func NewSubLocationService(repo repositories.SubLocationRepository) SubLocationService {
	return &subLocationService{
		repo:        repo,
		BaseService: NewBaseService(),
	}
}

// CreateSubLocation creates a new sub-location with validation
func (s *subLocationService) CreateSubLocation(subLocation *models.SubLocation) error {
	// Validate sub-location
	if err := s.ValidateSubLocation(subLocation); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create sub-location
	if err := s.repo.Create(subLocation); err != nil {
		return fmt.Errorf("failed to create sub-location: %w", err)
	}

	logrus.WithField("sub_location_id", subLocation.ID).Info("Sub-location created successfully")
	return nil
}

// GetSubLocation retrieves a sub-location by ID
func (s *subLocationService) GetSubLocation(id uint) (*models.SubLocation, error) {
	return s.repo.GetByID(id)
}

// GetSubLocationWithRelations retrieves a sub-location by ID with its relations
func (s *subLocationService) GetSubLocationWithRelations(id uint) (*models.SubLocation, error) {
	return s.repo.GetByIDWithRelations(id)
}

// UpdateSubLocation updates an existing sub-location with validation
func (s *subLocationService) UpdateSubLocation(subLocation *models.SubLocation) error {
	// Validate sub-location
	if err := s.ValidateSubLocation(subLocation); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if sub-location exists
	existing, err := s.repo.GetByID(subLocation.ID)
	if err != nil {
		return fmt.Errorf("sub-location not found: %w", err)
	}

	// Update sub-location
	if err := s.repo.Update(subLocation); err != nil {
		return fmt.Errorf("failed to update sub-location: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"sub_location_id": subLocation.ID,
		"old_name":        existing.Name,
		"new_name":        subLocation.Name,
	}).Info("Sub-location updated successfully")
	return nil
}

// DeleteSubLocation deletes a sub-location with dependency checking
func (s *subLocationService) DeleteSubLocation(id uint, force bool) error {
	// Check if sub-location exists
	subLocation, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("sub-location not found: %w", err)
	}

	// Check dependencies
	canDelete, err := s.CanDeleteSubLocation(id)
	if err != nil {
		return fmt.Errorf("failed to check if sub-location can be deleted: %w", err)
	}

	if !canDelete && !force {
		// Return dependencies information
		deps, err := s.GetSubLocationDependencies(id)
		if err != nil {
			return fmt.Errorf("failed to get sub-location dependencies: %w", err)
		}
		return fmt.Errorf("sub-location has %d items", deps.ItemsCount)
	}

	if !canDelete && force {
		// Force delete - this will cascade delete dependencies
		logrus.WithFields(logrus.Fields{
			"sub_location_id": id,
			"force":           true,
		}).Warn("Force deleting sub-location with dependencies")
	}

	// Delete sub-location (cascade will handle dependencies if force is true)
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete sub-location: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"sub_location_id":   id,
		"sub_location_name": subLocation.Name,
		"force":             force,
	}).Info("Sub-location deleted successfully")
	return nil
}

// GetSubLocations retrieves sub-locations with optional filters
func (s *subLocationService) GetSubLocations(filters repositories.SubLocationFilters) ([]models.SubLocation, error) {
	return s.repo.GetAll(filters)
}

// CountSubLocations returns the count of sub-locations matching the filters
func (s *subLocationService) CountSubLocations(filters repositories.SubLocationFilters) (int64, error) {
	return s.repo.Count(filters)
}

// GetSubLocationWithItems retrieves a sub-location with its items
func (s *subLocationService) GetSubLocationWithItems(id uint) (*models.SubLocation, error) {
	return s.repo.GetWithItems(id)
}

// GetSubLocationsByLocation retrieves sub-locations by location ID
func (s *subLocationService) GetSubLocationsByLocation(locationID uint) ([]models.SubLocation, error) {
	return s.repo.GetByLocationID(locationID)
}

// ValidateSubLocation validates a sub-location's data
func (s *subLocationService) ValidateSubLocation(subLocation *models.SubLocation) error {
	return subLocation.Validate()
}

// CanDeleteSubLocation checks if a sub-location can be deleted
func (s *subLocationService) CanDeleteSubLocation(id uint) (bool, error) {
	// Check if sub-location has items
	hasItems, err := s.repo.HasItems(id)
	if err != nil {
		return false, fmt.Errorf("failed to check items: %w", err)
	}

	// Sub-location can be deleted only if it has no items
	return !hasItems, nil
}

// GetSubLocationDependencies returns the dependencies of a sub-location
func (s *subLocationService) GetSubLocationDependencies(id uint) (*SubLocationDependencies, error) {
	deps := &SubLocationDependencies{}

	// Get count
	itemsCount, err := s.repo.GetRelatedItemsCount(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get items count: %w", err)
	}
	deps.ItemsCount = itemsCount

	// If there are dependencies, get the actual data
	if deps.ItemsCount > 0 {
		subLocation, err := s.repo.GetByIDWithRelations(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get sub-location with relations: %w", err)
		}

		deps.Items = subLocation.Items
	}

	return deps, nil
}

// SearchSubLocations searches sub-locations by name, optionally filtered by location
func (s *subLocationService) SearchSubLocations(query string, locationID *uint) ([]models.SubLocation, error) {
	return s.repo.SearchByName(query, locationID)
}
