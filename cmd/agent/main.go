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

	if cfg.ClaudeAPIKey == "" {
		log.Error("Claude API key not configured. Set CLAUDE_API_KEY")
		os.Exit(1)
	}

	// Initialize clients
	redditClient := internal.NewRedditMCP(cfg.RedditMCPURL)
	log.Info("Reddit MCP client initialized (URL: %s)", cfg.RedditMCPURL)

	threadsClient := internal.NewThreadsClient(cfg.ThreadsAccessToken, cfg.ThreadsAPIKey)
	log.Info("Threads API client initialized")

	claudeGen := internal.NewClaudeGenerator(cfg.ClaudeAPIKey)
	log.Info("Claude content generator initialized")

	postGen := internal.NewAgent(claudeGen, cfg.PostContentTheme)
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
