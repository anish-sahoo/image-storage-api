package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/anish-sahoo/image-storage-api/internal/models"
	"github.com/anish-sahoo/image-storage-api/internal/utils"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

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
		w.Header().Set("HX-Redirect", "/login.html")
		w.WriteHeader(http.StatusUnauthorized)
		return "", false
	}
	username, err := ValidateJWT(cookie.Value)
	if err != nil || username == "" {
		w.Header().Set("HX-Redirect", "/login.html")
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
	}

	var b strings.Builder
	b.WriteString("<div class='file-item'>")
	for _, f := range files {
		b.WriteString(utils.RenderFileHTML(f))
	}
	b.WriteString("</div>")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}

func handleFileUpload(w http.ResponseWriter, r *http.Request, username string) {
	const maxUploadSize = 50 << 20 // 50 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	err := r.ParseMultipartForm(maxUploadSize)
	if err != nil {
		log.Err(err).Msg("Error parsing multipart form")
		http.Error(w, "File too big or bad request", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	tag := r.FormValue("tag")

	if err != nil {
		log.Err(err).Msg("Error retrieving file from form data")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	originalFileName := utils.CleanFileName(fileHeader.Filename)
	fileSize := fileHeader.Size

	location := "./data/images/" + originalFileName

	os.MkdirAll(filepath.Dir(location), os.ModePerm)
	dst, err := os.Create(location)

	if err != nil {
		log.Err(err).Msg("Error creating file on server")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		log.Err(err).Msg("Error saving file")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	user, err := getUserByUsername(username)
	if err != nil {
		log.Err(err).Msg("User not found")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	err = insertFile(models.File{
		Name:          originalFileName,
		Filetype:      path.Ext(originalFileName),
		Location:      location,
		OwnerId:       user.ID,
		FileSizeBytes: int64(fileSize),
		Tag:           tag,
	})
	if err != nil {
		log.Err(err).Msg("Error saving file metadata")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Info().
		Str("username", username).
		Str("filename", originalFileName).
		Int64("size_bytes", fileSize).
		Msg("File uploaded successfully")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", "refresh-file-list")
	err = serveHTMLFile(w, "upload-success.html")
	if err != nil {
		log.Err(err).Msg("Error serving upload success template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

	data := struct {
		Username string
	}{
		Username: username,
	}
	serveHTMLTemplate(w, "authenticated.html", data)
}

// api/images/
func AllImagesAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	tag := r.URL.Query().Get("tag")

	if limitStr == "" || offsetStr == "" || tag == "" {
		http.Error(w, "Missing query parameters", http.StatusBadRequest)
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
	json.NewEncoder(w).Encode(resp)
}

// api/images/{id}/download
func ImageDownloadAPIHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Err(err).Str("vars", fmt.Sprintf("%v", vars)).Msg(" Error atoi ImageDownloadAPIHandler")
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	file, err := getFile(id)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", utils.FileNameToContentType(file))
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", file.Name))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.FileSizeBytes))

	f, err := os.Open(file.Location)
	if err != nil {
		http.Error(w, "Unable to open file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, "Error sending file", http.StatusInternalServerError)
	}
}

func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	log.Info().Int("id", id).Msg("Deleting")
	w.Header().Set("HX-Trigger", "refresh-file-list")
	w.WriteHeader(http.StatusOK)
}
