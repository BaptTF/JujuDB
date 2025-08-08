package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
)

// AuthHandler handles authentication-related operations
type AuthHandler struct {
	Password string
	Store    *sessions.CookieStore
	TestMode bool
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(store *sessions.CookieStore) *AuthHandler {
	password := os.Getenv("APP_PASSWORD")
	if password == "" {
		password = "famille123" // fallback for development
	}
	
	// Check if we're in test mode
	testMode := os.Getenv("TEST_MODE") == "true"
	
	return &AuthHandler{
		Password: password,
		Store:    store,
		TestMode: testMode,
	}
}

// Login handles POST /login for authentication
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Ensures only POST is allowed
	if r.Method != http.MethodPost {
		logrus.WithFields(logrus.Fields{
			"handler": "auth",
			"action":  "Login",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Warn("Method not allowed for login")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}



	password := r.FormValue("password")
	if password != h.Password {
		logrus.WithFields(logrus.Fields{
			"handler": "auth",
			"action":  "Login",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Warn("Invalid password attempt")
		http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
		return
	}

	// Create session
	session, err := h.Store.Get(r, "jujudb-session")
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "auth",
			"action":  "Login",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Error("Failed to get session store")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// Set authentication in session
	session.Values["authenticated"] = true
	session.Values["login_time"] = time.Now().Unix()
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   30 * 24 * 3600, // 30 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	
	err = session.Save(r, w)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "auth",
			"action":  "Login",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Error("Failed to save session")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Root handles GET / and redirects depending on authentication state
func (h *AuthHandler) Root(w http.ResponseWriter, r *http.Request) {
    // Check if user is already authenticated
    if h.isAuthenticated(r) {
        logrus.WithFields(logrus.Fields{
            "handler": "auth",
            "action":  "Root",
            "method":  r.Method,
            "path":    r.URL.Path,
        }).Debug("Authenticated user redirected to /dashboard")
        http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
        return
    }
    logrus.WithFields(logrus.Fields{
        "handler": "auth",
        "action":  "Root",
        "method":  r.Method,
        "path":    r.URL.Path,
    }).Debug("Unauthenticated user redirected to /login")
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Logout handles POST /logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    // Get and clear session
    session, err := h.Store.Get(r, "jujudb-session")
    if err == nil {
        session.Values["authenticated"] = false
        session.Options.MaxAge = -1 // Delete the session
        session.Save(r, w)
    }

    logrus.WithFields(logrus.Fields{
        "handler": "auth",
        "action":  "Logout",
        "method":  r.Method,
        "path":    r.URL.Path,
    }).Info("User logged out, redirecting to root")
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

// AuthMiddleware checks if user is authenticated
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip auth for login page and static files
        if r.URL.Path == "/" || r.URL.Path == "/login" || r.URL.Path[0:8] == "/static/" {
            next.ServeHTTP(w, r)
            return
        }

        if !h.isAuthenticated(r) {
            logrus.WithFields(logrus.Fields{
                "handler": "auth",
                "action":  "AuthMiddleware",
                "method":  r.Method,
                "path":    r.URL.Path,
            }).Warn("Unauthenticated access, redirecting to /")
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        next.ServeHTTP(w, r)
    })
}

// isAuthenticated checks if the request has a valid session
func (h *AuthHandler) isAuthenticated(r *http.Request) bool {
	// In test mode, check for simple cookie for backward compatibility
	if h.TestMode {
		if cookie, err := r.Cookie("auth"); err == nil {
			if cookie.Value == "authenticated" || cookie.Value == "test-auth-token" {
				return true
			}
		}
	}
	
	// Get session
	session, err := h.Store.Get(r, "jujudb-session")
	if err != nil {
		return false
	}
	
	// Check if authenticated
	auth, ok := session.Values["authenticated"]
	if !ok {
		return false
	}
	
	authBool, ok := auth.(bool)
	return ok && authBool
}
