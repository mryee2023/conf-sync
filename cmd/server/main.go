package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/mryee2023/conf-sync/internal/logger"
	"github.com/mryee2023/conf-sync/internal/server"
)

var (
	gistID   string
	logLevel string
)

const (
	envGistToken = "GIST_TOKEN"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "conf-sync-server",
		Short: "Server for conf-sync",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level, err := logger.ParseLevel(logLevel)
			if err != nil {
				logger.Fatal("%v", err)
			}
			logger.SetLevel(level)

			if gistID == "" {
				logger.Fatal("Gist ID is required")
			}

			token := os.Getenv(envGistToken)
			if token == "" {
				logger.Fatal("GitHub token is required. Set GIST_TOKEN environment variable")
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&gistID, "gist-id", "g", "", "Gist ID")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")

	var cmdUpload = &cobra.Command{
		Use:   "upload [files...]",
		Short: "Upload files to Gist",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			manager := server.NewGistManager(os.Getenv(envGistToken), gistID)
			if err := manager.UploadFiles(args); err != nil {
				logger.Fatal("%v", err)
			}
		},
	}

	var cmdList = &cobra.Command{
		Use:   "list",
		Short: "List files in Gist",
		Run: func(cmd *cobra.Command, args []string) {
			manager := server.NewGistManager(os.Getenv(envGistToken), gistID)
			if err := manager.ListFiles(); err != nil {
				logger.Fatal("%v", err)
			}
		},
	}

	var cmdDelete = &cobra.Command{
		Use:   "delete [files...]",
		Short: "Delete files from Gist",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			manager := server.NewGistManager(os.Getenv(envGistToken), gistID)
			if err := manager.DeleteFiles(args); err != nil {
				logger.Fatal("%v", err)
			}
		},
	}

	rootCmd.AddCommand(cmdUpload, cmdList, cmdDelete)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("%v", err)
		os.Exit(1)
	}
}
