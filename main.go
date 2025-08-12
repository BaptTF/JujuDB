package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"jujudb/handlers"
	"jujudb/services"
)

var (
	db    *sql.DB
	store *sessions.CookieStore
	tmpl  *template.Template
)

type Item struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	Category    string    `json:"category"`
	Quantity    int       `json:"quantity"`
	ExpiryDate  *string   `json:"expiry_date"`
	AddedDate   time.Time `json:"added_date"`
}

type SearchResult struct {
	Item     Item    `json:"item"`
	Distance int     `json:"distance"`
	Score    float64 `json:"score"`
}

func init() {
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	
	// Check if we're in production (behind Traefik with HTTPS)
	isProduction := os.Getenv("PRODUCTION") == "true" || os.Getenv("HTTPS") == "true"
	
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   isProduction, // Enable secure cookies in production
		SameSite: http.SameSiteStrictMode,
	}

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
	
	// Database connection
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "jujudb"
	}
	
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "jujudb123"
	}
	
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "jujudb"
	}

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", 
		dbHost, dbUser, dbPassword, dbName)
	
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
	tmpl = template.Must(template.ParseGlob("templates/*.html"))

	// Instantiate handlers
	authHandler := handlers.NewAuthHandler(store)
	// Initialize Meilisearch service
	meilisearchHost := "http://meilisearch:7700"
	meilisearchKey := "jujudb-master-key-change-in-production"
	meilisearchService, err := services.NewMeilisearchService(meilisearchHost, meilisearchKey)
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
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logrus.WithField("port", port).Info("JujuDB démarré")
	if err := http.ListenAndServe(":"+port, r); err != nil {
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
