package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Category represents an item category
type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// CategoriesHandler handles all category-related operations
type CategoriesHandler struct {
	DB *sql.DB
}

// NewCategoriesHandler creates a new categories handler
func NewCategoriesHandler(db *sql.DB) *CategoriesHandler {
	return &CategoriesHandler{DB: db}
}

// GetCategories handles GET /api/categories
func (h *CategoriesHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, name FROM categories ORDER BY name")
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "GetCategories",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Error("Failed to query categories")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"handler": "categories",
				"action":  "GetCategories",
				"method":  r.Method,
				"path":    r.URL.Path,
			}).Error("Failed to scan category row")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		categories = append(categories, category)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// CreateCategory handles POST /api/categories
func (h *CategoriesHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "CreateCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Warn("Invalid JSON payload for creating category")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.DB.QueryRow("INSERT INTO categories (name) VALUES ($1) RETURNING id", category.Name).Scan(&category.ID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "CreateCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"name":    category.Name,
		}).Error("Failed to insert category")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

// UpdateCategory handles PUT /api/categories/{id}
func (h *CategoriesHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "UpdateCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      vars["id"],
		}).Warn("Invalid category ID")
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var category Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "UpdateCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Warn("Invalid JSON payload for updating category")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec("UPDATE categories SET name = $1 WHERE id = $2", category.Name, id)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "UpdateCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"name":    category.Name,
		}).Error("Failed to update category")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "UpdateCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Error("Failed to get rows affected for update")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "UpdateCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Warn("Category not found for update")
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	category.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// DeleteCategory handles DELETE /api/categories/{id}
func (h *CategoriesHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "DeleteCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      vars["id"],
		}).Warn("Invalid category ID")
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	force := r.URL.Query().Get("force") == "true"

	// Check for related items
	type relatedItem struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	rows, qerr := h.DB.Query(`SELECT id, name FROM items WHERE category_id = $1 ORDER BY name`, id)
	if qerr != nil {
		logrus.WithError(qerr).WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "DeleteCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "depQuery",
		}).Error("Failed to load related items before delete")
		http.Error(w, qerr.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var related []relatedItem
	for rows.Next() {
		var it relatedItem
		if err := rows.Scan(&it.ID, &it.Name); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"handler": "categories",
				"action":  "DeleteCategory",
				"method":  r.Method,
				"path":    r.URL.Path,
				"id":      id,
				"stage":   "scanRelated",
			}).Error("Failed to scan related item row")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		related = append(related, it)
	}

	if len(related) > 0 && !force {
		logrus.WithFields(logrus.Fields{
			"handler":       "categories",
			"action":        "DeleteCategory",
			"method":        r.Method,
			"path":          r.URL.Path,
			"id":            id,
			"related_count": len(related),
		}).Info("Category has related items, returning conflict for confirmation")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		resp := map[string]interface{}{
			"code":          "HAS_DEPENDENCIES",
			"type":          "category",
			"id":            id,
			"related_items": related,
			"message":       "Des articles sont liés à cette catégorie",
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	// Transactional delete: delete related items first (if any), then the category
	tx, txErr := h.DB.Begin()
	if txErr != nil {
		logrus.WithError(txErr).WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "DeleteCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "txBegin",
		}).Error("Failed to start transaction for delete")
		http.Error(w, txErr.Error(), http.StatusInternalServerError)
		return
	}
	var txErrDefer error
	defer func() {
		if txErrDefer != nil {
			_ = tx.Rollback()
		}
	}()

	if len(related) > 0 {
		if _, txErrDefer = tx.Exec(`DELETE FROM items WHERE category_id = $1`, id); txErrDefer != nil {
			logrus.WithError(txErrDefer).WithFields(logrus.Fields{
				"handler": "categories",
				"action":  "DeleteCategory",
				"method":  r.Method,
				"path":    r.URL.Path,
				"id":      id,
				"stage":   "deleteRelated",
			}).Error("Failed to delete related items before deleting category")
			http.Error(w, txErrDefer.Error(), http.StatusInternalServerError)
			return
		}
	}

	result, derr := tx.Exec(`DELETE FROM categories WHERE id = $1`, id)
	if derr != nil {
		logrus.WithError(derr).WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "DeleteCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "deleteCategory",
		}).Error("Failed to delete category")
		http.Error(w, derr.Error(), http.StatusInternalServerError)
		return
	}
	rowsAffected, aerr := result.RowsAffected()
	if aerr != nil {
		logrus.WithError(aerr).WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "DeleteCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "rowsAffected",
		}).Error("Failed to get rows affected for category delete")
		http.Error(w, aerr.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "DeleteCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
		}).Warn("Category not found for delete")
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	if txErrDefer = tx.Commit(); txErrDefer != nil {
		logrus.WithError(txErrDefer).WithFields(logrus.Fields{
			"handler": "categories",
			"action":  "DeleteCategory",
			"method":  r.Method,
			"path":    r.URL.Path,
			"id":      id,
			"stage":   "txCommit",
		}).Error("Failed to commit transaction for category delete")
		http.Error(w, txErrDefer.Error(), http.StatusInternalServerError)
		return
	}

	logrus.WithFields(logrus.Fields{
		"handler":       "categories",
		"action":        "DeleteCategory",
		"method":        r.Method,
		"path":          r.URL.Path,
		"id":            id,
		"force":         force,
		"related_count": len(related),
	}).Info("Category deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}
