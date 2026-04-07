package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/anish-sahoo/image-storage-api/internal/models"
)

func CleanFileName(filename string) string {
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = regexp.MustCompile(`[^a-zA-Z0-9._-]`).ReplaceAllString(filename, "")
	return filename
}

func ExtToContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	default:
		return "application/octet-stream"
	}
}

func FormatFileSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
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
