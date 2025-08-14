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
	return fmt.Sprintf(`
<div id="file-%d" class="file-item">
    <h3>%s</h3>
    <p>Type: %s</p>
    <p>Size: %.2f KB</p>
    <p>Created: %s</p>
    <p>Tag: %s</p>
    <a href="/api/images/%d/download">Download</a>
    <!--
	<button hx-delete="/images/%d/delete"
            hx-target="#file-%d"
            hx-swap="outerHTML"
            hx-confirm="Are you sure you want to delete '%s'?">
        Delete
    </button>
	-->
</div>
`, f.ID,
		template.HTMLEscapeString(f.Name),
		template.HTMLEscapeString(f.Filetype),
		float64(f.FileSizeBytes)/1024,
		f.CreatedAt.Format("Jan 02, 2006 15:04"),
		template.HTMLEscapeString(f.Tag),
		f.ID,
		f.ID,
		f.ID,
		template.HTMLEscapeString(f.Name),
	)
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
