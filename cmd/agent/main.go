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
	cfg := config.Load()
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
	twitterClient := internal.NewTwitterClient(cfg.TwitterBearerToken)
	blueskyClient := internal.NewBlueskyClient(cfg.BlueskyAccessToken, cfg.BlueskyDID)
	geminiGen, err := internal.NewGeminiGenerator(cfg.GeminiAPIKey)
	if err != nil {
		slog.Error("Failed to initialize Gemini generator: %v", err)
		os.Exit(1)
	}
	postGen := internal.NewAgent(geminiGen, cfg.PostContentTheme)

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
		twitterClient,
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

	// Wait for shutdown signal
	<-sigChan

	slog.Info("Shutdown signal received. Gracefully stopping...")
	schedulerAgent.Stop()

	slog.Info("Social Media Agent stopped.")
}
