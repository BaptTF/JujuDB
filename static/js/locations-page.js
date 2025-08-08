// Locations Page Coordinator
class LocationsPage {
    constructor() {
        this.locations = [];
        this.subLocations = [];
        this.categories = [];
        
        // Initialize feature modules
        this.locationsManager = new LocationsManager(this);
        this.subLocationsManager = new SubLocationsManager(this);
        this.categoriesManager = new CategoriesManager(this);
        
        // Deletion confirmation modal state/elements
        this.pendingDelete = null;
        this.confirmModal = {
            el: document.getElementById('confirmDeleteModal'),
            title: document.getElementById('confirmDeleteTitle'),
            message: document.getElementById('confirmDeleteMessage'),
            list: document.getElementById('confirmDeleteList'),
            btnClose: document.getElementById('closeConfirmDelete'),
            btnCancel: document.getElementById('cancelDeleteBtn'),
            btnConfirm: document.getElementById('confirmDeleteBtn'),
        };

        this.init();
    }

    init() {
        this.bindEvents();
        this.loadInitialData();
    }

    bindEvents() {
        // Tab switching
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const tab = e.currentTarget && e.currentTarget.dataset ? e.currentTarget.dataset.tab : null;
                if (tab) this.switchTab(tab);
            });
        });

        // Filter for sub-locations
        const filterLocationSelect = document.getElementById('filterLocationSelect');
        if (filterLocationSelect) {
            filterLocationSelect.addEventListener('change', () => {
                this.subLocationsManager.filterSubLocationsList();
            });
        }

        // Modal buttons
        if (this.confirmModal.btnClose) this.confirmModal.btnClose.addEventListener('click', () => this.hideConfirmDelete());
        if (this.confirmModal.btnCancel) this.confirmModal.btnCancel.addEventListener('click', () => this.hideConfirmDelete());
        if (this.confirmModal.el) {
            // Close when clicking outside content
            this.confirmModal.el.addEventListener('click', (e) => {
                if (e.target === this.confirmModal.el) this.hideConfirmDelete();
            });
        }
    }

    switchTab(tabName) {
        // Update tab buttons
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');

        // Update tab content
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.remove('active');
        });
        document.getElementById(`${tabName}-tab`).classList.add('active');
    }

    async loadInitialData() {
        try {
            await this.locationsManager.loadLocations();
            await this.categoriesManager.loadCategories();
            await this.subLocationsManager.loadSubLocations();
            
            // Populate filter dropdowns
            this.populateFilterDropdowns();

            // Render initial location, sub-location, and category lists
            this.locationsManager.renderLocationLists();
        } catch (error) {
            console.error('Error loading initial data:', error);
            this.showError('Erreur lors du chargement des données');
        }
    }

    populateFilterDropdowns() {
        // Populate parent location select for sub-locations
        const parentSelect = document.getElementById('parentLocationSelect');
        const filterSelect = document.getElementById('filterLocationSelect');
        
        // Clear existing options (keep first option)
        [parentSelect, filterSelect].forEach(select => {
            if (select) {
                while (select.children.length > 1) {
                    select.removeChild(select.lastChild);
                }
            }
        });

        // Add location options
        this.locations.forEach(location => {
            if (parentSelect) {
                const option = document.createElement('option');
                option.value = location.id;
                option.textContent = location.name;
                parentSelect.appendChild(option);
            }
            
            if (filterSelect) {
                const option = document.createElement('option');
                option.value = location.id;
                option.textContent = location.name;
                filterSelect.appendChild(option);
            }
        });
    }

    showLoading(containerId) {
        const container = document.getElementById(containerId);
        if (container) {
            container.innerHTML = '<div class="loading">Chargement...</div>';
        }
    }

    showError(message) {
        alert(message);
    }

    showSuccess(message) {
        alert(message);
    }

    // Deletion confirmation modal API
    showConfirmDelete(data, onConfirm) {
        if (!this.confirmModal.el) {
            // Fallback if modal not available
            if (confirm('Confirmer la suppression et des éléments liés ?')) {
                onConfirm && onConfirm();
            }
            return;
        }

        this.pendingDelete = { type: data.type, id: data.id, onConfirm };

        // Title/message per type
        const typeLabels = {
            location: 'cet emplacement',
            sub_location: 'ce sous-emplacement',
            category: 'cette catégorie',
        };
        const typeTitle = {
            location: 'Supprimer l\'emplacement',
            sub_location: 'Supprimer le sous-emplacement',
            category: 'Supprimer la catégorie',
        };

        if (this.confirmModal.title) this.confirmModal.title.textContent = typeTitle[data.type] || 'Confirmer la suppression';
        if (this.confirmModal.message) this.confirmModal.message.textContent = `Des articles sont liés à ${typeLabels[data.type] || 'cet élément'}. La suppression entraînera la suppression des articles suivants :`;

        // Populate list: include sub-locations when deleting a location
        if (this.confirmModal.list) {
            const items = Array.isArray(data.related_items) ? data.related_items : [];
            const sublocs = Array.isArray(data.related_sublocations) ? data.related_sublocations : [];

            let html = '';
            if (data.type === 'location' && sublocs.length > 0) {
                html += `<h4>Sous-emplacements affectés (${sublocs.length})</h4>`;
                html += sublocs.map(sl => `<div class=\"modal-list-item\">• ${sl.name} (ID ${sl.id})</div>`).join('');
            }
            if (items.length > 0) {
                html += `<h4 style=\"margin-top:12px;\">Articles affectés (${items.length})</h4>`;
                html += items.map(it => `<div class=\"modal-list-item\">• ${it.name} (ID ${it.id})</div>`).join('');
            }
            if (!html) {
                html = '<div class=\"modal-list-item\">Aucun élément lié</div>';
            }
            this.confirmModal.list.innerHTML = html;
        }

        // Ensure confirm button triggers provided action once
        if (this.confirmModal.btnConfirm) {
            const newBtn = this.confirmModal.btnConfirm.cloneNode(true);
            this.confirmModal.btnConfirm.parentNode.replaceChild(newBtn, this.confirmModal.btnConfirm);
            this.confirmModal.btnConfirm = newBtn;
            this.confirmModal.btnConfirm.addEventListener('click', async () => {
                const action = this.pendingDelete && this.pendingDelete.onConfirm;
                this.hideConfirmDelete();
                if (action) await action();
                this.pendingDelete = null;
            });
        }

        this.confirmModal.el.style.display = 'block';
    }

    hideConfirmDelete() {
        if (this.confirmModal && this.confirmModal.el) {
            this.confirmModal.el.style.display = 'none';
        }
        this.pendingDelete = null;
    }
}

// Initialize the locations page when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.locationsPage = new LocationsPage();
});
