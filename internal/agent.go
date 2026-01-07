package internal

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"
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

// GeminiGenerator uses Google's Gemini to generate posts.
type GeminiGenerator struct {
	client *genai.Client
}

// NewGeminiGenerator creates a new Gemini-based generator.
func NewGeminiGenerator(apiKey string) (*GeminiGenerator, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiGenerator{
		client: client,
	}, nil
}

// GeneratePost creates a Threads post from a Reddit post using Gemini.
func (gg *GeminiGenerator) GeneratePost(ctx context.Context, redditPost *RedditPost, theme string) (string, error) {
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

	resp, err := gg.client.Models.GenerateContent(ctx, "gemini-2.5-flash", []*genai.Content{
		{
			Parts: []*genai.Part{
				{Text: prompt},
			},
		},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to call Gemini API: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	var generatedPost string
	if len(resp.Candidates[0].Content.Parts) > 0 {
		if resp.Candidates[0].Content.Parts[0].Text != "" {
			generatedPost = resp.Candidates[0].Content.Parts[0].Text
		}
	}

	if generatedPost == "" {
		return "", fmt.Errorf("no text content in Gemini response")
	}

	if len(generatedPost) > 500 {
		generatedPost = TruncateForThreads(generatedPost, 500)
	}

	return generatedPost, nil
}
