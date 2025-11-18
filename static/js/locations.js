// Locations Management Module
class LocationsManager {
    constructor(app) {
        this.app = app;
        this.bindEvents();
    }

    bindEvents() {
        // Location management modal (only on dashboard)
        const manageLocationsBtn = document.getElementById('manageLocationsBtn');
        if (manageLocationsBtn) {
            manageLocationsBtn.addEventListener('click', () => {
                this.openLocationModal();
            });
        }
        
        // Close location modal (only on pages with location modal)
        const closeLocationBtn = document.querySelector('.close-location');
        if (closeLocationBtn) {
            closeLocationBtn.addEventListener('click', () => {
                this.closeLocationModal();
            });
        }
        
        // Location management tabs (only on locations page)
        const tabBtns = document.querySelectorAll('.tab-btn');
        if (tabBtns.length > 0) {
            tabBtns.forEach(btn => {
                btn.addEventListener('click', (e) => {
                    const tab = e.currentTarget && e.currentTarget.dataset ? e.currentTarget.dataset.tab : null;
                    if (tab) this.switchTab(tab);
                });
            });
        }
        
        // Add location button (only on locations page)
        const addLocationBtn = document.getElementById('addLocationBtn');
        if (addLocationBtn) {
            addLocationBtn.addEventListener('click', () => {
                this.addLocation();
            });
        }
        
        // Filter change events (only on dashboard)
        const locationFilter = document.getElementById('locationFilter');
        if (locationFilter) {
            locationFilter.addEventListener('change', () => {
                this.updateSubLocationFilter();
                if (this.app.filters) {
                    this.app.filters.applyFilters();
                }
            });
        }
        
        // Item location change event for modal (only on dashboard)
        const itemLocation = document.getElementById('itemLocation');
        if (itemLocation) {
            itemLocation.addEventListener('change', () => {
                this.updateItemSubLocationSelect();
            });
        }
    }

    async loadLocations() {
        try {
            const response = await fetch('/api/locations');
            if (response.ok) {
                this.app.locations = await response.json();
                this.populateLocationSelects();
            }
        } catch (error) {
            console.error('Error loading locations:', error);
        }
    }
    
    async loadSubLocations(locationId = null) {
        try {
            const url = locationId ? `/api/sub-locations?location_id=${locationId}` : '/api/sub-locations';
            const response = await fetch(url);
            if (response.ok) {
                this.app.subLocations = await response.json();
                this.populateSubLocationSelects();
            }
        } catch (error) {
            console.error('Error loading sub-locations:', error);
        }
    }
    
    populateLocationSelects() {
        const filterSelect = document.getElementById('locationFilter');
        const itemSelect = document.getElementById('itemLocation');
        const parentSelect = document.getElementById('parentLocationSelect');
        
        // Clear existing options (keep first option)
        [filterSelect, itemSelect, parentSelect].forEach(select => {
            if (select) {
                while (select.children.length > 1) {
                    select.removeChild(select.lastChild);
                }
            }
        });
        
        // Populate with locations
        this.app.locations.forEach(location => {
            if (filterSelect) {
                const option = document.createElement('option');
                option.value = location.id;
                option.textContent = location.name;
                filterSelect.appendChild(option);
            }
            
            if (itemSelect) {
                const option = document.createElement('option');
                option.value = location.id;
                option.textContent = location.name;
                itemSelect.appendChild(option);
            }
            
            if (parentSelect) {
                const option = document.createElement('option');
                option.value = location.id;
                option.textContent = location.name;
                parentSelect.appendChild(option);
            }
        });
    }
    
    populateSubLocationSelects() {
        this.updateSubLocationFilter();
        this.updateItemSubLocationSelect();
    }
    
    async updateSubLocationFilter() {
        const filterSelect = document.getElementById('subLocationFilter');
        const locationFilter = document.getElementById('locationFilter');
        
        if (!filterSelect || !locationFilter) return;
        
        const selectedLocationId = locationFilter.value;
        
        // Clear existing options (keep first option)
        while (filterSelect.children.length > 1) {
            filterSelect.removeChild(filterSelect.lastChild);
        }
        
        // Reset filter value
        filterSelect.value = '';
        
        if (!selectedLocationId) {
            // Hide sub-location filter when no location is selected
            filterSelect.style.display = 'none';
            filterSelect.classList.remove('show');
            return;
        }
        
        try {
            // Call API to get sub-locations for the selected location
            const response = await fetch(`/api/sub-locations?location_id=${selectedLocationId}`);
            if (response.ok) {
                const subLocations = await response.json();
                
                // Ensure subLocations is an array
                if (Array.isArray(subLocations) && subLocations.length > 0) {
                    // Show the filter select only when there are sub-locations
                    filterSelect.style.display = 'block';
                    filterSelect.classList.add('show');
                    
                    // Populate with sub-locations from API
                    subLocations.forEach(subLocation => {
                        const option = document.createElement('option');
                        option.value = subLocation.id;
                        option.textContent = subLocation.name;
                        filterSelect.appendChild(option);
                    });
                } else {
                    // If no sub-locations available, hide the filter
                    filterSelect.style.display = 'none';
                    filterSelect.classList.remove('show');
                }
            } else {
                // Hide filter on API error
                filterSelect.style.display = 'none';
                filterSelect.classList.remove('show');
            }
        } catch (error) {
            console.error('Error loading sub-locations for filter:', error);
            // Hide filter on error
            filterSelect.style.display = 'none';
            filterSelect.classList.remove('show');
        }
    }
    
    updateItemSubLocationSelect() {
        const itemSelect = document.getElementById('itemSubLocation');
        const locationSelect = document.getElementById('itemLocation');
        
        if (!itemSelect || !locationSelect) return;
        
        const selectedLocationId = locationSelect.value;
        
        // Clear existing options (keep first option)
        while (itemSelect.children.length > 1) {
            itemSelect.removeChild(itemSelect.lastChild);
        }
        
        // Reset selection
        itemSelect.value = '';
        
        if (!selectedLocationId) {
            // Hide sub-location select when no location is selected
            itemSelect.style.display = 'none';
            return;
        }
        // Ensure the select is visible when a location is selected
        itemSelect.style.display = '';
        
        // Show sub-location select
        itemSelect.style.display = 'block';
        
        // Filter sub-locations by selected location
        const filteredSubLocations = this.app.subLocations.filter(sub => sub.location_id == selectedLocationId);
        
        // Populate with filtered sub-locations
        filteredSubLocations.forEach(subLocation => {
            const option = document.createElement('option');
            option.value = subLocation.id;
            option.textContent = subLocation.name;
            itemSelect.appendChild(option);
        });
        
        // If no sub-locations available, hide the select; otherwise ensure it's visible
        if (filteredSubLocations.length === 0) {
            itemSelect.style.display = 'none';
        } else {
            itemSelect.style.display = '';
        }
    }

    openLocationModal() {
        document.getElementById('locationModal').style.display = 'block';
        this.loadLocationManagementData();
    }

    closeLocationModal() {
        document.getElementById('locationModal').style.display = 'none';
    }

    switchTab(tabName) {
        // Hide all tab contents
        document.querySelectorAll('.tab-content').forEach(tab => {
            tab.classList.remove('active');
        });
        
        // Remove active class from all tab buttons
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        
        // Show selected tab content
        const contentEl = document.getElementById(`${tabName}-tab`);
        if (contentEl) {
            contentEl.classList.add('active');
        }
        
        // Add active class to selected tab button
        const btnEl = document.querySelector(`[data-tab="${tabName}"]`);
        if (btnEl) {
            btnEl.classList.add('active');
        }
    }

    async loadLocationManagementData() {
        await this.loadLocations();
        await this.loadSubLocations();
        await this.app.categoriesManager.loadCategories();
        this.renderLocationLists();
    }

    renderLocationLists() {
        // Render locations list
        const locationsList = document.getElementById('locationsList');
        if (locationsList) {
            locationsList.innerHTML = this.app.locations.map(location => `
                <div class="list-item">
                    <span class="list-item-name">${location.name}</span>
                    <button class="btn btn-sm btn-danger" onclick="window.locationsPage.locationsManager.deleteLocation(${location.id})">Supprimer</button>
                </div>
            `).join('');
        }
        
        // Also render sub-locations and categories lists
        this.app.subLocationsManager.renderSubLocationsList();
        this.app.categoriesManager.renderCategoriesList();
    }

    async addLocation() {
        const name = document.getElementById('newLocationName').value.trim();
        if (!name) {
            this.app.showError('Le nom de l\'emplacement est obligatoire');
            return;
        }
        
        try {
            const response = await fetch('/api/locations', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ name })
            });
            
            if (response.ok) {
                document.getElementById('newLocationName').value = '';
                this.loadLocationManagementData();
                this.app.showSuccess('Emplacement ajouté avec succès');
            } else {
                throw new Error('Erreur lors de l\'ajout');
            }
        } catch (error) {
            console.error('Error adding location:', error);
            this.app.showError('Erreur lors de l\'ajout de l\'emplacement');
        }
    }

    async deleteLocation(id) {
        if (!confirm('Êtes-vous sûr de vouloir supprimer cet emplacement ?')) {
            return;
        }

        const doDelete = async (force = false) => {
            const url = force ? `/api/locations/${id}?force=true` : `/api/locations/${id}`;
            const response = await fetch(url, { method: 'DELETE' });
            return response;
        };

        try {
            let response = await doDelete(false);
            if (response.ok) {
                await this.loadLocationManagementData();
                this.app.showSuccess('Emplacement supprimé avec succès');
                return;
            }

            if (response.status === 409) {
                const data = await response.json().catch(() => ({}));
                // Show modal with related items and sub-locations, then confirm forced deletion
                window.locationsPage.showConfirmDelete({
                    type: 'location',
                    id,
                    related_items: data.related_items || [],
                    related_sublocations: data.related_sublocations || []
                }, async () => {
                    const forceResp = await doDelete(true);
                    if (forceResp.ok) {
                        await this.loadLocationManagementData();
                        this.app.showSuccess('Emplacement et éléments liés supprimés');
                    } else {
                        this.app.showError('Échec de la suppression forcée de l\'emplacement');
                    }
                });
                return;
            }

            // Other errors
            throw new Error('Erreur lors de la suppression');
        } catch (error) {
            console.error('Error deleting location:', error);
            this.app.showError('Erreur lors de la suppression de l\'emplacement');
        }
    }
}
