package utils

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"time"

	"path/filepath"

	"github.com/anish-sahoo/image-storage-api/internal/models"
)

func CleanFileName(filename string) string {
	filename = strings.ReplaceAll(filename, " ", "")
	filename = regexp.MustCompile(`[^a-zA-Z0-9.]`).ReplaceAllString(filename, "")
	return filename
}

func RenderFileHTML(f models.File) string {
	var b strings.Builder

	b.WriteString("<h3>")
	b.WriteString(template.HTMLEscapeString(f.Name))
	b.WriteString("</h3>")

	b.WriteString("<p>Type: ")
	b.WriteString(template.HTMLEscapeString(f.Filetype))
	b.WriteString("</p>")

	b.WriteString("<p>Size: ")
	b.WriteString(fmt.Sprintf("%.2f KB", float64(f.FileSizeBytes)/1024))
	b.WriteString("</p>")

	b.WriteString("<p>Created: ")
	b.WriteString(f.CreatedAt.Format("Jan 02, 2006 15:04"))
	b.WriteString("</p>")

	b.WriteString("<p>Tag: ")
	b.WriteString(template.HTMLEscapeString(f.Tag))
	b.WriteString("</p>")

	b.WriteString(fmt.Sprintf("<a href='/api/images/%d/download'", f.ID))
	b.WriteString(fmt.Sprintf("%d", f.ID))
	b.WriteString("'>Download</a>")

	return b.String()
}

func FileToPhotoResponse(file models.File) models.PhotoResponse {
	return models.PhotoResponse{
		ID:        file.ID,
		Name:      file.Name,
		URL:       fmt.Sprintf("/api/images/%d/download", file.ID),
		FileType:  file.Filetype,
		SizeBytes: file.FileSizeBytes,
		CreatedAt: file.CreatedAt.Format(time.RFC3339),
	}
}

func FileNameToContentType(file models.File) string {
	var contentType string
	switch strings.ToLower(filepath.Ext(file.Name)) {
	case ".pdf":
		contentType = "application/pdf"
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	default:
		contentType = "application/octet-stream"
	}

	return contentType
}
