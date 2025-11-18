// Dashboard App Coordinator
class AppCoordinator {
  constructor() {
    this.items = [];
    this.filteredItems = [];
    this.locations = [];
    this.subLocations = [];
    this.categories = [];
    this.isSearchMode = false;

    // Bind UI helpers
    this.showLoading = this.showLoading.bind(this);
    this.showError = this.showError.bind(this);
    this.showSuccess = this.showSuccess.bind(this);

    // Initialize feature modules
    this.locationsManager = new LocationsManager(this);
    this.subLocationsManager = new SubLocationsManager(this);
    this.categoriesManager = new CategoriesManager(this);
    this.filters = new FiltersManager(this);
    this.articles = new ArticlesManager(this);
    this.search = new SearchManager(this);

    // Load initial data
    this.init();
  }

  async init() {
    try {
      this.showLoading();
      await this.locationsManager.loadLocations();
      await this.subLocationsManager.loadSubLocations();
      await this.categoriesManager.loadCategories();
      await this.articles.loadItems();
    } catch (e) {
      console.error('Error during dashboard init:', e);
      this.showError("Erreur lors du chargement initial");
    }
  }

  showLoading() {
    const container = document.getElementById('itemsContainer');
    if (container) container.innerHTML = '<div class="loading">Chargement...</div>';
  }

  showError(message) {
    alert(message);
  }

  showSuccess(message) {
    alert(message);
  }
}

// Initialize on DOM ready
window.addEventListener('DOMContentLoaded', () => {
  window.app = new AppCoordinator();
});
