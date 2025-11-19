package models

import (
	"time"

	"gorm.io/gorm"
)

// Location represents a storage location
type Location struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"uniqueIndex;not null;size:100" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	SubLocations []SubLocation `gorm:"foreignKey:LocationID;constraint:OnDelete:CASCADE" json:"sub_locations,omitempty"`
	Items        []Item        `gorm:"foreignKey:LocationID;constraint:OnDelete:SET NULL" json:"items,omitempty"`
}

// Category represents an item category
type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"uniqueIndex;not null;size:100" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Items []Item `gorm:"foreignKey:CategoryID;constraint:OnDelete:SET NULL" json:"items,omitempty"`
}

// SubLocation represents a sub-location within a location
type SubLocation struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"not null;size:100" json:"name"`
	LocationID uint           `gorm:"not null;index" json:"location_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Location Location `gorm:"foreignKey:LocationID;constraint:OnDelete:CASCADE" json:"location,omitempty"`
	Items    []Item   `gorm:"foreignKey:SubLocationID;constraint:OnDelete:SET NULL" json:"items,omitempty"`
}

// Item represents an inventory item
type Item struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"not null;index;size:255" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	LocationID    *uint          `gorm:"index" json:"location_id"`
	SubLocationID *uint          `gorm:"index" json:"sub_location_id"`
	CategoryID    *uint          `gorm:"index" json:"category_id"`
	Quantity      int            `gorm:"default:1" json:"quantity"`
	ExpiryDate    *time.Time     `gorm:"index" json:"expiry_date"`
	AddedDate     time.Time      `gorm:"autoCreateTime" json:"added_date"`
	Notes         *string        `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships for eager loading
	Location    *Location    `gorm:"foreignKey:LocationID" json:"location,omitempty"`
	SubLocation *SubLocation `gorm:"foreignKey:SubLocationID" json:"sub_location,omitempty"`
	Category    *Category    `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// TableName returns the table name for Location model
func (Location) TableName() string {
	return "locations"
}

// TableName returns the table name for Category model
func (Category) TableName() string {
	return "categories"
}

// TableName returns the table name for SubLocation model
func (SubLocation) TableName() string {
	return "sub_locations"
}

// TableName returns the table name for Item model
func (Item) TableName() string {
	return "items"
}

// BeforeCreate hook for Item
func (i *Item) BeforeCreate(tx *gorm.DB) error {
	if i.AddedDate.IsZero() {
		i.AddedDate = time.Now()
	}
	return nil
}

// Validation methods

// Validate validates the Location data
func (l *Location) Validate() error {
	if l.Name == "" {
		return ErrLocationNameRequired
	}
	if len(l.Name) > 100 {
		return ErrLocationNameTooLong
	}
	return nil
}

// Validate validates the Category data
func (c *Category) Validate() error {
	if c.Name == "" {
		return ErrCategoryNameRequired
	}
	if len(c.Name) > 100 {
		return ErrCategoryNameTooLong
	}
	return nil
}

// Validate validates the SubLocation data
func (s *SubLocation) Validate() error {
	if s.Name == "" {
		return ErrSubLocationNameRequired
	}
	if len(s.Name) > 100 {
		return ErrSubLocationNameTooLong
	}
	if s.LocationID == 0 {
		return ErrSubLocationLocationRequired
	}
	return nil
}

// Validate validates the Item data
func (i *Item) Validate() error {
	if i.Name == "" {
		return ErrItemNameRequired
	}
	if len(i.Name) > 255 {
		return ErrItemNameTooLong
	}
	if i.Quantity < 0 {
		return ErrItemQuantityInvalid
	}
	return nil
}

// Helper methods

// IsExpired checks if the item is expired
func (i *Item) IsExpired() bool {
	if i.ExpiryDate == nil {
		return false
	}
	return time.Now().After(*i.ExpiryDate)
}

// DaysUntilExpiry returns the number of days until expiry
func (i *Item) DaysUntilExpiry() int {
	if i.ExpiryDate == nil {
		return -1 // No expiry date
	}
	days := int(time.Until(*i.ExpiryDate).Hours() / 24)
	if days < 0 {
		return 0 // Already expired
	}
	return days
}

// GetDisplayName returns a formatted display name for the item
func (i *Item) GetDisplayName() string {
	if i.Location != nil && i.SubLocation != nil {
		return i.Name + " (" + i.Location.Name + " - " + i.SubLocation.Name + ")"
	} else if i.Location != nil {
		return i.Name + " (" + i.Location.Name + ")"
	}
	return i.Name
}
