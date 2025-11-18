// Categories Management Module
class CategoriesManager {
    constructor(app) {
        this.app = app;
        this.bindEvents();
    }

    bindEvents() {
        // Add category button (only on locations page)
        const addCategoryBtn = document.getElementById('addCategoryBtn');
        if (addCategoryBtn) {
            addCategoryBtn.addEventListener('click', () => {
                this.addCategory();
            });
        }
        
        // Category filter change (only on dashboard)
        const categoryFilter = document.getElementById('categoryFilter');
        if (categoryFilter) {
            categoryFilter.addEventListener('change', () => {
                if (this.app.filters) {
                    this.app.filters.applyFilters();
                }
            });
        }
    }

    async loadCategories() {
        try {
            const response = await fetch('/api/categories');
            if (response.ok) {
                this.app.categories = await response.json();
                this.populateCategorySelects();
                // Re-render categories list if available
                if (typeof this.renderCategoriesList === 'function') {
                    this.renderCategoriesList();
                }
            }
        } catch (error) {
            console.error('Error loading categories:', error);
        }
    }
    
    populateCategorySelects() {
        const filterSelect = document.getElementById('categoryFilter');
        const itemSelect = document.getElementById('itemCategory');
        
        // Clear existing options (keep first option)
        [filterSelect, itemSelect].forEach(select => {
            if (select) {
                while (select.children.length > 1) {
                    select.removeChild(select.lastChild);
                }
            }
        });
        
        // Populate with categories
        this.app.categories.forEach(category => {
            if (filterSelect) {
                const option = document.createElement('option');
                option.value = category.id;
                option.textContent = category.name;
                filterSelect.appendChild(option);
            }
            
            if (itemSelect) {
                const option = document.createElement('option');
                option.value = category.id;
                option.textContent = category.name;
                itemSelect.appendChild(option);
            }
        });
    }

    renderCategoriesList() {
        // Render categories list
        const categoriesList = document.getElementById('categoriesList');
        if (!categoriesList) return;
        
        categoriesList.innerHTML = this.app.categories.map(category => `
            <div class="list-item">
                <span class="list-item-name">${category.name}</span>
                <button class="btn btn-sm btn-danger" onclick="window.locationsPage.categoriesManager.deleteCategory(${category.id})">Supprimer</button>
            </div>
        `).join('');
    }

    async addCategory() {
        const name = document.getElementById('newCategoryName').value.trim();
        if (!name) {
            this.app.showError('Le nom de la catégorie est obligatoire');
            return;
        }
        
        try {
            const response = await fetch('/api/categories', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ name })
            });
            
            if (response.ok) {
                document.getElementById('newCategoryName').value = '';
                await this.app.locationsManager.loadLocationManagementData();
                this.app.showSuccess('Catégorie ajoutée avec succès');
            } else {
                throw new Error('Erreur lors de l\'ajout');
            }
        } catch (error) {
            console.error('Error adding category:', error);
            this.app.showError('Erreur lors de l\'ajout de la catégorie');
        }
    }

    async deleteCategory(id) {
        if (!confirm('Êtes-vous sûr de vouloir supprimer cette catégorie ?')) {
            return;
        }

        const doDelete = async (force = false) => {
            const url = force ? `/api/categories/${id}?force=true` : `/api/categories/${id}`;
            const response = await fetch(url, { method: 'DELETE' });
            return response;
        };

        try {
            let response = await doDelete(false);
            if (response.ok) {
                await this.app.locationsManager.loadLocationManagementData();
                this.app.showSuccess('Catégorie supprimée avec succès');
                return;
            }

            if (response.status === 409) {
                const data = await response.json().catch(() => ({}));
                window.locationsPage.showConfirmDelete({
                    type: 'category',
                    id,
                    related_items: data.related_items || []
                }, async () => {
                    const forceResp = await doDelete(true);
                    if (forceResp.ok) {
                        await this.app.locationsManager.loadLocationManagementData();
                        this.app.showSuccess('Catégorie et éléments liés supprimés');
                    } else {
                        this.app.showError('Échec de la suppression forcée de la catégorie');
                    }
                });
                return;
            }

            throw new Error('Erreur lors de la suppression');
        } catch (error) {
            console.error('Error deleting category:', error);
            this.app.showError('Erreur lors de la suppression de la catégorie');
        }
    }
}
