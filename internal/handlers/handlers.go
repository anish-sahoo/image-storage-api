package handlers

import (
	"fmt"
	"net/http"
)

func ImagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		fmt.Fprintln(w, "GET request received!")
	case http.MethodPost:
		username, password, ok := r.BasicAuth()
		if !ok || !CheckCredentials(username, password) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		fmt.Fprintln(w, "POST request received and authorized!")
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
