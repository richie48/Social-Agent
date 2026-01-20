package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"social-agent/config"
	"social-agent/internal"
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
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Social Media Agent starting...")
	mode := "production"
	if *testMode {
		mode = "test"
	}
	slog.Debug("Configuration loaded. Mode: %s", mode)

	// Initialize clients
	twitterXClient := internal.NewTwitterXClient(cfg.TwitterXBearerToken)
	blueskyClient := internal.NewBlueskyClient(cfg.BlueskyAccessToken, cfg.BlueskyDID)
	geminiGen, err := internal.NewGeminiGenerator(cfg.GeminiAPIKey)
	if err != nil {
		slog.Error("Failed to initialize Gemini generator: %v", err)
		os.Exit(1)
	}
	slog.Info("Gemini content generator initialized")
	postGen := internal.NewAgent(geminiGen, cfg.PostContentTheme)
	slog.Debug("Post generator initialized with theme: %s", cfg.PostContentTheme)

	// Create scheduler
	schedulerConfig := internal.SchedulerConfig{
		PostingHours:      []int{cfg.PostingScheduleHour},
		FollowUsersPerDay: cfg.FollowUsersPerDay,
		LikePostsPerDay:   cfg.LikePostsPerDay,
		MaxContentAgeDays: cfg.MaxContentAgeDays,
		PostContentTheme:  cfg.PostContentTheme,
		TestMode:          *testMode,
	}

	schedulerAgent := internal.NewScheduler(
		twitterXClient,
		blueskyClient,
		postGen,
		schedulerConfig,
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the scheduler
	if err := schedulerAgent.Start(ctx); err != nil {
		slog.Error("Failed to start scheduler: %v", err)
		os.Exit(1)
	}

	slog.Info("Agent is running. Press Ctrl+C to shutdown.")
	slog.Info("Scheduled tasks:")
	slog.Info("  - Posts at: %02d:xx daily", cfg.PostingScheduleHour)
	slog.Info("  - Follow %d users daily", cfg.FollowUsersPerDay)
	slog.Info("  - Like %d posts daily", cfg.LikePostsPerDay)
	slog.Info("  - Monitoring Twitter/X work rants")

	// Wait for shutdown signal
	<-sigChan

	slog.Info("Shutdown signal received. Gracefully stopping...")
	schedulerAgent.Stop()

	slog.Info("Social Media Agent stopped.")
}
