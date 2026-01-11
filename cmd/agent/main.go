package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"threads-influencer/config"
	"threads-influencer/internal"
)

func main() {
	var (
		dryRun = flag.Bool("dry-run", false, "Run in dry-run mode (no actual posts/actions)")
		debug  = flag.Bool("debug", false, "Enable debug logging")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Threads Influencer Agent

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

	log.Info("Threads Influencer Agent starting...")
	log.Debug("Configuration loaded. Mode: %s", map[bool]string{true: "dry-run", false: "production"}[*dryRun])

	// Validate required configuration
	if cfg.ThreadsAPIKey == "" || cfg.ThreadsAccessToken == "" {
		log.Error("Threads API credentials not configured. Set THREADS_API_KEY and THREADS_ACCESS_TOKEN")
		os.Exit(1)
	}

	if cfg.GeminiAPIKey == "" {
		log.Error("Gemini API key not configured. Set GEMINI_API_KEY")
		os.Exit(1)
	}

	// Validate Reddit credentials
	if cfg.RedditClientID == "" || cfg.RedditClientSecret == "" || cfg.RedditUsername == "" || cfg.RedditPassword == "" {
		log.Error("Reddit API credentials not configured. Set REDDIT_CLIENT_ID, REDDIT_CLIENT_SECRET, REDDIT_USERNAME, and REDDIT_PASSWORD")
		os.Exit(1)
	}

	// Initialize clients
	redditClient := internal.NewRedditClient(cfg.RedditClientID, cfg.RedditClientSecret, cfg.RedditUsername, cfg.RedditPassword, cfg.RedditUserAgent)
	log.Info("Reddit API client initialized")

	threadsClient := internal.NewThreadsClient(cfg.ThreadsAccessToken, cfg.ThreadsAPIKey)
	log.Info("Threads API client initialized")

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
		RedditSubreddits:  cfg.RedditSubreddits,
		MaxContentAgeDays: cfg.MaxContentAgeDays,
		PostContentTheme:  cfg.PostContentTheme,
	}

	schedulerAgent := internal.NewScheduler(
		redditClient,
		threadsClient,
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
	log.Info("  - Monitoring subreddits: %v", cfg.RedditSubreddits)

	// Wait for shutdown signal
	<-sigChan

	log.Info("Shutdown signal received. Gracefully stopping...")
	schedulerAgent.Stop()

	log.Info("Threads Influencer Agent stopped.")
}
