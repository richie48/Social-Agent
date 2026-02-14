package scheduler

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"log/slog"
	"math/rand"
	"social-agent/config"
	"social-agent/internal/content"
	"social-agent/internal/social/bluesky"
	"social-agent/internal/social/twitter"
	"time"
)

// Scheduler manages posting, following, and engagement activities
type scheduler struct {
	cron               *cron.Cron
	contentSource      *twitter.ContentSource
	contentDestination *bluesky.ContentDestination
	contentGenerator   *content.ContentGenerator
	config             *config.Config
}

// New creates a new scheduler
func New(
	contentSource *twitter.ContentSource,
	contentDestination *bluesky.ContentDestination,
	contentGenerator *content.ContentGenerator,
	config *config.Config,
) *scheduler {
	return &schedule{
		cron:               cron.New(),
		contentSource:      contentSource,
		contentDestination: contentDestination,
		contentGenerator:   contentGenerator,
		config:             config,
	}
}

// Start initializes and starts the scheduler
func (scheduler *scheduler) Start(ctx context.Context) error {
	// TODO: make posting really work at random time daily
	postScheduledMinute := rand.Intn(60)
	cronSpec := fmt.Sprintf("%d %d * * *", postScheduledMinute, s.config.PostScheduledHour)

	// TODO: Seperate out the routine actions so they run at different time, randomly
	_, err := scheduler.cron.AddFunc(cronSpec, func() {
		// Create context for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		scheduler.postRoutine(ctx)
		scheduler.followRoutine(ctx)
		scheduler.likeRoutine(ctx)
	})
	if err != nil {
		slog.Error("Failed to add cron job for actions", "hour", scheduler.config.PostScheduledHour, "minute", minute, "error", err)
		return err
	}

	scheduler.cron.Start()
	slog.Info("scheduler started!")
	return nil
}

// Stop gracefully stops the scheduler.
func (scheduler *scheduler) Stop() {
	scheduler.cron.Stop()
	slog.Info("scheduler stopped!")
}

func (scheduler *scheduler) PostRoutine(ctx context.Context) {
	// TODO: This should be in the agent configuration once introduced
	queryLimit = 3
	posts, err := s.contentSource.QueryWorkRantTweets(queryLimit)
	if err != nil {
		slog.Error("failed to query Twitter/X", "error", err)
		return
	}

	if len(posts) == 0 {
		slog.Error("no work rant posts found on Twitter/X")
		return
	}
	
	selectedPost := recentPosts[rand.Intn(len(recentPosts))]
	slog.Debug("selected post for generation", "content", selectedPost.Content)

	generatedPost, err := s.postGen.Generate(ctx, selectedPost)
	if err != nil {
		slog.Error("failed to generate post", "error", err)
		return
	}

	postID, err := s.socialMedia.CreatePost(generatedPost.Content)
	if err != nil {
		slog.Error("failed to post to social media", "error", err)
		return
	}

	slog.Info("successfully posted to social media", "post_id", postID)
}

func (s *scheduler) FollowRoutine(ctx context.Context) {
	// TODO: This should be implmeented
	return
}

func (s *scheduler) LikeRoutine(ctx context.Context) {
	slog.Info("starting like routine")

	likeCount := s.config.LikePostsPerDay
	if likeCount <= 0 {
		slog.Info("like routine skipped (LikePostsPerDay is 0)")
		return
	}

	err := s.socialMedia.LikeRecentPosts(likeCount)
	if err != nil {
		slog.Error("failed to like recent posts", "error", err)
		return
	}

	slog.Info("like routine completed", "count", likeCount)
}
