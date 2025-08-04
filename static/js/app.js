// JujuDB Frontend Application
class JujuDB {
    constructor() {
        this.items = [];
        this.filteredItems = [];
        this.isSearchMode = false;
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadItems();
    }

    bindEvents() {
        // Search functionality
        document.getElementById('searchBtn').addEventListener('click', () => this.performSearch());
        document.getElementById('searchInput').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') this.performSearch();
        });

        // Filters
        document.getElementById('locationFilter').addEventListener('change', () => this.applyFilters());
        document.getElementById('categoryFilter').addEventListener('change', () => this.applyFilters());
        document.getElementById('clearFilters').addEventListener('click', () => this.clearFilters());

        // Add item button
        document.getElementById('addItemBtn').addEventListener('click', () => this.openAddModal());

        // Modal events
        document.querySelector('.close').addEventListener('click', () => this.closeModal());
        document.getElementById('cancelBtn').addEventListener('click', () => this.closeModal());
        document.getElementById('itemForm').addEventListener('submit', (e) => this.handleFormSubmit(e));

        // Close modal when clicking outside
        window.addEventListener('click', (e) => {
            const modal = document.getElementById('itemModal');
            if (e.target === modal) {
                this.closeModal();
            }
        });
    }

    async loadItems() {
        try {
            this.showLoading();
            const response = await fetch('/api/items');
            if (!response.ok) throw new Error('Erreur lors du chargement des articles');
            
            this.items = await response.json();
            this.filteredItems = [...this.items];
            this.isSearchMode = false;
            this.renderItems();
        } catch (error) {
            console.error('Erreur:', error);
            this.showError('Erreur lors du chargement des articles');
        }
    }

    async performSearch() {
        const query = document.getElementById('searchInput').value.trim();
        if (!query) {
            this.loadItems();
            return;
        }

        try {
            this.showLoading();
            const locationFilter = document.getElementById('locationFilter').value;
            const categoryFilter = document.getElementById('categoryFilter').value;
            
            let url = `/api/search?q=${encodeURIComponent(query)}`;
            if (locationFilter) url += `&location=${encodeURIComponent(locationFilter)}`;
            if (categoryFilter) url += `&category=${encodeURIComponent(categoryFilter)}`;

            const response = await fetch(url);
            if (!response.ok) throw new Error('Erreur lors de la recherche');
            
            const searchResults = await response.json();
            this.filteredItems = searchResults.map(result => result.item);
            this.isSearchMode = true;
            this.renderItems(searchResults);
        } catch (error) {
            console.error('Erreur:', error);
            this.showError('Erreur lors de la recherche');
        }
    }

    applyFilters() {
        if (this.isSearchMode) {
            this.performSearch();
            return;
        }

        const locationFilter = document.getElementById('locationFilter').value;
        const categoryFilter = document.getElementById('categoryFilter').value;

        this.filteredItems = this.items.filter(item => {
            const locationMatch = !locationFilter || item.location === locationFilter;
            const categoryMatch = !categoryFilter || item.category === categoryFilter;
            return locationMatch && categoryMatch;
        });

        this.renderItems();
    }

    clearFilters() {
        document.getElementById('searchInput').value = '';
        document.getElementById('locationFilter').value = '';
        document.getElementById('categoryFilter').value = '';
        this.loadItems();
    }

    renderItems(searchResults = null) {
        const container = document.getElementById('itemsContainer');
        
        if (this.filteredItems.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <h3>Aucun article trouvé</h3>
                    <p>${this.isSearchMode ? 'Essayez de modifier votre recherche' : 'Commencez par ajouter votre premier article'}</p>
                </div>
            `;
            return;
        }

        container.innerHTML = this.filteredItems.map((item, index) => {
            const searchResult = searchResults ? searchResults[index] : null;
            return this.createItemCard(item, searchResult);
        }).join('');

        // Bind edit and delete buttons
        container.querySelectorAll('.edit-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const itemId = parseInt(e.target.dataset.id);
                this.openEditModal(itemId);
            });
        });

        container.querySelectorAll('.delete-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const itemId = parseInt(e.target.dataset.id);
                this.deleteItem(itemId);
            });
        });
    }

    createItemCard(item, searchResult = null) {
        const expiryDate = item.expiry_date ? new Date(item.expiry_date) : null;
        const today = new Date();
        let expiryClass = '';
        let expiryText = '';

        if (expiryDate) {
            const daysUntilExpiry = Math.ceil((expiryDate - today) / (1000 * 60 * 60 * 24));
            if (daysUntilExpiry < 0) {
                expiryClass = 'expired';
                expiryText = 'Expiré';
            } else if (daysUntilExpiry <= 3) {
                expiryClass = 'warning';
                expiryText = `Expire dans ${daysUntilExpiry} jour${daysUntilExpiry > 1 ? 's' : ''}`;
            } else if (daysUntilExpiry <= 7) {
                expiryClass = 'warning';
                expiryText = `Expire le ${expiryDate.toLocaleDateString('fr-FR')}`;
            } else {
                expiryClass = 'safe';
                expiryText = `Expire le ${expiryDate.toLocaleDateString('fr-FR')}`;
            }
        }

        const addedDate = new Date(item.added_date).toLocaleDateString('fr-FR');

        return `
            <div class="item-card ${searchResult ? 'search-result-item' : ''}">
                ${searchResult ? `<div class="search-score">Score: ${(searchResult.score * 100).toFixed(0)}%</div>` : ''}
                <div class="item-header">
                    <h3 class="item-name">${this.escapeHtml(item.name)}</h3>
                    <div class="item-actions">
                        <button class="btn btn-small btn-secondary edit-btn" data-id="${item.id}">Modifier</button>
                        <button class="btn btn-small btn-danger delete-btn" data-id="${item.id}">Supprimer</button>
                    </div>
                </div>
                <div class="item-info">
                    <span class="item-location">${this.escapeHtml(item.location)}</span>
                    ${item.category ? `<span class="item-category">${this.escapeHtml(item.category)}</span>` : ''}
                </div>
                ${item.description ? `<p class="item-description">${this.escapeHtml(item.description)}</p>` : ''}
                <div class="item-meta">
                    <span class="item-quantity">Quantité: ${item.quantity}</span>
                    ${expiryText ? `<span class="item-expiry ${expiryClass}">${expiryText}</span>` : ''}
                </div>
                <div class="item-meta">
                    <span>Ajouté le ${addedDate}</span>
                </div>
            </div>
        `;
    }

    openAddModal() {
        document.getElementById('modalTitle').textContent = 'Ajouter un article';
        document.getElementById('itemForm').reset();
        document.getElementById('itemId').value = '';
        document.getElementById('itemModal').style.display = 'block';
    }

    openEditModal(itemId) {
        const item = this.items.find(i => i.id === itemId);
        if (!item) return;

        document.getElementById('modalTitle').textContent = 'Modifier l\'article';
        document.getElementById('itemId').value = item.id;
        document.getElementById('itemName').value = item.name;
        document.getElementById('itemDescription').value = item.description || '';
        document.getElementById('itemLocation').value = item.location;
        document.getElementById('itemCategory').value = item.category || '';
        document.getElementById('itemQuantity').value = item.quantity;
        document.getElementById('itemExpiry').value = item.expiry_date || '';
        document.getElementById('itemModal').style.display = 'block';
    }

    closeModal() {
        document.getElementById('itemModal').style.display = 'none';
    }

    async handleFormSubmit(e) {
        e.preventDefault();
        
        const itemId = document.getElementById('itemId').value;
        const itemData = {
            name: document.getElementById('itemName').value,
            description: document.getElementById('itemDescription').value,
            location: document.getElementById('itemLocation').value,
            category: document.getElementById('itemCategory').value,
            quantity: parseInt(document.getElementById('itemQuantity').value),
            expiry_date: document.getElementById('itemExpiry').value || null
        };

        try {
            let response;
            if (itemId) {
                // Update existing item
                response = await fetch(`/api/items/${itemId}`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(itemData)
                });
            } else {
                // Create new item
                response = await fetch('/api/items', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(itemData)
                });
            }

            if (!response.ok) throw new Error('Erreur lors de la sauvegarde');

            this.closeModal();
            this.loadItems();
            this.showSuccess(itemId ? 'Article modifié avec succès' : 'Article ajouté avec succès');
        } catch (error) {
            console.error('Erreur:', error);
            this.showError('Erreur lors de la sauvegarde de l\'article');
        }
    }

    async deleteItem(itemId) {
        if (!confirm('Êtes-vous sûr de vouloir supprimer cet article ?')) return;

        try {
            const response = await fetch(`/api/items/${itemId}`, {
                method: 'DELETE'
            });

            if (!response.ok) throw new Error('Erreur lors de la suppression');

            this.loadItems();
            this.showSuccess('Article supprimé avec succès');
        } catch (error) {
            console.error('Erreur:', error);
            this.showError('Erreur lors de la suppression de l\'article');
        }
    }

    showLoading() {
        document.getElementById('itemsContainer').innerHTML = `
            <div class="loading">
                Chargement en cours...
            </div>
        `;
    }

    showError(message) {
        this.showNotification(message, 'error');
    }

    showSuccess(message) {
        this.showNotification(message, 'success');
    }

    showNotification(message, type) {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 15px 20px;
            border-radius: 8px;
            color: white;
            font-weight: 600;
            z-index: 2000;
            max-width: 300px;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
            transform: translateX(100%);
            transition: transform 0.3s ease;
        `;
        
        if (type === 'error') {
            notification.style.background = 'linear-gradient(135deg, #f56565 0%, #e53e3e 100%)';
        } else {
            notification.style.background = 'linear-gradient(135deg, #48bb78 0%, #38a169 100%)';
        }
        
        notification.textContent = message;
        document.body.appendChild(notification);

        // Animate in
        setTimeout(() => {
            notification.style.transform = 'translateX(0)';
        }, 100);

        // Remove after 3 seconds
        setTimeout(() => {
            notification.style.transform = 'translateX(100%)';
            setTimeout(() => {
                document.body.removeChild(notification);
            }, 300);
        }, 3000);
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// Initialize the application when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new JujuDB();
});
