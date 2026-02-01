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

// Scheduler manages posting, following, and engagement activities.
type Scheduler struct {
	cron          *cron.Cron
	contentSource twitter.ContentSource
	socialMedia   bluesky.ContentDestination
	postGen       *content.Agent
	config        *config.Config
}

// New creates a new scheduler.
func New(
	contentSource twitter.ContentSource,
	socialMedia bluesky.ContentDestination,
	postGen *content.Agent,
	config *config.Config,
) *Scheduler {
	return &Scheduler{
		cron:          cron.New(),
		contentSource: contentSource,
		socialMedia:   socialMedia,
		postGen:       postGen,
		config:        config,
	}
}

// Start initializes and starts the scheduler.
func (s *Scheduler) Start(ctx context.Context) error {
	minute := rand.Intn(60)
	cronSpec := fmt.Sprintf("%d %d * * *", minute, s.config.PostingScheduleHour)

	_, err := s.cron.AddFunc(cronSpec, func() {
		s.postRoutine(context.Background())
	})
	if err != nil {
		slog.Error("failed to schedule post at", "hour", s.config.PostingScheduleHour, "minute", minute, "error", err)
		return err
	}

	slog.Info("scheduled post creation at", "hour", s.config.PostingScheduleHour, "minute", minute)

	followHour := 9 + rand.Intn(10)
	followMin := rand.Intn(60)
	followCron := fmt.Sprintf("%d %d * * *", followMin, followHour)

	_, err = s.cron.AddFunc(followCron, func() {
		s.followRoutine(context.Background())
	})
	if err != nil {
		slog.Error("failed to schedule follow routine", "error", err)
		return err
	}

	slog.Info("scheduled follow routine", "hour", followHour, "minute", followMin)

	likeHour := 10 + rand.Intn(9)
	likeMin := rand.Intn(60)
	likeCron := fmt.Sprintf("%d %d * * *", likeMin, likeHour)

	_, err = s.cron.AddFunc(likeCron, func() {
		s.likeRoutine(context.Background())
	})
	if err != nil {
		slog.Error("failed to schedule like routine", "error", err)
		return err
	}

	slog.Info("scheduled like routine", "hour", likeHour, "minute", likeMin)

	s.cron.Start()
	slog.Info("scheduler started")

	return nil
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	s.cron.Stop()
	slog.Info("scheduler stopped")
}

// RunPostRoutine exposes postRoutine for testing and direct invocation
func (s *Scheduler) RunPostRoutine(ctx context.Context) {
	s.postRoutine(ctx)
}

// RunFollowRoutine exposes followRoutine for testing and direct invocation
func (s *Scheduler) RunFollowRoutine(ctx context.Context) {
	s.followRoutine(ctx)
}

// RunLikeRoutine exposes likeRoutine for testing and direct invocation
func (s *Scheduler) RunLikeRoutine(ctx context.Context) {
	s.likeRoutine(ctx)
}

func (s *Scheduler) postRoutine(ctx context.Context) {
	slog.Info("starting post creation routine")

	posts, err := s.contentSource.QueryWorkRantTweets(3)
	if err != nil {
		slog.Error("failed to query Twitter/X", "error", err)
		return
	}

	if len(posts) == 0 {
		slog.Error("no work rant posts found on Twitter/X")
		return
	}

	cutoffTime := time.Now().AddDate(0, 0, -s.config.MaxContentAgeDays)
	var recentPosts []*twitter.Post
	for i, post := range posts {
		if post.CreatedAt.After(cutoffTime) {
			recentPosts = append(recentPosts, &posts[i])
		}
	}

	if len(recentPosts) == 0 {
		slog.Error("no recent work rant posts found on Twitter/X")
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

func (s *Scheduler) followRoutine(ctx context.Context) {
	// TODO: This should be implmeented
	return
}

func (s *Scheduler) likeRoutine(ctx context.Context) {
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
