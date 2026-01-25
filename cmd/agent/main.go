package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"social-agent/config"
	"social-agent/internal/agent"
	"social-agent/internal/scheduler"
	"social-agent/internal/social/bluesky"
	"social-agent/internal/social/twitter"
	"syscall"
)

func main() {
	var testMode = flag.Bool("test-mode", false, "Run in test mode (executes routines once and exits)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Social Media Agent\nUsage:\nsocial-agent [options]\nOptions:")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Load configuration and initialize logger
	loadedConfig := config.Load()
	// TODO: Take log level input as a service argument, use to set minimum level
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	slog.Debug("Configuration loaded. Mode: %s", func() string {
		if *testMode {
			return "test"
		}
		return "production"
	})

	// Initialize clients
	twitterClient := twitter.New(loadedConfig.TwitterBearerToken)
	blueskyClient := bluesky.New(loadedConfig.BlueskyAccessToken, loadedConfig.BlueskyDID)
	postGenerator, err := agent.New(loadedConfig.GeminiAPIKey, loadedConfig.PostContentTheme)
	if err != nil {
		slog.Error("Failed to initialize social agent: %v", err)
		os.Exit(1)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	actionScheduler := scheduler.New(
		twitterClient,
		blueskyClient,
		postGenerator,
		loadedConfig,
	)

	// In test mode, run routines once and exit
	if *testMode {
		slog.Info("test mode: running routines once")
		actionScheduler.RunPostRoutine(ctx)
		actionScheduler.RunFollowRoutine(ctx)
		actionScheduler.RunLikeRoutine(ctx)
		slog.Info("test mode: all routines completed successfully!")
		os.Exit(0)
	}

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if err := actionScheduler.Start(ctx); err != nil {
		slog.Error("Failed to start scheduler: %v", err)
		os.Exit(1)
	}

	// Wait for shutdown signal
	<-sigChan

	slog.Info("Shutdown signal received. Gracefully stopping...")
	actionScheduler.Stop()
}
