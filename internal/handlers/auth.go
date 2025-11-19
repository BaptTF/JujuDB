package handlers

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"jujudb/internal/config"
)

// AuthHandler handles authentication-related operations
type AuthHandler struct {
	Password string
	Store    *sessions.CookieStore
	TestMode bool
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(store *sessions.CookieStore, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		Password: cfg.Auth.Password,
		Store:    store,
		TestMode: cfg.Auth.TestMode,
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

	// Try to get existing session first
	session, err := h.Store.Get(r, "jujudb-session")
	if err != nil {
		// Session is invalid (likely due to key rotation), clear the cookie first
		logrus.WithError(err).WithFields(logrus.Fields{
			"handler": "auth",
			"action":  "Login",
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Info("Invalid session detected during login, clearing cookie and creating new session")

		// Clear the invalid cookie by setting it to expire
		http.SetCookie(w, &http.Cookie{
			Name:     "jujudb-session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   false, // Will be set correctly based on config
			SameSite: http.SameSiteStrictMode,
		})

		// Create a new request without the invalid cookie to avoid the error
		rCopy := *r
		rCopy.Header = make(http.Header)
		for k, v := range r.Header {
			rCopy.Header[k] = v
		}
		rCopy.Header.Del("Cookie")

		// Create new session with the clean request
		session, err = h.Store.New(&rCopy, "jujudb-session")
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"handler": "auth",
				"action":  "Login",
				"method":  r.Method,
				"path":    r.URL.Path,
			}).Error("Failed to create new session")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	// Set authentication in session
	session.Values["authenticated"] = true
	session.Values["login_time"] = time.Now().Unix()

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

		// Check if session is valid, if not clear the invalid cookie
		if !h.isSessionValid(r) {
			h.clearInvalidSession(w, r)
			logrus.WithFields(logrus.Fields{
				"handler": "auth",
				"action":  "AuthMiddleware",
				"method":  r.Method,
				"path":    r.URL.Path,
			}).Warn("Invalid session detected and cleared, redirecting to /")
			http.Redirect(w, r, "/", http.StatusSeeOther)
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
		// Session is invalid (likely due to key rotation), log and return false
		logrus.WithError(err).Debug("Invalid session detected, likely due to key rotation")
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

// isSessionValid checks if the session cookie is valid (can be decoded)
func (h *AuthHandler) isSessionValid(r *http.Request) bool {
	_, err := h.Store.Get(r, "jujudb-session")
	return err == nil
}

// clearInvalidSession clears an invalid session cookie
func (h *AuthHandler) clearInvalidSession(w http.ResponseWriter, r *http.Request) {
	// Create a new session and set it to expire immediately
	session, _ := h.Store.New(r, "jujudb-session")
	session.Options.MaxAge = -1
	session.Save(r, w)
}
