// Filters Management Module
class FiltersManager {
    constructor(app) {
        this.app = app;
        this.bindEvents();
    }

    bindEvents() {
        // Clear filters button (only on dashboard)
        const clearFiltersBtn = document.getElementById('clearFilters');
        if (clearFiltersBtn) {
            clearFiltersBtn.addEventListener('click', () => {
                this.clearFilters();
            });
        }
    }

    applyFilters() {
        const locationFilterEl = document.getElementById('locationFilter');
        const subLocationFilterEl = document.getElementById('subLocationFilter');
        const categoryFilterEl = document.getElementById('categoryFilter');
        
        // Only apply filters if elements exist (dashboard page)
        if (!locationFilterEl || !subLocationFilterEl || !categoryFilterEl) {
            return;
        }
        
        const locationFilter = locationFilterEl.value;
        const subLocationFilter = subLocationFilterEl.value;
        const categoryFilter = categoryFilterEl.value;
        
        this.app.filteredItems = this.app.items.filter(item => {
            const matchesLocation = !locationFilter || item.location_id == locationFilter;
            const matchesSubLocation = !subLocationFilter || item.sub_location_id == subLocationFilter;
            const matchesCategory = !categoryFilter || item.category_id == categoryFilter;
            return matchesLocation && matchesSubLocation && matchesCategory;
        });
        
        this.app.articles.renderItems();
    }

    clearFilters() {
        const locationFilterEl = document.getElementById('locationFilter');
        const subLocationFilterEl = document.getElementById('subLocationFilter');
        const categoryFilterEl = document.getElementById('categoryFilter');
        
        // Only clear filters if elements exist (dashboard page)
        if (locationFilterEl) locationFilterEl.value = '';
        if (subLocationFilterEl) subLocationFilterEl.value = '';
        if (categoryFilterEl) categoryFilterEl.value = '';
        
        this.app.filteredItems = [...this.app.items];
        if (this.app.articles) {
            this.app.articles.renderItems();
        }
    }
}
