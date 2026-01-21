package internal

import (
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"log/slog"
	"math/rand"
	"time"
)

// SocialMediaClient defines the interface for social media clients.
type SocialMediaClient interface {
	CreatePost(text string) (string, error)
	FollowUser(userHandle string) error
	LikePost(postID string) error
	GetRecentPosts(limit int) ([]string, error)
}

// Scheduler manages posting, following, and engagement activities.
type Scheduler struct {
	cron          *cron.Cron
	contentSource ContentSource
	socialMedia   SocialMediaClient
	agent         *Agent
	config        SchedulerConfig
	testMode      bool
}

// ContentSource defines the interface for content source clients.
type ContentSource interface {
	QueryWorkRantTweets(limit int) ([]Post, error)
}

// SchedulerConfig configures the scheduler's behavior.
type SchedulerConfig struct {
	PostingHours      []int
	FollowUsersPerDay int
	LikePostsPerDay   int
	MaxContentAgeDays int
	PostContentTheme  string
	TestMode          bool
}

// NewScheduler creates a new scheduler.
func NewScheduler(
	contentSource ContentSource,
	socialMedia SocialMediaClient,
	agent *Agent,
	config SchedulerConfig,
) *Scheduler {
	return &Scheduler{
		cron:          cron.New(),
		contentSource: contentSource,
		socialMedia:   socialMedia,
		agent:         agent,
		config:        config,
		testMode:      config.TestMode,
	}
}

// Start initializes and starts the scheduler.
func (s *Scheduler) Start(ctx context.Context) error {
	// In test mode, run routines once and return
	if s.testMode {
		slog.Info("test mode: running routines once")
		s.postRoutine(ctx)
		s.followRoutine(ctx)
		s.likeRoutine(ctx)
		slog.Info("test mode: all routines completed")
		return nil
	}

	for _, hour := range s.config.PostingHours {
		minute := rand.Intn(60)
		cronSpec := fmt.Sprintf("%d %d * * *", minute, hour)

		_, err := s.cron.AddFunc(cronSpec, func() {
			s.postRoutine(context.Background())
		})
		if err != nil {
			slog.Error("failed to schedule post at %d:%d - %v", hour, minute, err)
			return err
		}

		slog.Info("scheduled post creation at %02d:%02d", hour, minute)
	}

	followHour := 9 + rand.Intn(10)
	followMin := rand.Intn(60)
	followCron := fmt.Sprintf("%d %d * * *", followMin, followHour)

	_, err := s.cron.AddFunc(followCron, func() {
		s.followRoutine(context.Background())
	})
	if err != nil {
		slog.Error("failed to schedule follow routine - %v", err)
		return err
	}

	slog.Info("scheduled follow routine at %02d:%02d", followHour, followMin)

	likeHour := 10 + rand.Intn(9)
	likeMin := rand.Intn(60)
	likeCron := fmt.Sprintf("%d %d * * *", likeMin, likeHour)

	_, err = s.cron.AddFunc(likeCron, func() {
		s.likeRoutine(context.Background())
	})
	if err != nil {
		slog.Error("failed to schedule like routine - %v", err)
		return err
	}

	slog.Info("scheduled like routine at %02d:%02d", likeHour, likeMin)

	s.cron.Start()
	slog.Info("scheduler started")

	return nil
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	s.cron.Stop()
	slog.Info("scheduler stopped")
}

func (s *Scheduler) postRoutine(ctx context.Context) {
	slog.Info("starting post creation routine")

	posts, err := s.contentSource.QueryWorkRantTweets(10)
	if err != nil {
		slog.Error("failed to query Twitter/X: %v", err)
		return
	}

	if len(posts) == 0 {
		slog.Error("no work rant posts found on Twitter/X")
		return
	}

	cutoffTime := time.Now().AddDate(0, 0, -s.config.MaxContentAgeDays)
	var recentPosts []*Post
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
	slog.Debug("selected post: %s", selectedPost.Content)

	generatedPost, err := s.agent.Generate(ctx, selectedPost)
	if err != nil {
		slog.Error("failed to generate post: %v", err)
		return
	}

	postID, err := s.socialMedia.CreatePost(generatedPost.Content)
	if err != nil {
		slog.Error("failed to post to social media: %v", err)
		return
	}

	slog.Info("successfully posted to social media (ID: %s)", postID)
}

func (s *Scheduler) followRoutine(ctx context.Context) {
	slog.Info("starting follow routine")

	posts, err := s.contentSource.QueryWorkRantTweets(20)
	if err != nil {
		slog.Error("failed to fetch posts for follow routine: %v", err)
		return
	}

	authorMap := make(map[string]bool)
	var authors []string
	for _, post := range posts {
		if !authorMap[post.Author] && post.Author != "" {
			authors = append(authors, post.Author)
			authorMap[post.Author] = true
		}
	}

	if len(authors) == 0 {
		slog.Error("no valid authors found to follow")
		return
	}

	followCount := s.config.FollowUsersPerDay
	if followCount > len(authors) {
		followCount = len(authors)
	}

	for i := 0; i < followCount; i++ {
		idx := rand.Intn(len(authors))
		author := authors[idx]

		authors = append(authors[:idx], authors[idx+1:]...)

		err := s.socialMedia.FollowUser(author)
		if err != nil {
			slog.Error("failed to follow user %s: %v", author, err)
			continue
		}

		slog.Info("followed user: %s", author)

		time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)
	}
}

func (s *Scheduler) likeRoutine(ctx context.Context) {
	slog.Info("starting like routine")

	postIDs, err := s.socialMedia.GetRecentPosts(50)
	if err != nil {
		slog.Error("failed to fetch recent posts: %v", err)
		return
	}

	if len(postIDs) == 0 {
		slog.Error("no posts found on timeline")
		return
	}

	likeCount := s.config.LikePostsPerDay
	if likeCount > len(postIDs) {
		likeCount = len(postIDs)
	}

	for i := 0; i < likeCount; i++ {
		idx := rand.Intn(len(postIDs))
		postID := postIDs[idx]

		postIDs = append(postIDs[:idx], postIDs[idx+1:]...)

		err := s.socialMedia.LikePost(postID)
		if err != nil {
			slog.Error("failed to like post %s: %v", postID, err)
			continue
		}

		slog.Info("liked post: %s", postID)

		time.Sleep(time.Duration(1+rand.Intn(2)) * time.Second)
	}
}
