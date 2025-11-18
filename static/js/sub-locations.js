// Sub-Locations Management Module
class SubLocationsManager {
    constructor(app) {
        this.app = app;
        this.bindEvents();
    }

    bindEvents() {
        // Add sub-location button (only on locations page)
        const addSubLocationBtn = document.getElementById('addSubLocationBtn');
        if (addSubLocationBtn) {
            addSubLocationBtn.addEventListener('click', () => {
                this.addSubLocation();
            });
        }
        
        // Sub-location filter change (only on dashboard)
        const subLocationFilter = document.getElementById('subLocationFilter');
        if (subLocationFilter) {
            subLocationFilter.addEventListener('change', () => {
                if (this.app.filters) {
                    this.app.filters.applyFilters();
                }
            });
        }
        
        // Parent location filter for management interface (only on locations page)
        const parentLocationSelect = document.getElementById('parentLocationSelect');
        if (parentLocationSelect) {
            parentLocationSelect.addEventListener('change', () => {
                this.filterSubLocationsList();
            });
        }
    }

    async loadSubLocations(locationId = null) {
        try {
            const url = locationId ? `/api/sub-locations?location_id=${locationId}` : '/api/sub-locations';
            const response = await fetch(url);
            if (response.ok) {
                const data = await response.json();
                this.app.subLocations = Array.isArray(data) ? data : [];
                // Update selects via LocationsManager if available
                if (this.app.locationsManager && typeof this.app.locationsManager.populateSubLocationSelects === 'function') {
                    this.app.locationsManager.populateSubLocationSelects();
                }
                // Re-render sub-locations list if present
                if (typeof this.renderSubLocationsList === 'function') {
                    this.renderSubLocationsList();
                }
            } else {
                this.app.subLocations = [];
            }
        } catch (error) {
            console.error('Error loading sub-locations:', error);
            this.app.subLocations = [];
        }
    }

    renderSubLocationsList() {
        // Render sub-locations list
        const subLocationsList = document.getElementById('subLocationsList');
        if (!subLocationsList) return;
        
        // Get selected parent location for filtering
        const selectedParentId = document.getElementById('filterLocationSelect')?.value;
        
        // Filter sub-locations based on selected parent location
        let filteredSubLocations = this.app.subLocations;
        if (selectedParentId) {
            filteredSubLocations = this.app.subLocations.filter(sub => sub.location_id == selectedParentId);
        }
        
        if (filteredSubLocations.length === 0) {
            subLocationsList.innerHTML = '<div class="no-items">Aucun sous-emplacement trouvé</div>';
            return;
        }
        
        subLocationsList.innerHTML = filteredSubLocations.map(subLocation => {
            const location = this.app.locations.find(l => l.id === subLocation.location_id);
            return `
                <div class="list-item">
                    <span class="list-item-name">${subLocation.name}<span class="list-item-parent"> (${location ? location.name : 'N/A'})</span></span>
                    <button class="btn btn-sm btn-danger" onclick="window.locationsPage.subLocationsManager.deleteSubLocation(${subLocation.id})">Supprimer</button>
                </div>
            `;
        }).join('');
    }
    
    filterSubLocationsList() {
        // Re-render the list with current filter
        this.renderSubLocationsList();
    }

    async addSubLocation() {
        const name = document.getElementById('newSubLocationName').value.trim();
        const locationId = document.getElementById('parentLocationSelect').value;
        
        if (!name || !locationId) {
            this.app.showError('Le nom et l\'emplacement parent sont obligatoires');
            return;
        }
        
        try {
            const response = await fetch('/api/sub-locations', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ name, location_id: parseInt(locationId) })
            });
            
            if (response.ok) {
                document.getElementById('newSubLocationName').value = '';
                document.getElementById('parentLocationSelect').value = '';
                await this.app.locationsManager.loadLocationManagementData();
                this.app.showSuccess('Sous-emplacement ajouté avec succès');
            } else {
                throw new Error('Erreur lors de l\'ajout');
            }
        } catch (error) {
            console.error('Error adding sub-location:', error);
            this.app.showError('Erreur lors de l\'ajout du sous-emplacement');
        }
    }

    async deleteSubLocation(id) {
        if (!confirm('Êtes-vous sûr de vouloir supprimer ce sous-emplacement ?')) {
            return;
        }

        const doDelete = async (force = false) => {
            const url = force ? `/api/sub-locations/${id}?force=true` : `/api/sub-locations/${id}`;
            const response = await fetch(url, { method: 'DELETE' });
            return response;
        };

        try {
            let response = await doDelete(false);
            if (response.ok) {
                await this.app.locationsManager.loadLocationManagementData();
                this.app.showSuccess('Sous-emplacement supprimé avec succès');
                return;
            }

            if (response.status === 409) {
                const data = await response.json().catch(() => ({}));
                window.locationsPage.showConfirmDelete({
                    type: 'sub_location',
                    id,
                    related_items: data.related_items || []
                }, async () => {
                    const forceResp = await doDelete(true);
                    if (forceResp.ok) {
                        await this.app.locationsManager.loadLocationManagementData();
                        this.app.showSuccess('Sous-emplacement et éléments liés supprimés');
                    } else {
                        this.app.showError('Échec de la suppression forcée du sous-emplacement');
                    }
                });
                return;
            }

            throw new Error('Erreur lors de la suppression');
        } catch (error) {
            console.error('Error deleting sub-location:', error);
            this.app.showError('Erreur lors de la suppression du sous-emplacement');
        }
    }
}
