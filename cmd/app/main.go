package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/mryee2023/conf-sync/internal/gist"
	"github.com/mryee2023/conf-sync/internal/logger"
	"github.com/mryee2023/conf-sync/internal/config"
)

var (
	gistID string
	watchInterval time.Duration
	execCommand string
	logLevel string
	configFile string
)

const (
	envGistToken = "GIST_TOKEN"
)

func writeFile(filename string, content []byte) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	return os.WriteFile(filename, content, 0644)
}

func executeCommand(cmd string) error {
	if cmd == "" {
		return nil
	}
	logger.Info("Executing command: %s", cmd)
	command := exec.Command("sh", "-c", cmd)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}

func watchFiles(client *gist.Client, mappings []FileMapping) error {
	logger.Info("Watching %d file(s) for changes (check interval: %v)...", len(mappings), watchInterval)
	if execCommand != "" {
		logger.Info("Will execute command after updates: %s", execCommand)
	}
	
	// Create a map for quick lookup
	gistToLocal := make(map[string]string)
	for _, m := range mappings {
		gistToLocal[filepath.Base(m.GistFile)] = m.LocalFile
		logger.Debug("Mapping Gist file '%s' to local file '%s'", m.GistFile, m.LocalFile)
		// Ensure the local directory exists
		dir := filepath.Dir(m.LocalFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %v", m.LocalFile, err)
		}
	}

	lastCheck := time.Now()
	logger.Debug("Starting watch loop at %v", lastCheck)
	
	for {
		time.Sleep(watchInterval)
		currentTime := time.Now()
		logger.Debug("Checking for updates at %v...", currentTime)

		gistFiles, err := client.DownloadFiles()
		if err != nil {
			logger.Error("Error getting files: %v", err)
			continue
		}

		logger.Debug("Found %d files in Gist", len(gistFiles))
		hasUpdates := false
		for gistName, content := range gistFiles {
			logger.Debug("Checking Gist file: %s (updated at: %v)", gistName, content.UpdatedAt)
			localPath, exists := gistToLocal[gistName]
			if !exists {
				logger.Debug("Skipping file %s as it's not in our watch list", gistName)
				continue
			}

			if content.UpdatedAt.After(lastCheck) {
				if content.IsDeleted {
					logger.Info("File %s was deleted from Gist", gistName)
					continue
				}
				logger.Info("Changes detected in %s, updating local file %s...", gistName, localPath)
				if err := writeFile(localPath, content.Content); err != nil {
					logger.Error("Error writing file %s: %v", localPath, err)
					continue
				}
				logger.Info("Updated %s", localPath)
				hasUpdates = true
			} else {
				logger.Debug("No changes for %s (last update: %v, last check: %v)", 
					gistName, content.UpdatedAt, lastCheck)
			}
		}

		if hasUpdates && execCommand != "" {
			if err := executeCommand(execCommand); err != nil {
				logger.Error("Error executing command: %v", err)
			}
		}

		lastCheck = currentTime
	}
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "conf-sync",
		Short: "Sync configuration files with GitHub Gist",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level, err := logger.ParseLevel(logLevel)
			if err != nil {
				logger.Fatal("%v", err)
			}
			logger.SetLevel(level)
			
			if os.Getenv(envGistToken) == "" {
				logger.Fatal("GitHub token is required. Set GIST_TOKEN environment variable")
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&gistID, "gist-id", "g", "", "Gist ID")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")

	var cmdClient = &cobra.Command{
		Use:   "client",
		Short: "Run in client mode",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadClientConfig(configFile)
			if err != nil {
				logger.Fatal("Failed to load config: %v", err)
			}

			interval, err := cfg.GetCheckInterval()
			if err != nil {
				logger.Fatal("Invalid check interval: %v", err)
			}

			client := gist.NewClient(os.Getenv(envGistToken), cfg.GistID)
			var mappings []FileMapping
			for _, m := range cfg.Mappings {
				mappings = append(mappings, FileMapping{
					GistFile:  m.GistFile,
					LocalFile: m.LocalPath,
				})
				if m.Exec != "" {
					logger.Info("Will execute '%s' after updating %s", m.Exec, m.LocalPath)
				}
			}

			logger.Info("Starting client mode with %d file mappings", len(mappings))
			logger.Info("Check interval: %v", interval)
			watchInterval = interval
			watchFiles(client, mappings)
		},
	}
	cmdClient.Flags().StringVarP(&configFile, "config", "c", "/etc/conf-sync/client.yaml", "Path to client config file")

	// Server mode commands
	var cmdUpload = &cobra.Command{
		Use:   "upload [files...]",
		Short: "Upload files to Gist (server mode)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if gistID == "" {
				logger.Fatal("Gist ID is required in server mode")
			}
			client := gist.NewClient(os.Getenv(envGistToken), gistID)
			files := make(map[string][]byte)
			for _, file := range args {
				content, err := os.ReadFile(file)
				if err != nil {
					logger.Error("Error reading %s: %v", file, err)
					continue
				}
				files[filepath.Base(file)] = content
			}
			if err := client.UploadFiles(files); err != nil {
				logger.Fatal("%v", err)
			}
			logger.Info("Successfully uploaded %d file(s)", len(files))
		},
	}

	var cmdList = &cobra.Command{
		Use:   "list",
		Short: "List files in Gist (server mode)",
		Run: func(cmd *cobra.Command, args []string) {
			if gistID == "" {
				logger.Fatal("Gist ID is required in server mode")
			}
			client := gist.NewClient(os.Getenv(envGistToken), gistID)
			files, err := client.DownloadFiles()
			if err != nil {
				logger.Fatal("%v", err)
			}
			logger.Info("Files in Gist:")
			for name, content := range files {
				if content.IsDeleted {
					logger.Info("- %s (deleted)", name)
				} else {
					logger.Info("- %s (%d bytes)", name, len(content.Content))
				}
			}
		},
	}

	var cmdDelete = &cobra.Command{
		Use:   "delete [files...]",
		Short: "Delete files from Gist (server mode)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if gistID == "" {
				logger.Fatal("Gist ID is required in server mode")
			}
			client := gist.NewClient(os.Getenv(envGistToken), gistID)
			for _, file := range args {
				if err := client.DeleteFile(filepath.Base(file)); err != nil {
					logger.Error("Error deleting %s: %v", file, err)
					continue
				}
				logger.Info("Deleted %s", file)
			}
		},
	}

	// Add commands
	rootCmd.AddCommand(cmdClient)
	rootCmd.AddCommand(cmdUpload, cmdList, cmdDelete)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("%v", err)
	}
}

// FileMapping represents a mapping between a Gist file and a local file
type FileMapping struct {
	GistFile   string // filename in Gist
	LocalFile  string // local file path
}
