package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// MeilisearchService handles Meilisearch operations using HTTP API
type MeilisearchService struct {
	host      string
	apiKey    string
	indexName string
	client    *http.Client
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

// NewMeilisearchService creates a new Meilisearch service
func NewMeilisearchService(host, masterKey string) (*MeilisearchService, error) {
	service := &MeilisearchService{
		host:      host,
		apiKey:    masterKey,
		indexName: "items",
		client:    &http.Client{Timeout: 30 * time.Second},
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
	req, err := http.NewRequest("GET", s.host+"/health", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}
	return nil
}

// createIndex creates the Meilisearch index if it doesn't exist
func (s *MeilisearchService) createIndex() error {
	// Check if index exists
	req, err := http.NewRequest("GET", s.host+"/indexes/"+s.indexName, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		logrus.Info("Meilisearch index already exists")
		return nil
	}

	// Create index
	logrus.Info("Creating Meilisearch index: items")
	indexConfig := map[string]interface{}{
		"uid":        s.indexName,
		"primaryKey": "id",
	}

	body, err := json.Marshal(indexConfig)
	if err != nil {
		return err
	}

	req, err = http.NewRequest("POST", s.host+"/indexes", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err = s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to create index, status: %d", resp.StatusCode)
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
	err := s.updateSettings("searchableAttributes", searchableAttrs)
	if err != nil {
		return fmt.Errorf("failed to update searchable attributes: %w", err)
	}

	// Set filterable attributes (fields that can be used for filtering)
	filterableAttrs := []string{
		"location_id",
		"sub_location_id",
		"category_id",
		"expiry_date",
		"added_date",
	}
	err = s.updateSettings("filterableAttributes", filterableAttrs)
	if err != nil {
		return fmt.Errorf("failed to update filterable attributes: %w", err)
	}

	// Set sortable attributes
	sortableAttrs := []string{
		"added_date",
		"expiry_date",
		"name",
	}
	err = s.updateSettings("sortableAttributes", sortableAttrs)
	if err != nil {
		return fmt.Errorf("failed to update sortable attributes: %w", err)
	}

	logrus.Info("Meilisearch index configured successfully")
	return nil
}

// updateSettings updates a specific setting for the index
func (s *MeilisearchService) updateSettings(settingName string, value interface{}) error {
	settings := map[string]interface{}{
		settingName: value,
	}

	body, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", s.host+"/indexes/"+s.indexName+"/settings", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to update %s, status: %d", settingName, resp.StatusCode)
	}

	return nil
}

// IndexItem adds or updates an item in the search index
func (s *MeilisearchService) IndexItem(item SearchableItem) error {
	documents := []SearchableItem{item}
	return s.IndexItems(documents)
}

// IndexItems adds or updates multiple items in the search index
func (s *MeilisearchService) IndexItems(items []SearchableItem) error {
	if len(items) == 0 {
		return nil
	}

	body, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	req, err := http.NewRequest("POST", s.host+"/indexes/"+s.indexName+"/documents", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to index items: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to index items, status: %d", resp.StatusCode)
	}

	logrus.WithField("items_count", len(items)).Info("Items indexed in Meilisearch")
	return nil
}

// DeleteItem removes an item from the search index
func (s *MeilisearchService) DeleteItem(itemID int) error {
	req, err := http.NewRequest("DELETE", s.host+"/indexes/"+s.indexName+"/documents/"+strconv.Itoa(itemID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete item from index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("failed to delete item, status: %d", resp.StatusCode)
	}

	logrus.WithField("item_id", itemID).Debug("Item deleted from Meilisearch")
	return nil
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

// SearchResponse represents Meilisearch search response
type SearchResponse struct {
	Hits                 []map[string]interface{} `json:"hits"`
	EstimatedTotalHits   int64                    `json:"estimatedTotalHits"`
	ProcessingTimeMs     int64                    `json:"processingTimeMs"`
}

// Search performs a search query with filters
func (s *MeilisearchService) Search(req SearchRequest) ([]SearchableItem, error) {
	// Build search request
	searchReq := map[string]interface{}{
		"q":      req.Query,
		"limit":  req.Limit,
		"offset": req.Offset,
		"sort":   []string{"added_date:desc"},
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
		searchReq["filter"] = strings.Join(filters, " AND ")
	}

	// Marshal search request
	body, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	// Perform search
	httpReq, err := http.NewRequest("POST", s.host+"/indexes/"+s.indexName+"/search", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search failed with status: %d", resp.StatusCode)
	}

	// Parse response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var searchResp SearchResponse
	if err := json.Unmarshal(respBody, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Convert results
	var items []SearchableItem
	for _, hit := range searchResp.Hits {
		var item SearchableItem

		// Parse fields from hit
		if id, ok := hit["id"].(float64); ok {
			item.ID = int(id)
		} else if idStr, ok := hit["id"].(string); ok {
			if parsedID, err := strconv.Atoi(idStr); err == nil {
				item.ID = parsedID
			}
		}
		if name, ok := hit["name"].(string); ok {
			item.Name = name
		}
		if desc, ok := hit["description"].(string); ok {
			item.Description = desc
		}
		if loc, ok := hit["location"].(string); ok {
			item.Location = loc
		}
		if subLoc, ok := hit["sub_location"].(string); ok {
			item.SubLocation = subLoc
		}
		if cat, ok := hit["category"].(string); ok {
			item.Category = cat
		}
		if qty, ok := hit["quantity"].(float64); ok {
			item.Quantity = int(qty)
		}
		if notes, ok := hit["notes"].(string); ok && notes != "" {
			item.Notes = &notes
		}
		if expiryDate, ok := hit["expiry_date"].(string); ok && expiryDate != "" {
			item.ExpiryDate = &expiryDate
		}

		// Handle nullable integer fields
		if locID, ok := hit["location_id"].(float64); ok {
			id := int(locID)
			item.LocationID = &id
		}
		if subLocID, ok := hit["sub_location_id"].(float64); ok {
			id := int(subLocID)
			item.SubLocationID = &id
		}
		if catID, ok := hit["category_id"].(float64); ok {
			id := int(catID)
			item.CategoryID = &id
		}

		// Handle added_date
		if addedDate, ok := hit["added_date"].(string); ok {
			if parsedDate, err := time.Parse(time.RFC3339, addedDate); err == nil {
				item.AddedDate = parsedDate
			}
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
