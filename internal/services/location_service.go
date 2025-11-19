package services

import (
	"fmt"

	"jujudb/internal/models"
	"jujudb/internal/repositories"

	"github.com/sirupsen/logrus"
)

// locationService implements LocationService interface
type locationService struct {
	repo repositories.LocationRepository
	*BaseService
}

// NewLocationService creates a new location service
func NewLocationService(repo repositories.LocationRepository) LocationService {
	return &locationService{
		repo:        repo,
		BaseService: NewBaseService(),
	}
}

// CreateLocation creates a new location with validation
func (s *locationService) CreateLocation(location *models.Location) error {
	// Validate location
	if err := s.ValidateLocation(location); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create location
	if err := s.repo.Create(location); err != nil {
		return fmt.Errorf("failed to create location: %w", err)
	}

	logrus.WithField("location_id", location.ID).Info("Location created successfully")
	return nil
}

// GetLocation retrieves a location by ID
func (s *locationService) GetLocation(id uint) (*models.Location, error) {
	return s.repo.GetByID(id)
}

// GetLocationWithRelations retrieves a location by ID with its relations
func (s *locationService) GetLocationWithRelations(id uint) (*models.Location, error) {
	return s.repo.GetByIDWithRelations(id)
}

// UpdateLocation updates an existing location with validation
func (s *locationService) UpdateLocation(location *models.Location) error {
	// Validate location
	if err := s.ValidateLocation(location); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if location exists
	existing, err := s.repo.GetByID(location.ID)
	if err != nil {
		return fmt.Errorf("location not found: %w", err)
	}

	// Update location
	if err := s.repo.Update(location); err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"location_id": location.ID,
		"old_name":    existing.Name,
		"new_name":    location.Name,
	}).Info("Location updated successfully")
	return nil
}

// DeleteLocation deletes a location with dependency checking
func (s *locationService) DeleteLocation(id uint, force bool) error {
	// Check if location exists
	location, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("location not found: %w", err)
	}

	// Check dependencies
	canDelete, err := s.CanDeleteLocation(id)
	if err != nil {
		return fmt.Errorf("failed to check if location can be deleted: %w", err)
	}

	if !canDelete && !force {
		// Return dependencies information
		deps, err := s.GetLocationDependencies(id)
		if err != nil {
			return fmt.Errorf("failed to get location dependencies: %w", err)
		}
		return fmt.Errorf("location has dependencies: %d items, %d sub-locations",
			deps.ItemsCount, deps.SubLocationsCount)
	}

	if !canDelete && force {
		// Force delete - this will cascade delete dependencies
		logrus.WithFields(logrus.Fields{
			"location_id": id,
			"force":       true,
		}).Warn("Force deleting location with dependencies")
	}

	// Delete location (cascade will handle dependencies if force is true)
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete location: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"location_id":   id,
		"location_name": location.Name,
		"force":         force,
	}).Info("Location deleted successfully")
	return nil
}

// GetLocations retrieves locations with optional filters
func (s *locationService) GetLocations(filters repositories.LocationFilters) ([]models.Location, error) {
	return s.repo.GetAll(filters)
}

// CountLocations returns the count of locations matching the filters
func (s *locationService) CountLocations(filters repositories.LocationFilters) (int64, error) {
	return s.repo.Count(filters)
}

// GetLocationWithItems retrieves a location with its items
func (s *locationService) GetLocationWithItems(id uint) (*models.Location, error) {
	return s.repo.GetWithItems(id)
}

// GetLocationWithSubLocations retrieves a location with its sub-locations
func (s *locationService) GetLocationWithSubLocations(id uint) (*models.Location, error) {
	return s.repo.GetWithSubLocations(id)
}

// ValidateLocation validates a location's data
func (s *locationService) ValidateLocation(location *models.Location) error {
	return location.Validate()
}

// CanDeleteLocation checks if a location can be deleted
func (s *locationService) CanDeleteLocation(id uint) (bool, error) {
	// Check if location has items
	hasItems, err := s.repo.HasItems(id)
	if err != nil {
		return false, fmt.Errorf("failed to check items: %w", err)
	}

	// Check if location has sub-locations
	hasSubLocations, err := s.repo.HasSubLocations(id)
	if err != nil {
		return false, fmt.Errorf("failed to check sub-locations: %w", err)
	}

	// Location can be deleted only if it has no dependencies
	return !hasItems && !hasSubLocations, nil
}

// GetLocationDependencies returns the dependencies of a location
func (s *locationService) GetLocationDependencies(id uint) (*LocationDependencies, error) {
	deps := &LocationDependencies{}

	// Get counts
	itemsCount, err := s.repo.GetRelatedItemsCount(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get items count: %w", err)
	}
	deps.ItemsCount = itemsCount

	subLocationsCount, err := s.repo.GetRelatedSubLocationsCount(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sub-locations count: %w", err)
	}
	deps.SubLocationsCount = subLocationsCount

	// If there are dependencies, get the actual data
	if deps.ItemsCount > 0 || deps.SubLocationsCount > 0 {
		location, err := s.repo.GetByIDWithRelations(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get location with relations: %w", err)
		}

		deps.Items = location.Items
		deps.SubLocations = location.SubLocations
	}

	return deps, nil
}

// SearchLocations searches locations by name
func (s *locationService) SearchLocations(query string) ([]models.Location, error) {
	return s.repo.SearchByName(query)
}
