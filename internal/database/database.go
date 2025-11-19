package database

import (
	"fmt"
	"time"

	"jujudb/internal/config"
	"jujudb/internal/models"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database wraps the GORM database connection
type Database struct {
	*gorm.DB
}

// NewDatabase creates a new database connection and performs migrations
func NewDatabase(cfg config.DatabaseConfig) (*Database, error) {
	// Build connection string
	dsn := cfg.GetGormConnectionString()

	// Configure GORM logger
	var gormLogLevel logger.LogLevel
	switch cfg.LogLevel {
	case "silent":
		gormLogLevel = logger.Silent
	case "error":
		gormLogLevel = logger.Error
	case "warn":
		gormLogLevel = logger.Warn
	case "info":
		gormLogLevel = logger.Info
	default:
		gormLogLevel = logger.Info
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Info("Successfully connected to database")

	// Auto-migrate schemas
	if err := migrateDatabase(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Seed initial data
	if err := seedDatabase(db); err != nil {
		return nil, fmt.Errorf("failed to seed database: %w", err)
	}

	return &Database{DB: db}, nil
}

// migrateDatabase performs auto-migration of all models
func migrateDatabase(db *gorm.DB) error {
	logrus.Info("Running database migrations")

	// Auto-migrate all models in order
	models := []interface{}{
		&models.Location{},
		&models.Category{},
		&models.SubLocation{},
		&models.Item{},
	}

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	// Create additional indexes if needed
	if err := createAdditionalIndexes(db); err != nil {
		return fmt.Errorf("failed to create additional indexes: %w", err)
	}

	logrus.Info("Database migrations completed successfully")
	return nil
}

// createAdditionalIndexes creates any additional indexes not covered by GORM tags
func createAdditionalIndexes(db *gorm.DB) error {
	// Create composite indexes for better query performance
	indexes := []struct {
		table   string
		columns []string
		name    string
	}{
		{"items", []string{"location_id", "sub_location_id"}, "idx_items_location_sublocation"},
		{"items", []string{"category_id", "expiry_date"}, "idx_items_category_expiry"},
		{"items", []string{"name", "added_date"}, "idx_items_name_added"},
	}

	for _, idx := range indexes {
		// Check if index exists
		var exists bool
		err := db.Raw(`
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes 
				WHERE tablename = ? AND indexname = ?
			)`, idx.table, idx.name).Scan(&exists).Error

		if err != nil {
			return fmt.Errorf("failed to check index %s: %w", idx.name, err)
		}

		if !exists {
			// Create index
			columnsStr := ""
			for i, col := range idx.columns {
				if i > 0 {
					columnsStr += ", "
				}
				columnsStr += col
			}

			err := db.Exec(fmt.Sprintf(
				"CREATE INDEX %s ON %s (%s)",
				idx.name, idx.table, columnsStr,
			)).Error

			if err != nil {
				return fmt.Errorf("failed to create index %s: %w", idx.name, err)
			}

			logrus.WithField("index", idx.name).Debug("Created additional index")
		}
	}

	return nil
}

// seedDatabase seeds the database with initial data
func seedDatabase(db *gorm.DB) error {
	logrus.Info("Seeding database with initial data")

	// Seed locations
	locations := []models.Location{
		{Name: "Congélateur"},
		{Name: "Réfrigérateur"},
		{Name: "Garde-manger"},
	}

	for _, location := range locations {
		var existing models.Location
		err := db.Where("name = ?", location.Name).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&location).Error; err != nil {
				return fmt.Errorf("failed to seed location %s: %w", location.Name, err)
			}
			logrus.WithField("location", location.Name).Debug("Seeded location")
		}
	}

	// Seed categories
	categories := []models.Category{
		{Name: "Viande"},
		{Name: "Légumes"},
		{Name: "Desserts"},
		{Name: "Poisson"},
		{Name: "Autres"},
	}

	for _, category := range categories {
		var existing models.Category
		err := db.Where("name = ?", category.Name).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&category).Error; err != nil {
				return fmt.Errorf("failed to seed category %s: %w", category.Name, err)
			}
			logrus.WithField("category", category.Name).Debug("Seeded category")
		}
	}

	// Seed sample items (only if no items exist)
	var itemCount int64
	db.Model(&models.Item{}).Count(&itemCount)
	if itemCount == 0 {
		// Get seeded locations and categories
		var freezer models.Location
		var fridge models.Location
		var pantry models.Location

		db.Where("name = ?", "Congélateur").First(&freezer)
		db.Where("name = ?", "Réfrigérateur").First(&fridge)
		db.Where("name = ?", "Garde-manger").First(&pantry)

		var meatCategory models.Category
		var vegCategory models.Category
		var dessertCategory models.Category
		var fishCategory models.Category
		var otherCategory models.Category

		db.Where("name = ?", "Viande").First(&meatCategory)
		db.Where("name = ?", "Légumes").First(&vegCategory)
		db.Where("name = ?", "Desserts").First(&dessertCategory)
		db.Where("name = ?", "Poisson").First(&fishCategory)
		db.Where("name = ?", "Autres").First(&otherCategory)

		// Create sample items
		sampleItems := []models.Item{
			{
				Name:        "Steaks de bœuf",
				Description: "Steaks de bœuf congelés, 4 pièces",
				LocationID:  &freezer.ID,
				CategoryID:  &meatCategory.ID,
				Quantity:    4,
				Notes:       &[]string{"Qualité supérieure"}[0],
			},
			{
				Name:        "Haricots verts",
				Description: "Haricots verts surgelés",
				LocationID:  &freezer.ID,
				CategoryID:  &vegCategory.ID,
				Quantity:    1,
			},
			{
				Name:        "Glace vanille",
				Description: "Bac de glace à la vanille",
				LocationID:  &freezer.ID,
				CategoryID:  &dessertCategory.ID,
				Quantity:    1,
			},
			{
				Name:        "Saumon fumé",
				Description: "Tranches de saumon fumé",
				LocationID:  &fridge.ID,
				CategoryID:  &fishCategory.ID,
				Quantity:    1,
			},
			{
				Name:        "Pâtes",
				Description: "Paquet de pâtes italiennes",
				LocationID:  &pantry.ID,
				CategoryID:  &otherCategory.ID,
				Quantity:    2,
			},
		}

		for _, item := range sampleItems {
			if err := db.Create(&item).Error; err != nil {
				return fmt.Errorf("failed to seed item %s: %w", item.Name, err)
			}
			logrus.WithField("item", item.Name).Debug("Seeded sample item")
		}
	}

	logrus.Info("Database seeding completed")
	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping tests the database connection
func (d *Database) Ping() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// GetStats returns database statistics
func (d *Database) GetStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	tables := []string{"locations", "categories", "sub_locations", "items"}

	for _, table := range tables {
		var count int64
		if err := d.Table(table).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to get count for table %s: %w", table, err)
		}
		stats[table] = count
	}

	return stats, nil
}
