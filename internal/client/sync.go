package client

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/mryee2023/conf-sync/internal/gist"
	"github.com/mryee2023/conf-sync/internal/logger"
)

// FileMapping represents a mapping between a Gist file and a local file
type FileMapping struct {
	GistFile    string
	LocalFile   string
	LastModify  time.Time
	ExecCommand string // Command to execute after update
}

// SyncOnce performs a single sync operation
func SyncOnce(client *gist.Client, mappings []FileMapping) error {
	files, err := client.DownloadFiles()
	if err != nil {
		return err
	}

	var updated bool
	for _, m := range mappings {
		file, ok := files[m.GistFile]
		if !ok {
			logger.Warn("File %s not found in Gist", m.GistFile)
			continue
		}

		// Create parent directories if they don't exist
		if err := os.MkdirAll(filepath.Dir(m.LocalFile), 0755); err != nil {
			logger.Error("Failed to create directories for %s: %v", m.LocalFile, err)
			continue
		}

		// Write file
		if err := os.WriteFile(m.LocalFile, file.Content, 0644); err != nil {
			logger.Error("Failed to write %s: %v", m.LocalFile, err)
			continue
		}

		updated = true
		logger.Info("Successfully updated %s", m.LocalFile)

		// Execute command if specified
		if m.ExecCommand != "" {
			logger.Info("Executing command: %s", m.ExecCommand)
			cmd := exec.Command("sh", "-c", m.ExecCommand)
			if out, err := cmd.CombinedOutput(); err != nil {
				logger.Error("Command failed: %v\nOutput: %s", err, out)
			} else {
				logger.Info("Command executed successfully")
			}
		}
	}

	if !updated {
		logger.Info("No files were updated")
	}
	return nil
}

// WatchFiles continuously watches for file changes
func WatchFiles(client *gist.Client, mappings []FileMapping) {
	logger.Info("Starting file watcher...")
	currentInterval := client.GetMinInterval() // Start with minimum interval

	for {
		logger.Debug("Checking for updates at %v...", time.Now().UTC())
		files, err := client.DownloadFiles()
		if err != nil {
			if client.IsRateLimited(err) {
				// Increase interval on rate limit, but cap it
				currentInterval = time.Duration(float64(currentInterval) * gist.BackoffMultiplier)
				if currentInterval > gist.MaxBackoffInterval {
					currentInterval = gist.MaxBackoffInterval
				}
				logger.Warn("Rate limited. Increasing check interval to %v", currentInterval)
			} else {
				logger.Error("Failed to download files: %v", err)
			}
			time.Sleep(currentInterval)
			continue
		}

		// If we successfully got files, try to gradually reduce the interval
		if currentInterval > client.GetMinInterval() {
			currentInterval = time.Duration(float64(currentInterval) / gist.BackoffMultiplier)
			if currentInterval < client.GetMinInterval() {
				currentInterval = client.GetMinInterval()
			}
			logger.Info("Reducing check interval to %v", currentInterval)
		}

		logger.Debug("Found %d files in Gist", len(files))

		for i := range mappings {
			mapping := &mappings[i]
			file, ok := files[mapping.GistFile]
			if !ok {
				logger.Debug("File %s not found in Gist", mapping.GistFile)
				continue
			}

			logger.Debug("Checking Gist file: %s (updated at: %v)", mapping.GistFile, file.UpdatedAt)

			// Compare with last modification time
			if !file.UpdatedAt.After(mapping.LastModify) {
				logger.Debug("No changes for %s (last update: %v, last modify: %v)",
					mapping.GistFile, file.UpdatedAt, mapping.LastModify)
				continue
			}

			logger.Info("File %s has been updated, downloading...", mapping.GistFile)

			// Create parent directories if they don't exist
			if err := os.MkdirAll(filepath.Dir(mapping.LocalFile), 0755); err != nil {
				logger.Error("Failed to create directories for %s: %v", mapping.LocalFile, err)
				continue
			}

			// Write file
			if err := os.WriteFile(mapping.LocalFile, file.Content, 0644); err != nil {
				logger.Error("Failed to write %s: %v", mapping.LocalFile, err)
				continue
			}

			// Update last modification time
			mapping.LastModify = file.UpdatedAt

			logger.Info("Successfully updated %s", mapping.LocalFile)

			// Execute command if configured
			if mapping.ExecCommand != "" {
				logger.Info("Executing command: %s", mapping.ExecCommand)
				cmd := exec.Command("sh", "-c", mapping.ExecCommand)
				if out, err := cmd.CombinedOutput(); err != nil {
					logger.Error("Command failed: %v\nOutput: %s", err, out)
				} else {
					logger.Info("Command executed successfully")
				}
			}
		}

		time.Sleep(currentInterval)
	}
}
