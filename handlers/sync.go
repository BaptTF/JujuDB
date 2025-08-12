package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"jujudb/services"
)

// SyncHandler handles synchronization operations
type SyncHandler struct {
	Sync *services.SyncService
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(syncService *services.SyncService) *SyncHandler {
	return &SyncHandler{Sync: syncService}
}

// SyncAll handles POST /api/sync/all - syncs all items to Meilisearch
func (h *SyncHandler) SyncAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	logrus.Info("Starting full sync to Meilisearch")
	
	err := h.Sync.SyncAllItems()
	if err != nil {
		logrus.WithError(err).Error("Failed to sync all items to Meilisearch")
		http.Error(w, "Sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": "All items synced to Meilisearch successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	
	logrus.Info("Full sync to Meilisearch completed successfully")
}
