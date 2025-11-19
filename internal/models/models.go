package models

import (
	"errors"
	"time"
)

// Validation errors
var (
	ErrLocationNameRequired        = errors.New("location name is required")
	ErrLocationNameTooLong         = errors.New("location name is too long (max 100 characters)")
	ErrCategoryNameRequired        = errors.New("category name is required")
	ErrCategoryNameTooLong         = errors.New("category name is too long (max 100 characters)")
	ErrSubLocationNameRequired     = errors.New("sub-location name is required")
	ErrSubLocationNameTooLong      = errors.New("sub-location name is too long (max 100 characters)")
	ErrSubLocationLocationRequired = errors.New("sub-location must belong to a location")
	ErrItemNameRequired            = errors.New("item name is required")
	ErrItemNameTooLong             = errors.New("item name is too long (max 255 characters)")
	ErrItemQuantityInvalid         = errors.New("item quantity must be non-negative")
)

// SearchResult represents a search result with scoring
type SearchResult struct {
	Item     Item    `json:"item"`
	Distance int     `json:"distance"`
	Score    float64 `json:"score"`
}

// Legacy DTOs for backward compatibility during transition
// These will be removed once the transition is complete

// ItemDTO represents an item in the legacy format
type ItemDTO struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Category    string    `json:"category"`
	Quantity    int       `json:"quantity"`
	ExpiryDate  *string   `json:"expiry_date"`
	AddedDate   time.Time `json:"added_date"`
}

// LocationDTO represents a location in the legacy format
type LocationDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SubLocationDTO represents a sub-location in the legacy format
type SubLocationDTO struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	LocationID int    `json:"location_id"`
}

// CategoryDTO represents a category in the legacy format
type CategoryDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
