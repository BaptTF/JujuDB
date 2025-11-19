package repositories

import (
	"jujudb/internal/models"
	"time"

	"gorm.io/gorm"
)

// ItemFilters defines filters for item queries
type ItemFilters struct {
	LocationID    *uint
	SubLocationID *uint
	CategoryID    *uint
	Name          string
	ExpiryBefore  *time.Time
	ExpiryAfter   *time.Time
	QuantityMin   *int
	QuantityMax   *int
	Limit         int
	Offset        int
	OrderBy       string
	OrderDir      string // "asc" or "desc"
}

// LocationFilters defines filters for location queries
type LocationFilters struct {
	Name   string
	Limit  int
	Offset int
}

// CategoryFilters defines filters for category queries
type CategoryFilters struct {
	Name   string
	Limit  int
	Offset int
}

// SubLocationFilters defines filters for sub-location queries
type SubLocationFilters struct {
	LocationID *uint
	Name       string
	Limit      int
	Offset     int
}

// ItemRepository defines the interface for item database operations
type ItemRepository interface {
	// Basic CRUD operations
	Create(item *models.Item) error
	GetByID(id uint) (*models.Item, error)
	GetByIDWithRelations(id uint) (*models.Item, error)
	Update(item *models.Item) error
	Delete(id uint) error
	SoftDelete(id uint) error

	// Query operations
	GetAll(filters ItemFilters) ([]models.Item, error)
	GetAllWithRelations(filters ItemFilters) ([]models.Item, error)
	Count(filters ItemFilters) (int64, error)

	// Bulk operations
	CreateBatch(items []models.Item) error
	UpdateBatch(items []models.Item) error
	DeleteBatch(ids []uint) error

	// Search and filtering
	SearchByName(query string, filters ItemFilters) ([]models.Item, error)
	GetExpiringItems(days int) ([]models.Item, error)
	GetLowStockItems(threshold int) ([]models.Item, error)

	// Relationship operations
	GetByLocationID(locationID uint) ([]models.Item, error)
	GetByCategoryID(categoryID uint) ([]models.Item, error)
	GetBySubLocationID(subLocationID uint) ([]models.Item, error)
}

// LocationRepository defines the interface for location database operations
type LocationRepository interface {
	// Basic CRUD operations
	Create(location *models.Location) error
	GetByID(id uint) (*models.Location, error)
	GetByIDWithRelations(id uint) (*models.Location, error)
	Update(location *models.Location) error
	Delete(id uint) error
	SoftDelete(id uint) error

	// Query operations
	GetAll(filters LocationFilters) ([]models.Location, error)
	Count(filters LocationFilters) (int64, error)

	// Relationship operations
	GetWithItems(id uint) (*models.Location, error)
	GetWithSubLocations(id uint) (*models.Location, error)

	// Validation and constraints
	HasItems(id uint) (bool, error)
	HasSubLocations(id uint) (bool, error)
	GetRelatedItemsCount(id uint) (int64, error)
	GetRelatedSubLocationsCount(id uint) (int64, error)

	// Search
	SearchByName(query string) ([]models.Location, error)
}

// CategoryRepository defines the interface for category database operations
type CategoryRepository interface {
	// Basic CRUD operations
	Create(category *models.Category) error
	GetByID(id uint) (*models.Category, error)
	GetByIDWithRelations(id uint) (*models.Category, error)
	Update(category *models.Category) error
	Delete(id uint) error
	SoftDelete(id uint) error

	// Query operations
	GetAll(filters CategoryFilters) ([]models.Category, error)
	Count(filters CategoryFilters) (int64, error)

	// Relationship operations
	GetWithItems(id uint) (*models.Category, error)

	// Validation and constraints
	HasItems(id uint) (bool, error)
	GetRelatedItemsCount(id uint) (int64, error)

	// Search
	SearchByName(query string) ([]models.Category, error)
}

// SubLocationRepository defines the interface for sub-location database operations
type SubLocationRepository interface {
	// Basic CRUD operations
	Create(subLocation *models.SubLocation) error
	GetByID(id uint) (*models.SubLocation, error)
	GetByIDWithRelations(id uint) (*models.SubLocation, error)
	Update(subLocation *models.SubLocation) error
	Delete(id uint) error
	SoftDelete(id uint) error

	// Query operations
	GetAll(filters SubLocationFilters) ([]models.SubLocation, error)
	Count(filters SubLocationFilters) (int64, error)

	// Relationship operations
	GetWithItems(id uint) (*models.SubLocation, error)
	GetByLocationID(locationID uint) ([]models.SubLocation, error)

	// Validation and constraints
	HasItems(id uint) (bool, error)
	GetRelatedItemsCount(id uint) (int64, error)

	// Search
	SearchByName(query string, locationID *uint) ([]models.SubLocation, error)
}

// Repository provides access to all repositories
type Repository struct {
	Items        ItemRepository
	Locations    LocationRepository
	Categories   CategoryRepository
	SubLocations SubLocationRepository
}

// NewRepository creates a new repository with all sub-repositories
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Items:        NewItemRepository(db),
		Locations:    NewLocationRepository(db),
		Categories:   NewCategoryRepository(db),
		SubLocations: NewSubLocationRepository(db),
	}
}

// BaseRepository provides common functionality for all repositories
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// GetDB returns the underlying GORM database instance
func (r *BaseRepository) GetDB() *gorm.DB {
	return r.db
}

// Transaction executes a function within a database transaction
func (r *BaseRepository) Transaction(fn func(*gorm.DB) error) error {
	return r.db.Transaction(fn)
}

// Begin starts a new transaction
func (r *BaseRepository) Begin() *gorm.DB {
	return r.db.Begin()
}

// Pagination helper
func applyPagination(query *gorm.DB, limit, offset int) *gorm.DB {
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	return query
}

// Ordering helper
func applyOrdering(query *gorm.DB, orderBy, orderDir string) *gorm.DB {
	if orderBy != "" {
		if orderDir == "desc" {
			query = query.Order(orderBy + " desc")
		} else {
			query = query.Order(orderBy + " asc")
		}
	}
	return query
}
