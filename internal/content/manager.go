package content

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"log/slog"
	"math/rand"
	"social-agent/config"
	"social-agent/internal/social/bluesky"
	"social-agent/internal/social/twitter"
)

// contentManager manages posting, following, and engagement activities
type contentManager struct {
	cron               *cron.Cron
	contentSource      twitter.ContentSource
	contentDestination bluesky.ContentDestination
	contentGenerator   ContentGenerator
	config             *config.Config
}

// NewManager creates a new contentManager
func NewManager(
	contentSource twitter.ContentSource,
	contentDestination bluesky.ContentDestination,
	contentGenerator ContentGenerator,
	config *config.Config,
) *contentManager {
	return &contentManager{
		cron:               cron.New(),
		contentSource:      contentSource,
		contentDestination: contentDestination,
		contentGenerator:   contentGenerator,
		config:             config,
	}
}

// Start initializes and starts the contentManager using the scheduled time and frequency
func (contentManager *contentManager) Start(ctx context.Context) error {
	// TODO: make posting really work at random time daily
	postScheduledMinute := rand.Intn(60)
	cronSpec := fmt.Sprintf("%d %d * * *", postScheduledMinute, contentManager.config.PostScheduledHour)

	// TODO: Seperate out the routine actions so they run at different time, randomly
	_, err := contentManager.cron.AddFunc(cronSpec, func() {
		// Create context for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		contentManager.PostRoutine(ctx)
		contentManager.FollowRoutine(ctx)
		contentManager.LikeRoutine(ctx)
	})
	if err != nil {
		slog.Error("Failed to add cron job for actions", "hour", contentManager.config.PostScheduledHour, "minute", postScheduledMinute, "error", err)
		return err
	}

	contentManager.cron.Start()
	slog.Info("content manager started!")
	return nil
}

// Stop gracefully stops the contentManager
func (contentManager *contentManager) Stop() {
	contentManager.cron.Stop()
	slog.Info("content manager stopped!")
}

// PostRoutine runs the routine to post content. Runs routine with context provided
func (contentManager *contentManager) PostRoutine(ctx context.Context) {
	// TODO: This should be in the agent configuration once introduced
	const queryLimit = 3
	posts, err := contentManager.contentSource.QueryWorkPosts(queryLimit)
	if err != nil {
		slog.Error("Failed to get posts from content source", "error", err)
		return
	}

	generatedPost, err := contentManager.contentGenerator.GeneratePost(ctx, posts)
	if err != nil {
		slog.Error("Failed to generate post using content source", "error", err)
		return
	}

	postID, err := contentManager.contentDestination.CreatePost(generatedPost)
	if err != nil {
		slog.Error("Failed to post to social media", "error", err)
		return
	}

	slog.Info("Successfully posted to social media", "post_id", postID)
}

// FollowRoutine runs the routine to follow users. Runs routine with context provided
func (contentManager *contentManager) FollowRoutine(ctx context.Context) {
	// TODO: Implement, idea at the moment is follow and like should be one routine.
	// Follow users for the posts i like
	return
}

// LikeRoutine runs the routine to like posts. Runs routine with context provided
func (contentManager *contentManager) LikeRoutine(ctx context.Context) {
	// TODO: Add early config validation so check like this are not needed
	likeCount := contentManager.config.LikePostsPerDay
	if likeCount <= 0 {
		slog.Warn("Like routine skipped, LikePostsPerDay need to be greaterthan 0")
		return
	}

	if err := contentManager.contentDestination.LikeRecentPosts(likeCount); err != nil {
		slog.Error("Failed to like recent posts", "error", err)
		return
	}

	slog.Info("Like routine completed", "likes", likeCount)
}
