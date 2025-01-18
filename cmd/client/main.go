package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/mryee2023/conf-sync/internal/client"
	"github.com/mryee2023/conf-sync/internal/config"
	"github.com/mryee2023/conf-sync/internal/gist"
	"github.com/mryee2023/conf-sync/internal/logger"
)

var (
	configFile string
	logLevel   string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "conf-sync-client",
		Short: "Client for conf-sync",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level, err := logger.ParseLevel(logLevel)
			if err != nil {
				logger.Fatal("%v", err)
			}
			logger.SetLevel(level)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "/etc/conf-sync/client.yaml", "Path to client config file")

	var cmdWatch = &cobra.Command{
		Use:   "watch",
		Short: "Watch for configuration changes",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadClientConfig(configFile)
			if err != nil {
				logger.Fatal("Failed to load config: %v", err)
			}

			interval, err := cfg.GetCheckInterval()
			if err != nil {
				logger.Fatal("Invalid check interval: %v", err)
			}

			// Create client and mappings
			gistClient := gist.NewReadOnlyClient(cfg.GistID)
			var mappings []client.FileMapping
			for _, m := range cfg.Mappings {
				mappings = append(mappings, client.FileMapping{
					GistFile:    m.GistFile,
					LocalFile:   m.LocalPath,
					LastModify:  gist.MinTime,
					ExecCommand: m.Exec,
				})
			}

			logger.Info("Starting client mode with %d file mappings", len(mappings))
			logger.Info("Check interval: %v", interval)
			client.WatchFiles(gistClient, mappings)
		},
	}

	var cmdSync = &cobra.Command{
		Use:   "sync",
		Short: "Sync files once",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.LoadClientConfig(configFile)
			if err != nil {
				logger.Fatal("Failed to load config: %v", err)
			}

			// Create client and mappings
			gistClient := gist.NewReadOnlyClient(cfg.GistID)
			var mappings []client.FileMapping
			for _, m := range cfg.Mappings {
				mappings = append(mappings, client.FileMapping{
					GistFile:    m.GistFile,
					LocalFile:   m.LocalPath,
					ExecCommand: m.Exec,
				})
			}

			if err := client.SyncOnce(gistClient, mappings); err != nil {
				logger.Fatal("Sync failed: %v", err)
			}
		},
	}

	rootCmd.AddCommand(cmdWatch, cmdSync)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("%v", err)
		os.Exit(1)
	}
}
