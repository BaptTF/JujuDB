package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/sirupsen/logrus"
)

// MeilisearchService handles Meilisearch operations using official client
type MeilisearchService struct {
	client    meilisearch.ServiceManager
	index     meilisearch.IndexManager
	indexName string
}

// SearchableItem represents an item in Meilisearch index
type SearchableItem struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	LocationID    *int      `json:"location_id"`
	SubLocationID *int      `json:"sub_location_id"`
	CategoryID    *int      `json:"category_id"`
	Location      string    `json:"location"`
	SubLocation   string    `json:"sub_location"`
	Category      string    `json:"category"`
	Quantity      int       `json:"quantity"`
	ExpiryDate    *string   `json:"expiry_date"`
	AddedDate     time.Time `json:"added_date"`
	Notes         *string   `json:"notes"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query         string
	LocationID    string
	SubLocationID string
	CategoryID    string
	Limit         int
	Offset        int
}

// NewMeilisearchService creates a new Meilisearch service
func NewMeilisearchService(host, masterKey string) (*MeilisearchService, error) {
	// Create Meilisearch client
	client := meilisearch.New(host, meilisearch.WithAPIKey(masterKey))

	indexName := "items"
	index := client.Index(indexName)

	service := &MeilisearchService{
		client:    client,
		index:     index,
		indexName: indexName,
	}

	// Test connection
	err := service.testConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Meilisearch: %w", err)
	}

	logrus.Info("Connected to Meilisearch")

	// Create index if it doesn't exist
	err = service.createIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	// Configure index settings
	err = service.configureIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to configure index: %w", err)
	}

	return service, nil
}

// testConnection tests the connection to Meilisearch
func (s *MeilisearchService) testConnection() error {
	_, err := s.client.Health()
	return err
}

// createIndex creates the Meilisearch index if it doesn't exist
func (s *MeilisearchService) createIndex() error {
	// Check if index exists
	_, err := s.index.GetStats()
	if err == nil {
		logrus.Info("Meilisearch index already exists")
		return nil
	}

	// Create index
	logrus.Info("Creating Meilisearch index: items")
	_, err = s.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        s.indexName,
		PrimaryKey: "id",
	})
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// Wait a bit for index creation
	time.Sleep(2 * time.Second)
	return nil
}

// configureIndex sets up searchable attributes and filters
func (s *MeilisearchService) configureIndex() error {
	// Set searchable attributes (fields that will be searched)
	searchableAttrs := []string{
		"name",
		"description",
		"location",
		"sub_location",
		"category",
		"notes",
	}

	// Set filterable attributes (fields that can be used for filtering)
	filterableAttrs := []string{
		"location_id",
		"sub_location_id",
		"category_id",
		"expiry_date",
		"added_date",
	}

	// Set sortable attributes
	sortableAttrs := []string{
		"added_date",
		"expiry_date",
		"name",
	}

	// Convert string slices to interface slices only for filterable attributes
	filterableInterfaces := make([]interface{}, len(filterableAttrs))
	for i, v := range filterableAttrs {
		filterableInterfaces[i] = v
	}

	// Update settings - different methods expect different types
	_, err := s.index.UpdateSearchableAttributes(&searchableAttrs)
	if err != nil {
		return fmt.Errorf("failed to update searchable attributes: %w", err)
	}

	_, err = s.index.UpdateFilterableAttributes(&filterableInterfaces)
	if err != nil {
		return fmt.Errorf("failed to update filterable attributes: %w", err)
	}

	_, err = s.index.UpdateSortableAttributes(&sortableAttrs)
	if err != nil {
		return fmt.Errorf("failed to update sortable attributes: %w", err)
	}

	logrus.Info("Meilisearch index configured successfully")
	return nil
}

// IndexItem adds or updates an item in the search index
func (s *MeilisearchService) IndexItem(item SearchableItem) error {
	return s.IndexItems([]SearchableItem{item})
}

// IndexItems adds or updates multiple items in the search index
func (s *MeilisearchService) IndexItems(items []SearchableItem) error {
	if len(items) == 0 {
		return nil
	}

	// Convert items to interface slice for Meilisearch
	docs := make([]interface{}, len(items))
	for i, item := range items {
		docs[i] = item
	}

	_, err := s.index.AddDocuments(docs, nil)
	if err != nil {
		return fmt.Errorf("failed to index items: %w", err)
	}

	logrus.WithField("count", len(items)).Debug("Items indexed successfully")
	return nil
}

// DeleteItem removes an item from the search index
func (s *MeilisearchService) DeleteItem(itemID int) error {
	_, err := s.index.DeleteDocument(strconv.Itoa(itemID))
	if err != nil {
		return fmt.Errorf("failed to delete item from index: %w", err)
	}

	logrus.WithField("item_id", itemID).Debug("Item deleted from index")
	return nil
}

// Search performs a search query with filters
func (s *MeilisearchService) Search(req SearchRequest) ([]SearchableItem, error) {
	// Build search request
	searchReq := &meilisearch.SearchRequest{
		Limit:  int64(req.Limit),
		Offset: int64(req.Offset),
		Sort:   []string{"added_date:desc"},
	}

	// Build filters
	var filters []string
	if req.LocationID != "" {
		filters = append(filters, fmt.Sprintf("location_id = %s", req.LocationID))
	}
	if req.SubLocationID != "" {
		filters = append(filters, fmt.Sprintf("sub_location_id = %s", req.SubLocationID))
	}
	if req.CategoryID != "" {
		filters = append(filters, fmt.Sprintf("category_id = %s", req.CategoryID))
	}

	if len(filters) > 0 {
		filterStr := strings.Join(filters, " AND ")
		searchReq.Filter = filterStr
	}

	// Perform search
	searchResp, err := s.index.Search(req.Query, searchReq)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	// Convert results
	var items []SearchableItem
	for _, hit := range searchResp.Hits {
		var item SearchableItem

		// Convert hit to JSON and then unmarshal to our struct
		hitBytes, err := json.Marshal(hit)
		if err != nil {
			logrus.WithError(err).Warning("Failed to marshal search hit")
			continue
		}

		if err := json.Unmarshal(hitBytes, &item); err != nil {
			logrus.WithError(err).Warning("Failed to unmarshal search hit")
			continue
		}

		items = append(items, item)
	}

	logrus.WithFields(logrus.Fields{
		"query":        req.Query,
		"results":      len(items),
		"total_hits":   searchResp.EstimatedTotalHits,
		"processing":   fmt.Sprintf("%dms", searchResp.ProcessingTimeMs),
	}).Debug("Meilisearch query completed")

	return items, nil
}
