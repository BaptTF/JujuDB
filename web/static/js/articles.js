// Articles Management Module
class ArticlesManager {
    constructor(app) {
        this.app = app;
        this.bindEvents();
    }

    bindEvents() {
        // Add item button
        document.getElementById('addItemBtn').addEventListener('click', () => {
            this.openModal();
        });

        // Save item on form submit
        document.getElementById('itemForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.saveItem();
        });

        // Modal close (X icon)
        document.querySelector('.close').addEventListener('click', () => {
            this.closeModal();
        });
        // Cancel button
        document.getElementById('cancelBtn').addEventListener('click', () => {
            this.closeModal();
        });

        // Location dependency for sub-locations
        document.getElementById('itemLocation').addEventListener('change', () => {
            this.loadItemSubLocations();
        });
    }

    async loadItems() {
        try {
            const response = await fetch('/api/items');
            if (response.ok) {
                this.app.items = await response.json();
                this.app.filteredItems = [...this.app.items];
                this.renderItems();
            }
        } catch (error) {
            console.error('Error loading items:', error);
            this.app.showError('Erreur lors du chargement des articles');
        }
    }

    renderItems(searchResults = null) {
        const container = document.getElementById('itemsContainer');
        container.innerHTML = '';

        if (this.app.filteredItems.length === 0) {
            container.innerHTML = '<div class="no-items">Aucun article trouvé</div>';
            return;
        }

        this.app.filteredItems.forEach((item, index) => {
            const itemElement = document.createElement('div');
            itemElement.className = 'item-card';

            const searchResult = searchResults ? searchResults[index] : null;
            const scoreDisplay = searchResult && searchResult.score !== undefined ?
                `<div class="search-score">Score: ${Math.round(searchResult.score * 100)}%</div>` : '';

            const expiryDate = item.expiry_date ? new Date(item.expiry_date).toLocaleDateString('fr-FR') : 'Non définie';
            const addedDate = new Date(item.added_at).toLocaleDateString('fr-FR');

            // Get location and sub-location names
            const location = this.app.locations.find(l => l.id == item.location_id);
            const subLocation = this.app.subLocations.find(sl => sl.id == item.sub_location_id);
            const category = this.app.categories.find(c => c.id == item.category_id);

            const locationDisplay = location ? location.name : 'Non défini';
            const subLocationDisplay = subLocation ?
                `<span class="item-sub-location">${subLocation.name}</span>` : '';
            const categoryDisplay = category ? category.name : 'Non définie';

            itemElement.innerHTML = `
                <div class="item-header">
                    <h3 class="item-name">${item.name}</h3>
                    <div class="item-actions">
                        <button class="btn-icon btn-edit" onclick="app.articles.editItem(${item.id})" title="Modifier"><i class="fa-solid fa-pen-to-square"></i></button>
                        <button class="btn-icon btn-delete" onclick="app.articles.deleteItem(${item.id})" title="Supprimer"><i class="fa-solid fa-trash"></i></button>
                    </div>
                </div>
                
                ${item.description ? `<div class="item-description">${item.description}</div>` : ''}
                
                <div class="item-meta">
                    <div class="item-location">
                        <span class="location-icon"><i class="fa-solid fa-location-dot"></i></span>
                        ${locationDisplay}${subLocationDisplay ? ` → ${subLocation.name}` : ''}
                    </div>
                    <div class="item-category">
                        <span class="category-icon"><i class="fa-solid fa-tags"></i></span>
                        ${categoryDisplay}
                    </div>
                </div>
                
                <div class="item-details">
                    <div class="item-quantity">
                        <span class="quantity-icon"><i class="fa-solid fa-box"></i></span>
                        Qté: ${item.quantity || 1}
                    </div>
                    <div class="item-expiry ${this.getExpiryClass(item.expiry_date)}">
                        <span class="expiry-icon"><i class="fa-solid fa-clock"></i></span>
                        ${expiryDate}
                    </div>
                </div>
                
                ${item.notes ? `<div class="item-notes"><span class="notes-icon"><i class="fa-solid fa-note-sticky"></i></span> ${item.notes}</div>` : ''}
                
                <div class="item-footer">
                    <small class="item-added">Ajouté le ${addedDate}</small>
                </div>
                
                ${scoreDisplay}
            `;

            container.appendChild(itemElement);
        });
    }

    openModal(item = null) {
        const modal = document.getElementById('itemModal');
        const title = document.querySelector('#itemModal h2');

        if (item) {
            title.textContent = 'Modifier l\'article';
            document.getElementById('itemId').value = item.id;
            document.getElementById('itemName').value = item.name;
            document.getElementById('itemLocation').value = item.location_id || '';
            document.getElementById('itemSubLocation').value = item.sub_location_id || '';
            document.getElementById('itemCategory').value = item.category_id || '';
            document.getElementById('itemExpiry').value = item.expiry_date || '';
            document.getElementById('itemNotes').value = item.notes || '';

            // Load sub-locations for the selected location
            if (item.location_id) {
                this.loadItemSubLocations();
            }
        } else {
            title.textContent = 'Ajouter un article';
            document.getElementById('itemForm').reset();
            document.getElementById('itemId').value = '';
        }

        modal.style.display = 'block';
    }

    closeModal() {
        document.getElementById('itemModal').style.display = 'none';
    }

    async saveItem() {
        const itemData = {
            name: document.getElementById('itemName').value,
            description: document.getElementById('itemDescription').value,
            location_id: parseInt(document.getElementById('itemLocation').value) || null,
            sub_location_id: parseInt(document.getElementById('itemSubLocation').value) || null,
            category_id: parseInt(document.getElementById('itemCategory').value) || null,
            quantity: parseInt(document.getElementById('itemQuantity').value) || 1,
            expiry_date: document.getElementById('itemExpiry').value || null,
            notes: document.getElementById('itemNotes').value
        };

        if (!itemData.name || !itemData.location_id) {
            this.app.showError('Le nom et l\'emplacement sont obligatoires');
            return;
        }

        const itemId = document.getElementById('itemId').value;
        const isEdit = itemId !== '';

        try {
            const url = isEdit ? `/api/items/${itemId}` : '/api/items';
            const method = isEdit ? 'PUT' : 'POST';

            const response = await fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(itemData)
            });

            if (response.ok) {
                this.closeModal();
                this.loadItems();
                this.app.showSuccess(isEdit ? 'Article modifié avec succès' : 'Article ajouté avec succès');
            } else {
                throw new Error('Erreur lors de la sauvegarde');
            }
        } catch (error) {
            console.error('Error saving item:', error);
            this.app.showError('Erreur lors de la sauvegarde de l\'article');
        }
    }

    async editItem(id) {
        const item = this.app.items.find(i => i.id === id);
        if (item) {
            this.openModal(item);
        }
    }

    async deleteItem(id) {
        if (!confirm('Êtes-vous sûr de vouloir supprimer cet article ?')) {
            return;
        }

        try {
            const response = await fetch(`/api/items/${id}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                this.loadItems();
                this.app.showSuccess('Article supprimé avec succès');
            } else {
                throw new Error('Erreur lors de la suppression');
            }
        } catch (error) {
            console.error('Error deleting item:', error);
            this.app.showError('Erreur lors de la suppression de l\'article');
        }
    }

    getExpiryClass(expiryDate) {
        if (!expiryDate) return 'no-expiry';

        const today = new Date();
        const expiry = new Date(expiryDate);
        const diffTime = expiry - today;
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

        if (diffDays < 0) return 'expired';
        if (diffDays <= 3) return 'expiring-soon';
        if (diffDays <= 7) return 'expiring-week';
        return 'fresh';
    }

    async loadItemSubLocations() {
        const locationId = document.getElementById('itemLocation').value;
        const subLocationSelect = document.getElementById('itemSubLocation');

        // Clear existing sub-location options (keep first option)
        while (subLocationSelect.children.length > 1) {
            subLocationSelect.removeChild(subLocationSelect.lastChild);
        }

        if (locationId) {
            // Show the sub-location select when a location is selected
            subLocationSelect.style.display = 'block';
            
            try {
                const response = await fetch(`/api/sub-locations?location_id=${locationId}`);
                if (response.ok) {
                    const subLocations = await response.json();
                    subLocations.forEach(subLocation => {
                        const option = document.createElement('option');
                        option.value = subLocation.id;
                        option.textContent = subLocation.name;
                        subLocationSelect.appendChild(option);
                    });
                }
            } catch (error) {
                console.error('Error loading sub-locations for item form:', error);
            }
        } 
    }
}
