// Search Management Module
class SearchManager {
    constructor(app) {
        this.app = app;
        this.searchTimeout = null;
        this.bindEvents();
    }

    bindEvents() {
        // Search functionality with debouncing
        document.getElementById('searchInput').addEventListener('input', (e) => {
            this.handleSearch(e.target.value);
        });
        
        // Search button
        document.getElementById('searchBtn').addEventListener('click', () => {
            this.performSearch();
        });
    }

    handleSearch(query) {
        // Clear previous timeout
        if (this.searchTimeout) {
            clearTimeout(this.searchTimeout);
        }
        
        if (!query.trim()) {
            this.app.filteredItems = [...this.app.items];
            this.app.isSearchMode = false;
            this.app.articles.renderItems();
            return;
        }
        
        // Debounce search to avoid too many API calls
        this.searchTimeout = setTimeout(() => {
            this.performSearch();
        }, 300);
    }

    async performSearch() {
        const query = document.getElementById('searchInput').value.trim();
        if (!query) {
            this.app.articles.loadItems();
            return;
        }

        try {
            this.app.showLoading();
            const locationFilter = document.getElementById('locationFilter').value;
            const subLocationFilter = document.getElementById('subLocationFilter').value;
            const categoryFilter = document.getElementById('categoryFilter').value;
            
            let url = `/api/search?q=${encodeURIComponent(query)}`;
            if (locationFilter) url += `&location_id=${encodeURIComponent(locationFilter)}`;
            if (subLocationFilter) url += `&sub_location_id=${encodeURIComponent(subLocationFilter)}`;
            if (categoryFilter) url += `&category_id=${encodeURIComponent(categoryFilter)}`;

            const response = await fetch(url);
            if (!response.ok) throw new Error('Erreur lors de la recherche');
            
            let searchResults = await response.json();
            if (!Array.isArray(searchResults)) {
                searchResults = [];
            }
            // Support both schemas: [{item, score}] or [item]
            let itemsList;
            let resultsForRender = null;
            if (searchResults.length > 0 && typeof searchResults[0] === 'object' && 'item' in searchResults[0]) {
                // Schema with score
                itemsList = searchResults.map(r => r.item);
                resultsForRender = searchResults; // preserve scores for UI
            } else {
                // Schema is raw items array
                itemsList = searchResults;
            }
            this.app.filteredItems = itemsList;
            this.app.isSearchMode = true;
            this.app.articles.renderItems(resultsForRender);
        } catch (error) {
            console.error('Erreur de recherche:', error);
            this.app.showError('Erreur lors de la recherche: ' + error.message);
            // Reset to show all items on error
            this.app.filteredItems = [...this.app.items];
            this.app.isSearchMode = false;
            this.app.articles.renderItems();
        }
    }
}
