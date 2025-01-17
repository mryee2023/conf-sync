package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/mryee2023/conf-sync/internal/gist"
	"github.com/mryee2023/conf-sync/internal/logger"
)

var (
	gistID string
	watchInterval time.Duration
	execCommand string
	logLevel string
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
				logger.Fatal("Invalid log level: %v", err)
			}
			logger.SetLevel(level)
			
			if os.Getenv(envGistToken) == "" {
				logger.Fatal("GitHub token is required. Set GIST_TOKEN environment variable")
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&gistID, "gist-id", "g", "", "Gist ID")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	rootCmd.MarkPersistentFlagRequired("gist-id")

	var cmdUpload = &cobra.Command{
		Use:   "upload [files...]",
		Short: "Upload files to Gist",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
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

	var cmdWatch = &cobra.Command{
		Use:   "watch [gist_file:local_path...]",
		Short: "Watch for Gist changes and update local files",
		Long: `Watch for changes in Gist files and update corresponding local files.
Example: conf-sync watch db.conf:/etc/myapp/db.conf config.yaml:/etc/myapp/config.yaml

The -i/--interval flag can be used to set the check interval (default: 10s).
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

The -e/--exec flag can be used to specify a command to execute after files are updated:
Example: conf-sync watch -e "docker restart myapp" db.conf:/etc/myapp/db.conf`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := gist.NewClient(os.Getenv(envGistToken), gistID)
			var mappings []FileMapping
			for _, arg := range args {
				parts := strings.SplitN(arg, ":", 2)
				if len(parts) != 2 {
					logger.Fatal("Invalid mapping format: %s", arg)
				}
				gistFile := parts[0]
				localFile := parts[1]
				mappings = append(mappings, FileMapping{
					GistFile:  gistFile,
					LocalFile: localFile,
				})
			}
			watchFiles(client, mappings)
		},
	}
	cmdWatch.Flags().DurationVarP(&watchInterval, "interval", "i", 10*time.Second, "Check interval for file changes")
	cmdWatch.Flags().StringVarP(&execCommand, "exec", "e", "", "Command to execute after files are updated")

	var cmdList = &cobra.Command{
		Use:   "list",
		Short: "List files in Gist",
		Run: func(cmd *cobra.Command, args []string) {
			client := gist.NewClient(os.Getenv(envGistToken), gistID)
			files, err := client.DownloadFiles()
			if err != nil {
				logger.Errorf("Error getting files: %v", err)
				os.Exit(1)
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

	var cmdSync = &cobra.Command{
		Use:   "sync [files...]",
		Short: "Sync files from Gist",
		Run: func(cmd *cobra.Command, args []string) {
			client := gist.NewClient(os.Getenv(envGistToken), gistID)
			files, err := client.DownloadFiles()
			if err != nil {
				logger.Errorf("Error getting files: %v", err)
				os.Exit(1)
			}

			for filename, content := range files {
				if content.IsDeleted {
					logger.Info("File %s has been deleted from gist", filename)
					continue
				}
				if len(args) > 0 {
					found := false
					for _, arg := range args {
						if filepath.Base(arg) == filename {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}
				if err := writeFile(filename, content.Content); err != nil {
					logger.Errorf("Error writing %s: %v", filename, err)
					continue
				}
				logger.Info("Synced %s", filename)
			}
		},
	}

	var cmdDelete = &cobra.Command{
		Use:   "delete [files...]",
		Short: "Delete files from Gist",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := gist.NewClient(os.Getenv(envGistToken), gistID)
			for _, file := range args {
				if err := client.DeleteFile(filepath.Base(file)); err != nil {
					logger.Errorf("Error deleting %s: %v", file, err)
					continue
				}
				logger.Info("Deleted %s", file)
			}
		},
	}

	rootCmd.AddCommand(cmdUpload, cmdWatch, cmdList, cmdSync, cmdDelete)

	if err := rootCmd.Execute(); err != nil {
		logger.Errorf("%v", err)
		os.Exit(1)
	}
}

// FileMapping represents a mapping between a Gist file and a local file
type FileMapping struct {
	GistFile   string // filename in Gist
	LocalFile  string // local file path
}
