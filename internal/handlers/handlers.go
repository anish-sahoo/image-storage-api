package handlers

import (
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// serveHTMLFile serves an HTML file from the web directory
func serveHTMLFile(w http.ResponseWriter, filename string) error {
	filepath := filepath.Join("web", filename)
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	w.Header().Set("Content-Type", "text/html")
	_, err = io.Copy(w, file)
	return err
}

// serveHTMLTemplate serves an HTML template with data
func serveHTMLTemplate(w http.ResponseWriter, filename string, data interface{}) error {
	filepath := filepath.Join("web", filename)
	tmpl, err := template.ParseFiles(filepath)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")
	return tmpl.Execute(w, data)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Warn().Msg("Invalid Method")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !CheckCredentials(username, password) {
		log.Warn().Str("username", username).Str("password", password).Msg("Could not find user")
		err := serveHTMLFile(w, "login-error.html")
		if err != nil {
			log.Err(err).Msg("Error serving login error template")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	token, err := GenerateJWT(username)
	if err != nil {
		log.Err(err).Msg("Error generating JWT")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
	})

	// If login was submitted from the dedicated login page, redirect to index
	// Otherwise, if submitted from within index, we can return the authenticated fragment.
	referer := r.Header.Get("Referer")
	if strings.Contains(referer, "/login.html") || strings.Contains(referer, "/index.html") || referer == "/" || referer == "" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}
	data := struct {
		Username string
	}{
		Username: username,
	}
	if err := serveHTMLTemplate(w, "authenticated.html", data); err != nil {
		log.Err(err).Msg("Error serving authenticated template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func ImagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// For htmx, return an HTML fragment of the file list
		cookie, err := r.Cookie("jwt")
		if err != nil {
			// Ask htmx to redirect to login page
			w.Header().Set("HX-Redirect", "/login.html")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		username, err := ValidateJWT(cookie.Value)
		if err != nil || username == "" {
			w.Header().Set("HX-Redirect", "/login.html")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For now, use sample data instead of DB retrieval
		sample := []struct {
			Name string
		}{
			{Name: "sunset.jpg"},
			{Name: "mountains.png"},
			{Name: "city.gif"},
		}
		// Build simple HTML list
		var b strings.Builder
		for _, f := range sample {
			b.WriteString("<div class='file-item'>")
			b.WriteString(template.HTMLEscapeString(f.Name))
			b.WriteString("</div>")
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(b.String()))
	case http.MethodPost:
		cookie, err := r.Cookie("jwt")
		if err != nil {
			log.Err(err).Msg("Error checking JWT Cookie")
			w.Header().Set("HX-Redirect", "/login.html")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		username, err := ValidateJWT(cookie.Value)
		if err != nil || username == "" {
			log.Err(err).Msg("Error validating JWT")
			w.Header().Set("HX-Redirect", "/login.html")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// In a real app, you'd parse the multipart form and save the file.
		// For now, just return a small success fragment and ask htmx to refresh the list.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Trigger a refresh of the file list via htmx
		w.Header().Set("HX-Trigger", "refresh-file-list")
		err = serveHTMLFile(w, "upload-success.html")
		if err != nil {
			log.Err(err).Msg("Error serving upload success template")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear the JWT cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Expire immediately
	})

	// Redirect to main page using HTMX
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func AuthCheckHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		// Not authenticated - return login form only
		serveHTMLFile(w, "login-form.html")
		return
	}

	username, err := ValidateJWT(cookie.Value)
	if err != nil || username == "" {
		// Invalid token - return login form only
		serveHTMLFile(w, "login-form.html")
		return
	}

	// Authenticated - return upload interface only
	data := struct {
		Username string
	}{
		Username: username,
	}
	serveHTMLTemplate(w, "authenticated.html", data)
}

// JSON API for images (sample data for now)
func ImagesAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	// Optional: auth check (same as HTML handler)
	if _, err := r.Cookie("jwt"); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	sample := []map[string]string{
		{"name": "sunset.jpg"},
		{"name": "mountains.png"},
		{"name": "city.gif"},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sample)
}
