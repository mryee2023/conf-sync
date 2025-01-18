package server

import (
	"os"
	"path/filepath"

	"github.com/mryee2023/conf-sync/internal/gist"
	"github.com/mryee2023/conf-sync/internal/logger"
)

// GistManager handles Gist operations for the server
type GistManager struct {
	client *gist.Client
}

// NewGistManager creates a new GistManager
func NewGistManager(token, gistID string) *GistManager {
	return &GistManager{
		client: gist.NewClient(token, gistID),
	}
}

// UploadFiles uploads multiple files to Gist
func (m *GistManager) UploadFiles(files []string) error {
	fileContents := make(map[string][]byte)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			logger.Error("Error reading %s: %v", file, err)
			continue
		}
		fileContents[filepath.Base(file)] = content
	}

	if err := m.client.UploadFiles(fileContents); err != nil {
		return err
	}

	logger.Info("Successfully uploaded %d file(s)", len(fileContents))
	return nil
}

// ListFiles lists all files in the Gist
func (m *GistManager) ListFiles() error {
	files, err := m.client.DownloadFiles()
	if err != nil {
		return err
	}

	logger.Info("Files in Gist:")
	for name, content := range files {
		if content.IsDeleted {
			logger.Info("- %s (deleted)", name)
		} else {
			logger.Info("- %s (%d bytes)", name, len(content.Content))
		}
	}
	return nil
}

// DeleteFiles deletes multiple files from Gist
func (m *GistManager) DeleteFiles(files []string) error {
	for _, file := range files {
		if err := m.client.DeleteFile(filepath.Base(file)); err != nil {
			logger.Error("Error deleting %s: %v", file, err)
			continue
		}
		logger.Info("Deleted %s", file)
	}
	return nil
}
