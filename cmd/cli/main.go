package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/openBPL/internal/config"
	"github.com/openBPL/pkg/core"
	"github.com/spf13/cobra"
)

var (
	version   = "0.1.0-dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "openbpl",
	Short: "OpenBPL - Open Brand Protection Library",
	Long: `OpenBPL is an open-source framework for monitoring, detecting, 
and acting against brand infringements across the internet.`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildTime),
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the OpenBPL monitoring engine",
	Long: `Start the OpenBPL monitoring engine with the specified configuration.
This will begin monitoring certificate transparency logs and taking actions
on detected threats according to your rules.`,
	RunE: runMonitoring,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  "Commands for managing OpenBPL configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new OpenBPL configuration",
	Long:  "Create a sample configuration file to get started with OpenBPL",
	RunE:  initConfig,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("OpenBPL %s\n", rootCmd.Version)
	},
}

func runMonitoring(cmd *cobra.Command, args []string) error {
	configPath, _ := cmd.Flags().GetString("config")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	storageType, _ := cmd.Flags().GetString("storage")
	duration, _ := cmd.Flags().GetDuration("duration")

	log.Printf("🚀 Starting OpenBPL monitoring engine...")
	log.Printf("📋 Config: %s", configPath)

	// Load configuration
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with CLI flags
	if dryRun {
		cfg.DryRun = true
	}
	if storageType != "" {
		cfg.Storage.Type = storageType
	}

	log.Printf("💾 Storage: %s", cfg.Storage.Type)
	if cfg.DryRun {
		log.Printf("🔍 Running in DRY-RUN mode (no enforcement actions will be taken)")
	}

	// Create and start the monitoring engine
	engine, err := core.NewEngine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("🛑 Shutdown signal received...")
		cancel()
	}()

	// Set duration if specified
	if duration > 0 {
		var durCancel context.CancelFunc
		ctx, durCancel = context.WithTimeout(ctx, duration)
		defer durCancel()
		log.Printf("⏰ Will run for %s", duration)
	}

	// Start monitoring
	log.Printf("🎯 Starting monitoring engine...")
	return engine.Run(ctx)
}

func initConfig(cmd *cobra.Command, args []string) error {
	configPath := "openbpl.yaml"

	if err := config.CreateSampleConfig(configPath); err != nil {
		return err
	}

	fmt.Printf("✅ Configuration file created: %s\n", configPath)
	fmt.Println("📝 Edit the file to customize your monitoring settings.")
	fmt.Println("🚀 Run 'openbpl run' to start monitoring.")

	return nil
}

func init() {
	// Add flags to run command
	runCmd.Flags().StringP("config", "c", "openbpl.yaml", "Configuration file path")
	runCmd.Flags().BoolP("dry-run", "d", false, "Run in dry-run mode (no enforcement actions)")
	runCmd.Flags().StringP("storage", "s", "", "Storage backend (memory, sqlite, postgres)")
	runCmd.Flags().Duration("duration", 0, "Run for specific duration (0 = run forever)")

	// Build command tree
	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(runCmd, configCmd, versionCmd)
}
