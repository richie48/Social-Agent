package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/anthropics/sdk-go/client"
)

// ContentGenerator generates Threads posts from Reddit posts.
type ContentGenerator interface {
	GeneratePost(ctx context.Context, redditPost *RedditPost, theme string) (string, error)
}

// Agent generates posts from Reddit and posts to Threads.
type Agent struct {
	contentGen ContentGenerator
	theme      string
}

// GeneratedPost is a post ready to be posted to Threads.
type GeneratedPost struct {
	Content      string
	SourceURL    string
	SourceAuthor string
	CreatedAt    time.Time
}

// NewAgent creates a new post generation agent.
func NewAgent(contentGen ContentGenerator, theme string) *Agent {
	return &Agent{
		contentGen: contentGen,
		theme:      theme,
	}
}

// Generate creates a Threads post from a Reddit post.
func (a *Agent) Generate(ctx context.Context, redditPost *RedditPost) (*GeneratedPost, error) {
	if redditPost == nil {
		return nil, fmt.Errorf("reddit post is nil")
	}

	threadsContent, err := a.contentGen.GeneratePost(ctx, redditPost, a.theme)
	if err != nil {
		return nil, fmt.Errorf("failed to generate post content: %w", err)
	}

	return &GeneratedPost{
		Content:      threadsContent,
		SourceURL:    redditPost.URL,
		SourceAuthor: redditPost.Author,
		CreatedAt:    time.Now(),
	}, nil
}

// TruncateForThreads ensures the post fits within Threads character limit (500 chars).
func TruncateForThreads(content string, maxChars int) string {
	if len(content) <= maxChars {
		return content
	}

	truncated := content[:maxChars]
	for i := len(truncated) - 1; i >= 0; i-- {
		if truncated[i] == ' ' {
			return truncated[:i] + "..."
		}
	}

	return truncated + "..."
}

// ClaudeGenerator uses Claude to generate posts.
type ClaudeGenerator struct {
	client *client.Client
	model  string
}

// NewClaudeGenerator creates a new Claude-based generator.
func NewClaudeGenerator(apiKey string) *ClaudeGenerator {
	return &ClaudeGenerator{
		client: client.NewClient(apiKey),
		model:  "claude-3-5-sonnet-20241022",
	}
}

// GeneratePost creates a Threads post from a Reddit post using Claude.
func (cg *ClaudeGenerator) GeneratePost(ctx context.Context, redditPost *RedditPost, theme string) (string, error) {
	if redditPost == nil {
		return "", fmt.Errorf("reddit post is nil")
	}

	prompt := fmt.Sprintf(`You are a humorous social media content creator specializing in workplace frustration content. 
Your task is to create an engaging Threads post based on a Reddit story that embodies the theme: "%s"

Reddit Post Title: %s
Reddit Post Content: %s

Requirements:
1. Transform the Reddit story into a relatable, humorous Threads post about workplace frustrations
2. The post should be between 100-500 characters
3. Use conversational, natural language appropriate for Threads
4. Incorporate subtle humor and frustration about office dynamics, coworkers, or work situations
5. Make it engaging and likely to resonate with people frustrated at work
6. Do NOT include hashtags unless they naturally fit
7. Keep it authentic and relatable, not preachy
8. Optionally include a mild question or observation that invites engagement

Generate ONLY the post content, nothing else.`, theme, redditPost.Title, redditPost.Content)

	resp, err := cg.client.Messages.New(ctx, &client.MessageNewParams{
		Model:     client.String(cg.model),
		MaxTokens: client.Int(500),
		Messages: []client.MessageParam{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to call Claude API: %w", err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	textContent, ok := resp.Content[0].(*client.TextBlock)
	if !ok {
		return "", fmt.Errorf("unexpected response type from Claude")
	}

	generatedPost := textContent.Text

	if len(generatedPost) > 500 {
		generatedPost = TruncateForThreads(generatedPost, 500)
	}

	return generatedPost, nil
}
