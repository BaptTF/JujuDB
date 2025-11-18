package models

import (
	"time"
)

// Item represents an inventory item
type Item struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Category    string    `json:"category"`
	Quantity    int       `json:"quantity"`
	ExpiryDate  *string   `json:"expiry_date"`
	AddedDate   time.Time `json:"added_date"`
}

// SearchResult represents a search result with scoring
type SearchResult struct {
	Item     Item    `json:"item"`
	Distance int     `json:"distance"`
	Score    float64 `json:"score"`
}

// Location represents a storage location
type Location struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SubLocation represents a sub-location within a location
type SubLocation struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	LocationID int    `json:"location_id"`
}

// Category represents an item category
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
