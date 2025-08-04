package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/texttheater/golang-levenshtein/levenshtein"

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
		log.Fatal("Erreur de connexion à la base de données:", err)
	}
	defer db.Close()

	// Test database connection
	if err = db.Ping(); err != nil {
		log.Fatal("Impossible de se connecter à la base de données:", err)
	}

	// Initialize database
	initDB()

	// Load templates
	tmpl = template.Must(template.ParseGlob("templates/*.html"))

	// Routes
	r := mux.NewRouter()
	
	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	
	// Authentication routes
	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", logoutHandler).Methods("POST")
	
	// Protected routes
	r.HandleFunc("/dashboard", requireAuth(dashboardHandler)).Methods("GET")
	
	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.Use(requireAuthMiddleware)
	api.HandleFunc("/items", getItemsHandler).Methods("GET")
	api.HandleFunc("/items", createItemHandler).Methods("POST")
	api.HandleFunc("/items/{id}", updateItemHandler).Methods("PUT")
	api.HandleFunc("/items/{id}", deleteItemHandler).Methods("DELETE")
	api.HandleFunc("/search", searchHandler).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("JujuDB démarré sur le port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func initDB() {
	query := `
	CREATE TABLE IF NOT EXISTS items (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		location VARCHAR(100) NOT NULL DEFAULT 'Congélateur',
		category VARCHAR(100),
		quantity INTEGER DEFAULT 1,
		expiry_date DATE,
		added_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_items_name ON items(name);
	CREATE INDEX IF NOT EXISTS idx_items_location ON items(location);
	CREATE INDEX IF NOT EXISTS idx_items_category ON items(category);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatal("Erreur lors de l'initialisation de la base de données:", err)
	}
}

func requireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "jujudb-session")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		handler(w, r)
	}
}

func requireAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "jujudb-session")
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Error(w, "Non autorisé", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "jujudb-session")
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		password := r.FormValue("password")
		correctPassword := os.Getenv("APP_PASSWORD")
		if correctPassword == "" {
			correctPassword = "famille123" // Default password
		}

		if password == correctPassword {
			session, _ := store.Get(r, "jujudb-session")
			session.Values["authenticated"] = true
			session.Save(r, w)
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		tmpl.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Error": "Mot de passe incorrect",
		})
		return
	}

	tmpl.ExecuteTemplate(w, "login.html", nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "jujudb-session")
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "dashboard.html", nil)
}

func getItemsHandler(w http.ResponseWriter, r *http.Request) {
	location := r.URL.Query().Get("location")
	category := r.URL.Query().Get("category")

	query := "SELECT id, name, description, location, category, quantity, expiry_date, added_date FROM items WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if location != "" {
		argCount++
		query += fmt.Sprintf(" AND location = $%d", argCount)
		args = append(args, location)
	}

	if category != "" {
		argCount++
		query += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, category)
	}

	query += " ORDER BY added_date DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		var expiryDate sql.NullString
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Location, 
			&item.Category, &item.Quantity, &expiryDate, &item.AddedDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if expiryDate.Valid {
			item.ExpiryDate = &expiryDate.String
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func createItemHandler(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `INSERT INTO items (name, description, location, category, quantity, expiry_date) 
			  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, added_date`
	
	var expiryDate interface{}
	if item.ExpiryDate != nil && *item.ExpiryDate != "" {
		expiryDate = *item.ExpiryDate
	}

	err := db.QueryRow(query, item.Name, item.Description, item.Location, 
		item.Category, item.Quantity, expiryDate).Scan(&item.ID, &item.AddedDate)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func updateItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `UPDATE items SET name=$1, description=$2, location=$3, category=$4, 
			  quantity=$5, expiry_date=$6 WHERE id=$7`
	
	var expiryDate interface{}
	if item.ExpiryDate != nil && *item.ExpiryDate != "" {
		expiryDate = *item.ExpiryDate
	}

	_, err = db.Exec(query, item.Name, item.Description, item.Location, 
		item.Category, item.Quantity, expiryDate, id)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	item.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM items WHERE id=$1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	location := r.URL.Query().Get("location")
	category := r.URL.Query().Get("category")

	if query == "" {
		http.Error(w, "Paramètre de recherche manquant", http.StatusBadRequest)
		return
	}

	sqlQuery := "SELECT id, name, description, location, category, quantity, expiry_date, added_date FROM items WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if location != "" {
		argCount++
		sqlQuery += fmt.Sprintf(" AND location = $%d", argCount)
		args = append(args, location)
	}

	if category != "" {
		argCount++
		sqlQuery += fmt.Sprintf(" AND category = $%d", argCount)
		args = append(args, category)
	}

	rows, err := db.Query(sqlQuery, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []SearchResult
	queryLower := strings.ToLower(query)

	for rows.Next() {
		var item Item
		var expiryDate sql.NullString
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Location, 
			&item.Category, &item.Quantity, &expiryDate, &item.AddedDate)
		if err != nil {
			continue
		}
		if expiryDate.Valid {
			item.ExpiryDate = &expiryDate.String
		}

		// Calculate Levenshtein distance
		nameLower := strings.ToLower(item.Name)
		descLower := strings.ToLower(item.Description)
		
		nameDistance := levenshtein.DistanceForStrings([]rune(queryLower), []rune(nameLower), levenshtein.DefaultOptions)
		descDistance := levenshtein.DistanceForStrings([]rune(queryLower), []rune(descLower), levenshtein.DefaultOptions)
		
		// Use the minimum distance
		distance := nameDistance
		if descDistance < nameDistance {
			distance = descDistance
		}

		// Calculate score (lower distance = higher score)
		maxLen := len(queryLower)
		if len(nameLower) > maxLen {
			maxLen = len(nameLower)
		}
		if len(descLower) > maxLen {
			maxLen = len(descLower)
		}

		score := 1.0 - (float64(distance) / float64(maxLen))
		
		// Only include results with reasonable similarity
		if score > 0.3 || strings.Contains(nameLower, queryLower) || strings.Contains(descLower, queryLower) {
			results = append(results, SearchResult{
				Item:     item,
				Distance: distance,
				Score:    score,
			})
		}
	}

	// Sort by score (highest first)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
