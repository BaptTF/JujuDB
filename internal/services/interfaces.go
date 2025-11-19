package services

import (
	"jujudb/internal/models"
	"jujudb/internal/repositories"
)

// ItemService defines the interface for item business logic
type ItemService interface {
	// CRUD operations
	CreateItem(item *models.Item) error
	GetItem(id uint) (*models.Item, error)
	GetItemWithRelations(id uint) (*models.Item, error)
	UpdateItem(item *models.Item) error
	DeleteItem(id uint) error

	// Query operations
	GetItems(filters repositories.ItemFilters) ([]models.Item, error)
	GetItemsWithRelations(filters repositories.ItemFilters) ([]models.Item, error)
	CountItems(filters repositories.ItemFilters) (int64, error)

	// Bulk operations
	CreateItems(items []models.Item) error
	UpdateItems(items []models.Item) error
	DeleteItems(ids []uint) error

	// Search and filtering
	SearchItems(query string, filters repositories.ItemFilters) ([]models.Item, error)
	GetExpiringItems(days int) ([]models.Item, error)
	GetLowStockItems(threshold int) ([]models.Item, error)

	// Relationship operations
	GetItemsByLocation(locationID uint) ([]models.Item, error)
	GetItemsByCategory(categoryID uint) ([]models.Item, error)
	GetItemsBySubLocation(subLocationID uint) ([]models.Item, error)

	// Business logic methods
	ValidateItem(item *models.Item) error
	CanDeleteItem(id uint) (bool, error)
}

// LocationService defines the interface for location business logic
type LocationService interface {
	// CRUD operations
	CreateLocation(location *models.Location) error
	GetLocation(id uint) (*models.Location, error)
	GetLocationWithRelations(id uint) (*models.Location, error)
	UpdateLocation(location *models.Location) error
	DeleteLocation(id uint, force bool) error

	// Query operations
	GetLocations(filters repositories.LocationFilters) ([]models.Location, error)
	CountLocations(filters repositories.LocationFilters) (int64, error)

	// Relationship operations
	GetLocationWithItems(id uint) (*models.Location, error)
	GetLocationWithSubLocations(id uint) (*models.Location, error)

	// Business logic methods
	ValidateLocation(location *models.Location) error
	CanDeleteLocation(id uint) (bool, error)
	GetLocationDependencies(id uint) (*LocationDependencies, error)
	SearchLocations(query string) ([]models.Location, error)
}

// CategoryService defines the interface for category business logic
type CategoryService interface {
	// CRUD operations
	CreateCategory(category *models.Category) error
	GetCategory(id uint) (*models.Category, error)
	GetCategoryWithRelations(id uint) (*models.Category, error)
	UpdateCategory(category *models.Category) error
	DeleteCategory(id uint, force bool) error

	// Query operations
	GetCategories(filters repositories.CategoryFilters) ([]models.Category, error)
	CountCategories(filters repositories.CategoryFilters) (int64, error)

	// Relationship operations
	GetCategoryWithItems(id uint) (*models.Category, error)

	// Business logic methods
	ValidateCategory(category *models.Category) error
	CanDeleteCategory(id uint) (bool, error)
	GetCategoryDependencies(id uint) (*CategoryDependencies, error)
	SearchCategories(query string) ([]models.Category, error)
}

// SubLocationService defines the interface for sub-location business logic
type SubLocationService interface {
	// CRUD operations
	CreateSubLocation(subLocation *models.SubLocation) error
	GetSubLocation(id uint) (*models.SubLocation, error)
	GetSubLocationWithRelations(id uint) (*models.SubLocation, error)
	UpdateSubLocation(subLocation *models.SubLocation) error
	DeleteSubLocation(id uint, force bool) error

	// Query operations
	GetSubLocations(filters repositories.SubLocationFilters) ([]models.SubLocation, error)
	CountSubLocations(filters repositories.SubLocationFilters) (int64, error)

	// Relationship operations
	GetSubLocationWithItems(id uint) (*models.SubLocation, error)
	GetSubLocationsByLocation(locationID uint) ([]models.SubLocation, error)

	// Business logic methods
	ValidateSubLocation(subLocation *models.SubLocation) error
	CanDeleteSubLocation(id uint) (bool, error)
	GetSubLocationDependencies(id uint) (*SubLocationDependencies, error)
	SearchSubLocations(query string, locationID *uint) ([]models.SubLocation, error)
}

// Service provides access to all services
type Service struct {
	Items        ItemService
	Locations    LocationService
	Categories   CategoryService
	SubLocations SubLocationService
}

// NewService creates a new service with all sub-services
func NewService(repos *repositories.Repository, syncService *SyncService) *Service {
	return &Service{
		Items:        NewItemService(repos.Items, syncService),
		Locations:    NewLocationService(repos.Locations),
		Categories:   NewCategoryService(repos.Categories),
		SubLocations: NewSubLocationService(repos.SubLocations),
	}
}

// LocationDependencies represents the dependencies of a location
type LocationDependencies struct {
	ItemsCount        int64                `json:"items_count"`
	SubLocationsCount int64                `json:"sub_locations_count"`
	Items             []models.Item        `json:"items,omitempty"`
	SubLocations      []models.SubLocation `json:"sub_locations,omitempty"`
}

// CategoryDependencies represents the dependencies of a category
type CategoryDependencies struct {
	ItemsCount int64         `json:"items_count"`
	Items      []models.Item `json:"items,omitempty"`
}

// SubLocationDependencies represents the dependencies of a sub-location
type SubLocationDependencies struct {
	ItemsCount int64         `json:"items_count"`
	Items      []models.Item `json:"items,omitempty"`
}

// BaseService provides common functionality for all services
type BaseService struct {
	// Common dependencies can be added here
	// For example: logger, config, etc.
}

// NewBaseService creates a new base service
func NewBaseService() *BaseService {
	return &BaseService{}
}
