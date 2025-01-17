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

func watchFiles(client *gist.Client, mappings []FileMapping) {
	logger.Info("Starting file watcher...")
	for {
		logger.Debug("Checking for updates at %v...", time.Now().UTC())
		files, err := client.DownloadFiles()
		if err != nil {
			logger.Error("Failed to download files: %v", err)
			time.Sleep(watchInterval)
			continue
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
			for _, m := range mappings {
				if m.LocalFile == mapping.LocalFile && m.ExecCommand != "" {
					logger.Info("Executing command: %s", m.ExecCommand)
					cmd := exec.Command("sh", "-c", m.ExecCommand)
					if out, err := cmd.CombinedOutput(); err != nil {
						logger.Error("Command failed: %v\nOutput: %s", err, out)
					} else {
						logger.Info("Command executed successfully")
					}
				}
			}
		}

		time.Sleep(watchInterval)
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

			// Use read-only client for client mode
			client := gist.NewReadOnlyClient(cfg.GistID)
			var mappings []FileMapping
			for _, m := range cfg.Mappings {
				mappings = append(mappings, FileMapping{
					GistFile:    m.GistFile,
					LocalFile:   m.LocalPath,
					LastModify:  time.Time{}, // Initialize last modification time
					ExecCommand: m.Exec,
				})
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
			token := os.Getenv(envGistToken)
			if token == "" {
				logger.Fatal("GitHub token is required for server mode. Set GIST_TOKEN environment variable")
			}
			client := gist.NewClient(token, gistID)
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
			token := os.Getenv(envGistToken)
			if token == "" {
				logger.Fatal("GitHub token is required for server mode. Set GIST_TOKEN environment variable")
			}
			client := gist.NewClient(token, gistID)
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
			token := os.Getenv(envGistToken)
			if token == "" {
				logger.Fatal("GitHub token is required for server mode. Set GIST_TOKEN environment variable")
			}
			client := gist.NewClient(token, gistID)
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
	GistFile    string
	LocalFile   string
	LastModify  time.Time
	ExecCommand string // Command to execute after update
}
