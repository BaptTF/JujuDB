package services

import (
	"fmt"

	"jujudb/internal/models"
	"jujudb/internal/repositories"

	"github.com/sirupsen/logrus"
)

// categoryService implements CategoryService interface
type categoryService struct {
	repo repositories.CategoryRepository
	*BaseService
}

// NewCategoryService creates a new category service
func NewCategoryService(repo repositories.CategoryRepository) CategoryService {
	return &categoryService{
		repo:        repo,
		BaseService: NewBaseService(),
	}
}

// CreateCategory creates a new category with validation
func (s *categoryService) CreateCategory(category *models.Category) error {
	// Validate category
	if err := s.ValidateCategory(category); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create category
	if err := s.repo.Create(category); err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	logrus.WithField("category_id", category.ID).Info("Category created successfully")
	return nil
}

// GetCategory retrieves a category by ID
func (s *categoryService) GetCategory(id uint) (*models.Category, error) {
	return s.repo.GetByID(id)
}

// GetCategoryWithRelations retrieves a category by ID with its relations
func (s *categoryService) GetCategoryWithRelations(id uint) (*models.Category, error) {
	return s.repo.GetByIDWithRelations(id)
}

// UpdateCategory updates an existing category with validation
func (s *categoryService) UpdateCategory(category *models.Category) error {
	// Validate category
	if err := s.ValidateCategory(category); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check if category exists
	existing, err := s.repo.GetByID(category.ID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Update category
	if err := s.repo.Update(category); err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"category_id": category.ID,
		"old_name":    existing.Name,
		"new_name":    category.Name,
	}).Info("Category updated successfully")
	return nil
}

// DeleteCategory deletes a category with dependency checking
func (s *categoryService) DeleteCategory(id uint, force bool) error {
	// Check if category exists
	category, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Check dependencies
	canDelete, err := s.CanDeleteCategory(id)
	if err != nil {
		return fmt.Errorf("failed to check if category can be deleted: %w", err)
	}

	if !canDelete && !force {
		// Return dependencies information
		deps, err := s.GetCategoryDependencies(id)
		if err != nil {
			return fmt.Errorf("failed to get category dependencies: %w", err)
		}
		return fmt.Errorf("category has %d items", deps.ItemsCount)
	}

	if !canDelete && force {
		// Force delete - this will cascade delete dependencies
		logrus.WithFields(logrus.Fields{
			"category_id": id,
			"force":       true,
		}).Warn("Force deleting category with dependencies")
	}

	// Delete category (cascade will handle dependencies if force is true)
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"category_id":   id,
		"category_name": category.Name,
		"force":         force,
	}).Info("Category deleted successfully")
	return nil
}

// GetCategories retrieves categories with optional filters
func (s *categoryService) GetCategories(filters repositories.CategoryFilters) ([]models.Category, error) {
	return s.repo.GetAll(filters)
}

// CountCategories returns the count of categories matching the filters
func (s *categoryService) CountCategories(filters repositories.CategoryFilters) (int64, error) {
	return s.repo.Count(filters)
}

// GetCategoryWithItems retrieves a category with its items
func (s *categoryService) GetCategoryWithItems(id uint) (*models.Category, error) {
	return s.repo.GetWithItems(id)
}

// ValidateCategory validates a category's data
func (s *categoryService) ValidateCategory(category *models.Category) error {
	return category.Validate()
}

// CanDeleteCategory checks if a category can be deleted
func (s *categoryService) CanDeleteCategory(id uint) (bool, error) {
	// Check if category has items
	hasItems, err := s.repo.HasItems(id)
	if err != nil {
		return false, fmt.Errorf("failed to check items: %w", err)
	}

	// Category can be deleted only if it has no items
	return !hasItems, nil
}

// GetCategoryDependencies returns the dependencies of a category
func (s *categoryService) GetCategoryDependencies(id uint) (*CategoryDependencies, error) {
	deps := &CategoryDependencies{}

	// Get count
	itemsCount, err := s.repo.GetRelatedItemsCount(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get items count: %w", err)
	}
	deps.ItemsCount = itemsCount

	// If there are dependencies, get the actual data
	if deps.ItemsCount > 0 {
		category, err := s.repo.GetByIDWithRelations(id)
		if err != nil {
			return nil, fmt.Errorf("failed to get category with relations: %w", err)
		}

		deps.Items = category.Items
	}

	return deps, nil
}

// SearchCategories searches categories by name
func (s *categoryService) SearchCategories(query string) ([]models.Category, error) {
	return s.repo.SearchByName(query)
}
