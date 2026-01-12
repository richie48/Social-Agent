package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"social-agent/config"
	"social-agent/internal"
	"syscall"
)

func main() {
	var (
		dryRun = flag.Bool("dry-run", false, "Run in dry-run mode (no actual posts/actions)")
		test   = flag.Bool("test", false, "Run in test mode (executes routines once and exits)")
		debug  = flag.Bool("debug", false, "Enable debug logging")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Social Media Agent

Usage:
  social-agent [options]

Options:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logLevel := cfg.LogLevel
	if *debug {
		logLevel = "debug"
	}
	log := internal.NewLogger(logLevel)

	log.Info("Social Media Agent starting...")
	mode := "production"
	if *dryRun {
		mode = "dry-run"
	} else if *test {
		mode = "test"
	}
	log.Debug("Configuration loaded. Mode: %s", mode)

	// Skip credential validation in test mode
	if !*test {
		// Validate required configuration
		if cfg.TwitterXBearerToken == "" {
			log.Error("Twitter/X Bearer token not configured. Set TWITTER_X_BEARER_TOKEN")
			os.Exit(1)
		}

		if cfg.BlueskyAccessToken == "" || cfg.BlueskyDID == "" {
			log.Error("Bluesky credentials not configured. Set BLUESKY_ACCESS_TOKEN and BLUESKY_DID")
			os.Exit(1)
		}

		if cfg.GeminiAPIKey == "" {
			log.Error("Gemini API key not configured. Set GEMINI_API_KEY")
			os.Exit(1)
		}
	}

	// Initialize clients
	twitterXClient := internal.NewTwitterXClient(cfg.TwitterXBearerToken)
	log.Info("Twitter/X API client initialized")

	blueskyClient := internal.NewBlueskyClient(cfg.BlueskyAccessToken, cfg.BlueskyDID)
	log.Info("Bluesky API client initialized")

	geminiGen, err := internal.NewGeminiGenerator(cfg.GeminiAPIKey)
	if err != nil {
		log.Error("Failed to initialize Gemini generator: %v", err)
		os.Exit(1)
	}
	log.Info("Gemini content generator initialized")

	postGen := internal.NewAgent(geminiGen, cfg.PostContentTheme)
	log.Debug("Post generator initialized with theme: %s", cfg.PostContentTheme)

	// Create scheduler
	schedulerConfig := internal.SchedulerConfig{
		PostingHours:      []int{cfg.PostingScheduleHour1, cfg.PostingScheduleHour2},
		FollowUsersPerDay: cfg.FollowUsersPerDay,
		LikePostsPerDay:   cfg.LikePostsPerDay,
		MaxContentAgeDays: cfg.MaxContentAgeDays,
		PostContentTheme:  cfg.PostContentTheme,
		TestMode:          *test,
	}

	schedulerAgent := internal.NewScheduler(
		twitterXClient,
		blueskyClient,
		postGen,
		schedulerConfig,
		log,
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the scheduler
	if err := schedulerAgent.Start(ctx); err != nil {
		log.Error("Failed to start scheduler: %v", err)
		os.Exit(1)
	}

	log.Info("Agent is running. Press Ctrl+C to shutdown.")
	log.Info("Scheduled tasks:")
	log.Info("  - Posts at: %02d:xx and %02d:xx daily", cfg.PostingScheduleHour1, cfg.PostingScheduleHour2)
	log.Info("  - Follow %d users daily", cfg.FollowUsersPerDay)
	log.Info("  - Like %d posts daily", cfg.LikePostsPerDay)
	log.Info("  - Monitoring Twitter/X work rants")

	// Wait for shutdown signal
	<-sigChan

	log.Info("Shutdown signal received. Gracefully stopping...")
	schedulerAgent.Stop()

	log.Info("Social Media Agent stopped.")
}
