package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anish-sahoo/image-storage-api/internal/models"
	"github.com/anish-sahoo/image-storage-api/internal/utils"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func serveHTMLFile(w http.ResponseWriter, filename string) error {
	fp := filepath.Join("web", filename)
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html")
	return tmpl.Execute(w, nil)
}

func serveHTMLTemplate(w http.ResponseWriter, filename string, data interface{}) error {
	fp := filepath.Join("web", filename)
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html")
	return tmpl.Execute(w, data)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
		log.Warn().Str("username", username).Msg("Invalid login attempt")
		if err := serveHTMLFile(w, "login-error.html"); err != nil {
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
		MaxAge:   3600 * 24,
	})

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func ImagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		username, ok := authenticateRequest(w, r)
		if !ok {
			return
		}
		renderFileList(w, username)
	case http.MethodPost:
		username, ok := authenticateRequest(w, r)
		if !ok {
			return
		}
		handleFileUpload(w, r, username)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func authenticateRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusUnauthorized)
		return "", false
	}
	username, err := ValidateJWT(cookie.Value)
	if err != nil || username == "" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusUnauthorized)
		return "", false
	}
	return username, true
}

func renderFileList(w http.ResponseWriter, username string) {
	files, err := getFilesByUser(username)
	if err != nil {
		log.Err(err).Msg("Internal error in renderFileList")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	funcMap := template.FuncMap{
		"formatSize": utils.FormatFileSize,
		"formatDate": func(t time.Time) string { return t.Format("Jan 02, 2006") },
		"isImage": func(ext string) bool {
			switch strings.ToLower(ext) {
			case ".jpg", ".jpeg", ".png", ".gif", ".webp":
				return true
			}
			return false
		},
	}

	fp := filepath.Join("web", "file-list.html")
	tmpl, err := template.New("file-list.html").Funcs(funcMap).ParseFiles(fp)
	if err != nil {
		log.Err(err).Msg("Error parsing file-list template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, files); err != nil {
		log.Err(err).Msg("Error executing file-list template")
	}
}

func handleFileUpload(w http.ResponseWriter, r *http.Request, username string) {
	const maxUploadSize = 50 << 20 // 50 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		log.Err(err).Msg("Error parsing multipart form")
		http.Error(w, "File too large or bad request", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Err(err).Msg("Error retrieving file from form")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tag := r.FormValue("tag")
	objectName := utils.CleanFileName(fileHeader.Filename)
	fileSize := fileHeader.Size
	ext := strings.ToLower(path.Ext(objectName))
	contentType := utils.ExtToContentType(ext)

	if err := s3Upload(objectName, file, fileSize, contentType); err != nil {
		log.Err(err).Msg("Error uploading to S3")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	user, err := getUserByUsername(username)
	if err != nil {
		log.Err(err).Msg("User not found")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := insertFile(models.File{
		Name:          objectName,
		Filetype:      ext,
		Location:      objectName,
		OwnerId:       user.ID,
		FileSizeBytes: fileSize,
		Tag:           tag,
	}); err != nil {
		log.Err(err).Msg("Error saving file metadata")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Info().
		Str("username", username).
		Str("filename", objectName).
		Int64("size_bytes", fileSize).
		Msg("File uploaded successfully")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "refresh-file-list")
	if err := serveHTMLFile(w, "upload-success.html"); err != nil {
		log.Err(err).Msg("Error serving upload-success template")
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func AuthCheckHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("jwt")
	if err != nil {
		serveHTMLFile(w, "login-form.html")
		return
	}
	username, err := ValidateJWT(cookie.Value)
	if err != nil || username == "" {
		serveHTMLFile(w, "login-form.html")
		return
	}
	data := struct{ Username string }{Username: username}
	serveHTMLTemplate(w, "authenticated.html", data)
}

// GET /api/images — public paginated list for the website photos page
func AllImagesAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	tag := r.URL.Query().Get("tag")

	if limitStr == "" || offsetStr == "" || tag == "" {
		http.Error(w, "Missing query parameters: limit, offset, tag", http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		http.Error(w, "Invalid limit", http.StatusBadRequest)
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		http.Error(w, "Invalid offset", http.StatusBadRequest)
		return
	}

	files, total, err := getImages(tag, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch images", http.StatusInternalServerError)
		return
	}

	photos := make([]models.PhotoResponse, len(files))
	for i, f := range files {
		photos[i] = utils.FileToPhotoResponse(f)
	}

	resp := models.ListPhotosResponse{
		Photos: photos,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
	if offset+limit < total {
		resp.NextPage = fmt.Sprintf("/api/images?tag=%s&limit=%d&offset=%d", tag, limit, offset+limit)
	}
	if offset > 0 {
		prevOffset := offset - limit
		if prevOffset < 0 {
			prevOffset = 0
		}
		resp.PreviousPage = fmt.Sprintf("/api/images?tag=%s&limit=%d&offset=%d", tag, limit, prevOffset)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(resp)
}

// GET /api/images/{id}/download — streams file from MinIO
func ImageDownloadAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	file, err := getFile(id)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	obj, err := s3GetObject(file.Location)
	if err != nil {
		log.Err(err).Str("location", file.Location).Msg("Error retrieving from S3")
		http.Error(w, "Unable to retrieve file", http.StatusInternalServerError)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Type", utils.ExtToContentType(file.Filetype))
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", file.Name))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.FileSizeBytes))

	if _, err := io.Copy(w, obj); err != nil {
		log.Err(err).Msg("Error streaming file to response")
	}
}

// DELETE /images/{id}/delete — requires auth, removes from MinIO and DB
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	_, ok := authenticateRequest(w, r)
	if !ok {
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	file, err := getFile(id)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if err := s3DeleteObject(file.Location); err != nil {
		log.Err(err).Str("location", file.Location).Msg("Error deleting from S3")
	}

	if err := deleteFile(id); err != nil {
		log.Err(err).Int("id", id).Msg("Error deleting file from DB")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Info().Int("id", id).Str("name", file.Name).Msg("File deleted")
	w.WriteHeader(http.StatusOK)
}
