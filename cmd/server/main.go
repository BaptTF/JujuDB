package main

import (
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"jujudb/internal/config"
	"jujudb/internal/database"
	"jujudb/internal/handlers"
	"jujudb/internal/repositories"
	"jujudb/internal/services"
)

var (
	store *sessions.CookieStore
	tmpl  *template.Template
)

func init() {
	// Configure logrus
	if os.Getenv("PRODUCTION") == "true" || os.Getenv("LOG_FORMAT") == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize session store
	store = sessions.NewCookieStore([]byte(cfg.Session.Key))

	// Initialize database
	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize database")
	}
	defer db.Close()

	// Load templates (legacy handlers may rely on this)
	tmpl = template.Must(template.ParseGlob("web/static/html/*.html"))

	// Initialize repositories
	repos := repositories.NewRepository(db.DB)

	// Initialize Meilisearch service
	meilisearchService, err := services.NewMeilisearchService(cfg.Meilisearch.Host, cfg.Meilisearch.MasterKey)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize Meilisearch service")
	}

	// Initialize sync service
	syncService := services.NewSyncService(db.DB, meilisearchService)

	// Initialize services
	serviceLayer := services.NewService(repos, syncService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(store)
	templatesHandler := handlers.NewTemplatesHandler()
	articlesHandler := handlers.NewArticlesHandler(serviceLayer.Items)
	locationsHandler := handlers.NewLocationsHandler(serviceLayer.Locations)
	subLocationsHandler := handlers.NewSubLocationsHandler(serviceLayer.SubLocations)
	categoriesHandler := handlers.NewCategoriesHandler(serviceLayer.Categories)
	searchHandler := handlers.NewSearchHandler(meilisearchService)
	syncHandler := handlers.NewSyncHandler(syncService)

	// Router
	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	// Public routes
	r.HandleFunc("/", authHandler.Root).Methods("GET")
	r.HandleFunc("/login", templatesHandler.LoginPage).Methods("GET")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/logout", authHandler.Logout).Methods("POST")

	// Protected pages
	r.Handle("/dashboard", authHandler.AuthMiddleware(http.HandlerFunc(templatesHandler.Dashboard))).Methods("GET")
	r.Handle("/locations", authHandler.AuthMiddleware(http.HandlerFunc(templatesHandler.LocationsPage))).Methods("GET")

	// API routes (protected)
	api := r.PathPrefix("/api").Subrouter()
	api.Use(func(next http.Handler) http.Handler { return authHandler.AuthMiddleware(next) })

	// Items
	api.HandleFunc("/items", articlesHandler.GetItems).Methods("GET")
	api.HandleFunc("/items", articlesHandler.CreateItem).Methods("POST")
	api.HandleFunc("/items/{id}", articlesHandler.UpdateItem).Methods("PUT")
	api.HandleFunc("/items/{id}", articlesHandler.DeleteItem).Methods("DELETE")

	// Search
	api.HandleFunc("/search", searchHandler.Search).Methods("GET")

	// Sync
	api.HandleFunc("/sync/all", syncHandler.SyncAll).Methods("POST")

	// Locations
	api.HandleFunc("/locations", locationsHandler.GetLocations).Methods("GET")
	api.HandleFunc("/locations", locationsHandler.CreateLocation).Methods("POST")
	api.HandleFunc("/locations/{id}", locationsHandler.UpdateLocation).Methods("PUT")
	api.HandleFunc("/locations/{id}", locationsHandler.DeleteLocation).Methods("DELETE")

	// Sub-locations
	api.HandleFunc("/sub-locations", subLocationsHandler.GetSubLocations).Methods("GET")
	api.HandleFunc("/sub-locations", subLocationsHandler.CreateSubLocation).Methods("POST")
	api.HandleFunc("/sub-locations/{id}", subLocationsHandler.UpdateSubLocation).Methods("PUT")
	api.HandleFunc("/sub-locations/{id}", subLocationsHandler.DeleteSubLocation).Methods("DELETE")

	// Categories
	api.HandleFunc("/categories", categoriesHandler.GetCategories).Methods("GET")
	api.HandleFunc("/categories", categoriesHandler.CreateCategory).Methods("POST")
	api.HandleFunc("/categories/{id}", categoriesHandler.UpdateCategory).Methods("PUT")
	api.HandleFunc("/categories/{id}", categoriesHandler.DeleteCategory).Methods("DELETE")

	logrus.WithField("port", cfg.Server.Port).Info("JujuDB démarré")
	if err := http.ListenAndServe(":"+cfg.Server.Port, r); err != nil {
		logrus.WithError(err).Fatal("Erreur serveur HTTP")
	}
}
