package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"jujudb/internal/config"
	"jujudb/internal/handlers"
	"jujudb/internal/services"
)

var (
	db    *sql.DB
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
	var err error

	// Load configuration
	cfg := config.Load()

	// Initialize session store
	store = sessions.NewCookieStore([]byte(cfg.Session.Key))

	// Database connection
	connStr := cfg.Database.GetConnectionString()

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		logrus.WithError(err).Fatal("Erreur de connexion à la base de données")
	}
	defer db.Close()

	// Test database connection
	if err = db.Ping(); err != nil {
		logrus.WithError(err).Fatal("Impossible de se connecter à la base de données")
	}

	// Initialize database
	initDB()

	// Load templates (legacy handlers may rely on this)
	tmpl = template.Must(template.ParseGlob("web/static/html/*.html"))

	// Instantiate handlers
	authHandler := handlers.NewAuthHandler(store)

	// Initialize Meilisearch service
	meilisearchService, err := services.NewMeilisearchService(cfg.Meilisearch.Host, cfg.Meilisearch.MasterKey)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize Meilisearch service")
	}

	// Initialize sync service
	syncService := services.NewSyncService(db, meilisearchService)

	templatesHandler := handlers.NewTemplatesHandler()
	articlesHandler := handlers.NewArticlesHandler(db, syncService)
	locationsHandler := handlers.NewLocationsHandler(db)
	subLocationsHandler := handlers.NewSubLocationsHandler(db)
	categoriesHandler := handlers.NewCategoriesHandler(db)
	searchHandler := handlers.NewSearchHandler(db, meilisearchService)
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

func initDB() {
	query := `
	-- Core reference tables
	CREATE TABLE IF NOT EXISTS locations (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS categories (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL UNIQUE
	);

	CREATE TABLE IF NOT EXISTS sub_locations (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		location_id INTEGER NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
		UNIQUE(name, location_id)
	);

	-- Items table (normalized)
	CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL,
		sub_location_id INTEGER REFERENCES sub_locations(id) ON DELETE SET NULL,
		category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
		quantity INTEGER DEFAULT 1,
		expiry_date DATE,
		added_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		notes TEXT
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_items_name ON items(name);
	CREATE INDEX IF NOT EXISTS idx_items_location_id ON items(location_id);
	CREATE INDEX IF NOT EXISTS idx_items_sub_location_id ON items(sub_location_id);
	CREATE INDEX IF NOT EXISTS idx_items_category_id ON items(category_id);
	CREATE INDEX IF NOT EXISTS idx_items_expiry_date ON items(expiry_date);

	-- Migrations for older schemas
	ALTER TABLE items ADD COLUMN IF NOT EXISTS location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL;
	ALTER TABLE items ADD COLUMN IF NOT EXISTS sub_location_id INTEGER REFERENCES sub_locations(id) ON DELETE SET NULL;
	ALTER TABLE items ADD COLUMN IF NOT EXISTS category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL;
	ALTER TABLE items ADD COLUMN IF NOT EXISTS notes TEXT;
	ALTER TABLE items ADD COLUMN IF NOT EXISTS location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL;
	ALTER TABLE items ADD COLUMN IF NOT EXISTS sub_location_id INTEGER REFERENCES sub_locations(id) ON DELETE SET NULL;
	ALTER TABLE items ADD COLUMN IF NOT EXISTS category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL;
	ALTER TABLE items ADD COLUMN IF NOT EXISTS notes TEXT;

	-- Seed base reference data
	INSERT INTO locations (name) VALUES ('Congélateur'), ('Réfrigérateur'), ('Garde-manger') ON CONFLICT DO NOTHING;
	INSERT INTO categories (name) VALUES ('Viande'), ('Légumes'), ('Desserts'), ('Poisson'), ('Autres') ON CONFLICT DO NOTHING;
	`

	_, err := db.Exec(query)
	if err != nil {
		logrus.WithError(err).Fatal("Erreur lors de l'initialisation de la base de données")
	}
}
